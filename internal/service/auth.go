package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/url"
	"time"

	"drill-platform/internal/domain/dto"
	"drill-platform/internal/domain/entity"
	"drill-platform/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	userRepo     *repository.UserRepo
	jwtSecret    string
	jwtExpire    time.Duration
	externalAuth ExternalAuthConfig
	casConfig    CASConfig
	casClient    *CASClient
	ldapConfig   LDAPConfig
	ldapClient   *LDAPClient
}

type ExternalAuthConfig struct {
	AutoCreateUser bool
	DefaultRole    string
	RoleMappings   map[string]string
}

type ExternalUser struct {
	Username   string
	RealName   string
	Email      string
	Phone      string
	Department string
	Groups     []string
}

func NewAuthService(userRepo *repository.UserRepo) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: "",
		jwtExpire: 24 * time.Hour,
	}
}

func (s *AuthService) SetJWTConfig(secret string, expireHours int) {
	s.jwtSecret = secret
	s.jwtExpire = time.Duration(expireHours) * time.Hour
}

func (s *AuthService) SetExternalAuthConfig(cfg ExternalAuthConfig) {
	if cfg.DefaultRole == "" {
		cfg.DefaultRole = "viewer"
	}
	s.externalAuth = cfg
}

func (s *AuthService) SetCASConfig(cfg CASConfig) {
	s.casConfig = cfg
	if cfg.Enabled && cfg.ServerURL != "" {
		s.casClient = NewCASClient(cfg.ServerURL)
	}
}

func (s *AuthService) SetLDAPConfig(cfg LDAPConfig) {
	s.ldapConfig = normalizeLDAPConfig(cfg)
	if cfg.Enabled {
		s.ldapClient = NewLDAPClient(cfg)
	}
}

func (s *AuthService) BuildCASLoginURL(serviceURL string) (string, error) {
	if !s.casConfig.Enabled {
		return "", errors.New("CAS 未启用")
	}
	return BuildCASLoginURL(firstNonEmpty(s.casConfig.PublicURL, s.casConfig.ServerURL), serviceURL)
}

func (s *AuthService) CASServiceURL(fallbackServiceURL, redirect string) string {
	serviceURL := firstNonEmpty(s.casConfig.ServiceURL, fallbackServiceURL)
	if redirect == "" {
		return serviceURL
	}
	u, err := url.Parse(serviceURL)
	if err != nil {
		return serviceURL
	}
	q := u.Query()
	q.Set("redirect", redirect)
	u.RawQuery = q.Encode()
	return u.String()
}

func (s *AuthService) LoginWithCASTicket(ticket, serviceURL string) (*dto.LoginResponse, error) {
	if !s.casConfig.Enabled || s.casClient == nil {
		return nil, errors.New("CAS 未启用")
	}
	username, err := s.casClient.ValidateTicket(ticket, serviceURL)
	if err != nil {
		return nil, err
	}

	ext := ExternalUser{Username: username, RealName: username}
	if s.ldapConfig.Enabled && s.ldapClient != nil {
		ldapUser, err := s.ldapClient.LookupUser(username)
		if err != nil {
			return nil, err
		}
		ext = ldapUser
	}
	return s.LoginWithExternalIdentity(ext)
}

func (s *AuthService) Login(req *dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := s.userRepo.FindByUsername(req.Username)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("用户名不存在")
	}
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("密码错误")
	}

	if user.Status != 1 {
		return nil, errors.New("账户已被禁用")
	}

	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		Token:      token,
		UserID:     user.ID,
		Username:   user.Username,
		RealName:   user.RealName,
		Role:       user.Role,
		Department: user.Department,
	}, nil
}

func (s *AuthService) LoginWithExternalIdentity(ext ExternalUser) (*dto.LoginResponse, error) {
	if ext.Username == "" {
		return nil, errors.New("外部认证用户名为空")
	}

	role := s.resolveExternalRole(ext.Groups)
	user, err := s.userRepo.FindByUsername(ext.Username)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if !s.externalAuth.AutoCreateUser {
			return nil, errors.New("用户不存在")
		}
		passwordHash, hashErr := generatedPasswordHash()
		if hashErr != nil {
			return nil, hashErr
		}
		user = &entity.User{
			Username:     ext.Username,
			RealName:     firstNonEmpty(ext.RealName, ext.Username),
			PasswordHash: passwordHash,
			Role:         role,
			Email:        ext.Email,
			Phone:        ext.Phone,
			Department:   ext.Department,
			Status:       1,
		}
		if err := s.userRepo.Create(user); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	} else {
		user.RealName = firstNonEmpty(ext.RealName, user.RealName, user.Username)
		user.Email = ext.Email
		user.Phone = ext.Phone
		user.Department = ext.Department
		user.Role = role
		if err := s.userRepo.Update(user); err != nil {
			return nil, err
		}
	}

	if user.Status != 1 {
		return nil, errors.New("账户已被禁用")
	}

	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}
	return &dto.LoginResponse{
		Token:      token,
		UserID:     user.ID,
		Username:   user.Username,
		RealName:   user.RealName,
		Role:       user.Role,
		Department: user.Department,
	}, nil
}

func (s *AuthService) resolveExternalRole(groups []string) string {
	for role, mappedGroup := range s.externalAuth.RoleMappings {
		for _, group := range groups {
			if group == mappedGroup {
				return role
			}
		}
	}
	return firstNonEmpty(s.externalAuth.DefaultRole, "viewer")
}

func generatedPasswordHash() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(hex.EncodeToString(buf)), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func (s *AuthService) generateToken(user *entity.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(s.jwtExpire).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

func (s *AuthService) ListUsers() ([]entity.User, error) {
	return s.userRepo.ListAll()
}

func (s *AuthService) ListUsersPaginated(page, pageSize int, role string) ([]entity.User, int64, error) {
	return s.userRepo.List(page, pageSize, role)
}

func (s *AuthService) GetDepartments() ([]string, error) {
	return s.userRepo.GetDistinctDepartments()
}

func (s *AuthService) GetUserByID(id uint64) (*entity.User, error) {
	user, err := s.userRepo.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	return user, err
}

func (s *AuthService) GetUsersByIDs(ids []uint64) (map[uint64]*entity.User, error) {
	return s.userRepo.FindByIDs(ids)
}

func (s *AuthService) CreateUser(req *dto.CreateUserRequest) (*entity.User, error) {
	existing, _ := s.userRepo.FindByUsername(req.Username)
	if existing != nil {
		return nil, ErrUserAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &entity.User{
		Username:     req.Username,
		RealName:     req.RealName,
		PasswordHash: string(hashedPassword),
		Role:         req.Role,
		Email:        req.Email,
		Phone:        req.Phone,
		Department:   req.Department,
		Status:       1,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *AuthService) UpdateUser(id uint64, req *dto.UpdateUserRequest) (*entity.User, error) {
	user, err := s.userRepo.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	if req.RealName != "" {
		user.RealName = req.RealName
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Role != "" {
		user.Role = req.Role
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Department != "" {
		user.Department = req.Department
	}
	if req.Status != nil {
		user.Status = *req.Status
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}
	return s.userRepo.FindByID(id)
}

func (s *AuthService) ResetPassword(id uint64, newPassword string) error {
	user, err := s.userRepo.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrUserNotFound
	}
	if err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hashedPassword)
	return s.userRepo.Update(user)
}

func (s *AuthService) DeleteUser(id uint64) error {
	_, err := s.userRepo.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrUserNotFound
	}
	if err != nil {
		return err
	}
	return s.userRepo.Delete(id)
}
