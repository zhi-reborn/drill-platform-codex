package service

import (
	"testing"

	"drill-platform/internal/domain/entity"
	"drill-platform/internal/repository"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAuthExternalTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(&entity.User{}); err != nil {
		t.Fatalf("migrate user: %v", err)
	}
	return db
}

func TestLoginWithExternalIdentityAutoCreatesAndSyncsLDAPUser(t *testing.T) {
	db := setupAuthExternalTestDB(t)
	origDB := repository.DB
	repository.DB = db
	defer func() { repository.DB = origDB }()

	svc := NewAuthService(repository.NewUserRepo())
	svc.SetJWTConfig("test-secret", 24)
	svc.SetExternalAuthConfig(ExternalAuthConfig{
		AutoCreateUser: true,
		DefaultRole:    "viewer",
		RoleMappings: map[string]string{
			"director": "drill-director",
		},
	})

	res, err := svc.LoginWithExternalIdentity(ExternalUser{
		Username:   "zhangsan",
		RealName:   "张三",
		Email:      "zhangsan@example.com",
		Phone:      "13800000000",
		Department: "技术部",
		Groups:     []string{"drill-director"},
	})
	if err != nil {
		t.Fatalf("login with external identity: %v", err)
	}
	if res.Token == "" {
		t.Fatal("expected token")
	}
	if res.Username != "zhangsan" || res.RealName != "张三" || res.Department != "技术部" {
		t.Fatalf("unexpected response user: %+v", res)
	}
	if res.Role != "director" {
		t.Fatalf("expected director role, got %s", res.Role)
	}

	created, err := repository.NewUserRepo().FindByUsername("zhangsan")
	if err != nil {
		t.Fatalf("find created user: %v", err)
	}
	if created.PasswordHash == "" {
		t.Fatal("expected generated password hash")
	}
	if created.Status != 1 {
		t.Fatalf("expected active user, got status %d", created.Status)
	}

	_, err = svc.LoginWithExternalIdentity(ExternalUser{
		Username:   "zhangsan",
		RealName:   "张三丰",
		Email:      "zhangsan.new@example.com",
		Phone:      "13900000000",
		Department: "运维部",
		Groups:     []string{"unknown-group"},
	})
	if err != nil {
		t.Fatalf("second external login: %v", err)
	}

	updated, err := repository.NewUserRepo().FindByUsername("zhangsan")
	if err != nil {
		t.Fatalf("find updated user: %v", err)
	}
	if updated.RealName != "张三丰" || updated.Email != "zhangsan.new@example.com" || updated.Department != "运维部" {
		t.Fatalf("expected LDAP profile sync, got %+v", updated)
	}
	if updated.Role != "viewer" {
		t.Fatalf("expected fallback viewer role after unmatched groups, got %s", updated.Role)
	}
}
