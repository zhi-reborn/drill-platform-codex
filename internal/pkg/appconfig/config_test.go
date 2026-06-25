package appconfig

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// writeYAML writes a minimal YAML config to a temp file and returns its path.
func writeYAML(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write yaml: %v", err)
	}
	return path
}

const baseYAML = `
server:
  port: 8080
  mode: debug
database:
  host: yaml-host
  port: 3306
  user: yaml-user
  password: yaml-pass
  database: yaml-db
redis:
  addr: yaml-redis:6379
  password: yaml-redis-pass
jwt:
  secret: yaml-secret
  expire: 24
public_base_url: http://yaml.example.com
cas:
  public_url: http://yaml.cas.example.com
  service_url: http://yaml.svc.example.com
worker:
  lease_ttl: 15s
  renew_interval: 5s
command_wait_timeout: 30s
login_log_file: logs/yaml.log
app_role: api
instance_id: yaml-node
`

// setEnv sets env vars for the test and registers a cleanup that restores the
// previous values.
func setEnv(t *testing.T, kv map[string]string) {
	t.Helper()
	for k, v := range kv {
		old, ok := os.LookupEnv(k)
		if err := os.Setenv(k, v); err != nil {
			t.Fatalf("setenv %s: %v", k, err)
		}
		k := k
		v := v
		t.Cleanup(func() {
			if ok {
				_ = os.Setenv(k, old)
			} else {
				_ = os.Unsetenv(k)
			}
			_ = v
		})
	}
}

// clearEnv unsets the given env vars for the test and restores them on cleanup.
func clearEnv(t *testing.T, keys ...string) {
	t.Helper()
	for _, k := range keys {
		old, ok := os.LookupEnv(k)
		if err := os.Unsetenv(k); err != nil {
			t.Fatalf("unsetenv %s: %v", k, err)
		}
		k := k
		t.Cleanup(func() {
			if ok {
				_ = os.Setenv(k, old)
			} else {
				_ = os.Unsetenv(k)
			}
		})
	}
}

// envKeys is the full set of env vars appconfig reads.
var envKeys = []string{
	"APP_ROLE", "INSTANCE_ID",
	"DATABASE_HOST", "DATABASE_PORT", "DATABASE_USER", "DATABASE_PASSWORD", "DATABASE_NAME",
	"REDIS_ADDR", "REDIS_PASSWORD",
	"JWT_SECRET",
	"PUBLIC_BASE_URL",
	"CAS_PUBLIC_URL", "CAS_SERVICE_URL",
	"WORKER_LEASE_TTL", "WORKER_RENEW_INTERVAL",
	"COMMAND_WAIT_TIMEOUT",
	"LOGIN_LOG_FILE",
}

func TestLoad_ReadsYAMLDefaults(t *testing.T) {
	clearEnv(t, envKeys...)
	path := writeYAML(t, baseYAML)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.AppRole != "api" {
		t.Errorf("AppRole = %q, want api", cfg.AppRole)
	}
	if cfg.Database.Host != "yaml-host" {
		t.Errorf("Database.Host = %q, want yaml-host", cfg.Database.Host)
	}
	if cfg.Redis.Addr != "yaml-redis:6379" {
		t.Errorf("Redis.Addr = %q, want yaml-redis:6379", cfg.Redis.Addr)
	}
	if cfg.JWT.Secret != "yaml-secret" {
		t.Errorf("JWT.Secret = %q, want yaml-secret", cfg.JWT.Secret)
	}
	if cfg.LoginLogFile != "logs/yaml.log" {
		t.Errorf("LoginLogFile = %q, want logs/yaml.log", cfg.LoginLogFile)
	}
}

func TestLoad_EnvOverridesYAML(t *testing.T) {
	clearEnv(t, envKeys...)
	setEnv(t, map[string]string{
		"APP_ROLE":            "worker",
		"INSTANCE_ID":         "env-node",
		"DATABASE_HOST":       "env-host",
		"DATABASE_PORT":       "3307",
		"DATABASE_USER":       "env-user",
		"DATABASE_PASSWORD":   "env-pass",
		"DATABASE_NAME":       "env-db",
		"REDIS_ADDR":          "env-redis:6380",
		"REDIS_PASSWORD":      "env-redis-pass",
		"JWT_SECRET":          "env-secret",
		"PUBLIC_BASE_URL":     "http://env.example.com",
		"CAS_PUBLIC_URL":      "http://env.cas.example.com",
		"CAS_SERVICE_URL":     "http://env.svc.example.com",
		"WORKER_LEASE_TTL":    "30s",
		"WORKER_RENEW_INTERVAL": "10s",
		"COMMAND_WAIT_TIMEOUT": "45s",
		"LOGIN_LOG_FILE":      "logs/env.log",
	})
	path := writeYAML(t, baseYAML)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.AppRole != "worker" {
		t.Errorf("AppRole = %q, want worker", cfg.AppRole)
	}
	if cfg.InstanceID != "env-node" {
		t.Errorf("InstanceID = %q, want env-node", cfg.InstanceID)
	}
	if cfg.Database.Host != "env-host" {
		t.Errorf("Database.Host = %q, want env-host", cfg.Database.Host)
	}
	if cfg.Database.Port != 3307 {
		t.Errorf("Database.Port = %d, want 3307", cfg.Database.Port)
	}
	if cfg.Database.User != "env-user" {
		t.Errorf("Database.User = %q, want env-user", cfg.Database.User)
	}
	if cfg.Database.Password != "env-pass" {
		t.Errorf("Database.Password = %q, want env-pass", cfg.Database.Password)
	}
	if cfg.Database.Name != "env-db" {
		t.Errorf("Database.Name = %q, want env-db", cfg.Database.Name)
	}
	if cfg.Redis.Addr != "env-redis:6380" {
		t.Errorf("Redis.Addr = %q, want env-redis:6380", cfg.Redis.Addr)
	}
	if cfg.Redis.Password != "env-redis-pass" {
		t.Errorf("Redis.Password = %q, want env-redis-pass", cfg.Redis.Password)
	}
	if cfg.JWT.Secret != "env-secret" {
		t.Errorf("JWT.Secret = %q, want env-secret", cfg.JWT.Secret)
	}
	if cfg.PublicBaseURL != "http://env.example.com" {
		t.Errorf("PublicBaseURL = %q, want http://env.example.com", cfg.PublicBaseURL)
	}
	if cfg.CAS.PublicURL != "http://env.cas.example.com" {
		t.Errorf("CAS.PublicURL = %q, want http://env.cas.example.com", cfg.CAS.PublicURL)
	}
	if cfg.CAS.ServiceURL != "http://env.svc.example.com" {
		t.Errorf("CAS.ServiceURL = %q, want http://env.svc.example.com", cfg.CAS.ServiceURL)
	}
	if cfg.Worker.LeaseTTL != 30*time.Second {
		t.Errorf("Worker.LeaseTTL = %v, want 30s", cfg.Worker.LeaseTTL)
	}
	if cfg.Worker.RenewInterval != 10*time.Second {
		t.Errorf("Worker.RenewInterval = %v, want 10s", cfg.Worker.RenewInterval)
	}
	if cfg.CommandWaitTimeout != 45*time.Second {
		t.Errorf("CommandWaitTimeout = %v, want 45s", cfg.CommandWaitTimeout)
	}
	if cfg.LoginLogFile != "logs/env.log" {
		t.Errorf("LoginLogFile = %q, want logs/env.log", cfg.LoginLogFile)
	}
}

func TestValidate_RejectsInvalidAppRole(t *testing.T) {
	cfg := DefaultConfig()
	cfg.AppRole = "bogus"
	cfg.JWT.Secret = "non-empty"
	cfg.Server.Mode = "debug"

	err := cfg.Validate()
	if err == nil {
		t.Fatalf("Validate: expected error for invalid AppRole, got nil")
	}
	if !strings.Contains(err.Error(), "app_role") && !strings.Contains(err.Error(), "APP_ROLE") {
		t.Fatalf("Validate error should mention app_role, got: %v", err)
	}
}

func TestValidate_RejectsEmptyJWTSecretInProduction(t *testing.T) {
	cfg := DefaultConfig()
	cfg.AppRole = "api"
	cfg.JWT.Secret = ""
	cfg.Server.Mode = "release"

	err := cfg.Validate()
	if err == nil {
		t.Fatalf("Validate: expected error for empty JWT secret in production, got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "jwt") {
		t.Fatalf("Validate error should mention jwt, got: %v", err)
	}
}

func TestValidate_AllowsEmptyJWTSecretInDebug(t *testing.T) {
	cfg := DefaultConfig()
	cfg.AppRole = "api"
	cfg.JWT.Secret = ""
	cfg.Server.Mode = "debug"

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate: expected nil for empty JWT secret in debug, got: %v", err)
	}
}

func TestValidate_RejectsWorkerRoleWithEmptyInstanceID(t *testing.T) {
	cfg := DefaultConfig()
	cfg.AppRole = "worker"
	cfg.InstanceID = ""
	cfg.JWT.Secret = "non-empty"
	cfg.Server.Mode = "release"

	err := cfg.Validate()
	if err == nil {
		t.Fatalf("Validate: expected error for worker role with empty InstanceID, got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "instance") {
		t.Fatalf("Validate error should mention instance, got: %v", err)
	}
}

func TestValidate_RejectsAllRoleWithEmptyInstanceID(t *testing.T) {
	cfg := DefaultConfig()
	cfg.AppRole = "all"
	cfg.InstanceID = ""
	cfg.JWT.Secret = "non-empty"
	cfg.Server.Mode = "release"

	err := cfg.Validate()
	if err == nil {
		t.Fatalf("Validate: expected error for all role with empty InstanceID, got nil")
	}
}

func TestValidate_AcceptsValidConfig(t *testing.T) {
	cfg := DefaultConfig()
	cfg.AppRole = "all"
	cfg.InstanceID = "node-a"
	cfg.JWT.Secret = "secret"
	cfg.Server.Mode = "release"

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate: expected nil for valid config, got: %v", err)
	}
}

func TestDefaultConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.AppRole != "all" {
		t.Errorf("default AppRole = %q, want all", cfg.AppRole)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("default Server.Port = %d, want 8080", cfg.Server.Port)
	}
	if cfg.Worker.LeaseTTL <= 0 {
		t.Errorf("default Worker.LeaseTTL = %v, want > 0", cfg.Worker.LeaseTTL)
	}
	if cfg.Worker.RenewInterval <= 0 {
		t.Errorf("default Worker.RenewInterval = %v, want > 0", cfg.Worker.RenewInterval)
	}
	if cfg.CommandWaitTimeout <= 0 {
		t.Errorf("default CommandWaitTimeout = %v, want > 0", cfg.CommandWaitTimeout)
	}
}

// Ensure context import is used even when no test currently needs it.
var _ = context.Background
