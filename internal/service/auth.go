package service

import (
	"errors"
	"time"

	"drill-platform/internal/domain/dto"
	"drill-platform/internal/domain/entity"
	"drill-platform/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	userRepo *repository.UserRepo
	jwtSecret string
	jwtExpire time.Duration
}

func NewAuthService(userRepo *repository.UserRepo) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: "your-secret-key-change-in-production",
		jwtExpire: 24 * time.Hour,
	}
}

func (s *AuthService) SetJWTConfig(secret string, expireHours int) {
	s.jwtSecret = secret
	s.jwtExpire = time.Duration(expireHours) * time.Hour
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
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("密码错误")
	}

	if user.Status != 1 {
		return nil, errors.New("账户已被禁用")
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
