// Package appconfig centralizes drill-platform runtime configuration.
//
// Configuration is loaded from a YAML file and then overridden by environment
// variables. The environment variables are the authoritative source in
// container deployments where YAML is a fallback for defaults.
package appconfig

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Role constants describe the valid APP_ROLE values.
const (
	RoleAPI    = "api"
	RoleWorker = "worker"
	RoleAll    = "all"
)

// ServerConfig holds HTTP server tuning knobs.
type ServerConfig struct {
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"` // debug, release, test
}

// DatabaseConfig mirrors repository.Config.
type DatabaseConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	User         string `yaml:"user"`
	Password     string `yaml:"password"`
	Name         string `yaml:"name"`
	MaxIdleConns int    `yaml:"maxIdleConns"`
	MaxOpenConns int    `yaml:"maxOpenConns"`
}

// RedisConfig describes the Redis connection. Addr is the canonical "host:port"
// form used by environment variables; Host/Port are kept for YAML backward
// compatibility.
type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	PoolSize int    `yaml:"poolSize"`
}

// JWTConfig holds JWT signing parameters.
type JWTConfig struct {
	Secret string `yaml:"secret"`
	Expire int    `yaml:"expire"` // hours
}

// AuthConfig mirrors service.ExternalAuthConfig.
type AuthConfig struct {
	Mode           string            `yaml:"mode"`
	AutoCreateUser bool              `yaml:"autoCreateUser"`
	DefaultRole    string            `yaml:"defaultRole"`
	RoleMappings   map[string]string `yaml:"roleMappings"`
}

// CASConfig mirrors service.CASConfig.
type CASConfig struct {
	Enabled    bool   `yaml:"enabled"`
	ServerURL  string `yaml:"serverURL"`
	PublicURL  string `yaml:"publicURL"`
	ServiceURL string `yaml:"serviceURL"`
}

// LDAPConfig mirrors service.LDAPConfig.
type LDAPConfig struct {
	Enabled             bool   `yaml:"enabled"`
	URL                 string `yaml:"url"`
	BindDN              string `yaml:"bindDN"`
	BindPassword        string `yaml:"bindPassword"`
	BaseDN              string `yaml:"baseDN"`
	UserFilter          string `yaml:"userFilter"`
	UsernameAttribute   string `yaml:"usernameAttribute"`
	RealNameAttribute   string `yaml:"realNameAttribute"`
	EmailAttribute      string `yaml:"emailAttribute"`
	PhoneAttribute      string `yaml:"phoneAttribute"`
	DepartmentAttribute string `yaml:"departmentAttribute"`
	GroupBaseDN         string `yaml:"groupBaseDN"`
	GroupFilter         string `yaml:"groupFilter"`
	GroupNameAttribute  string `yaml:"groupNameAttribute"`
	TimeoutSeconds      int    `yaml:"timeoutSeconds"`
}

// WorkerConfig holds the leader-election timing parameters consumed by
// worker.Config.
type WorkerConfig struct {
	LeaseTTL      time.Duration `yaml:"lease_ttl"`
	RenewInterval time.Duration `yaml:"renew_interval"`
}

// Config is the single source of truth for drill-platform runtime settings.
type Config struct {
	AppRole            string         `yaml:"app_role"`
	InstanceID         string         `yaml:"instance_id"`
	Server             ServerConfig   `yaml:"server"`
	Database           DatabaseConfig `yaml:"database"`
	Redis              RedisConfig    `yaml:"redis"`
	JWT                JWTConfig      `yaml:"jwt"`
	Auth               AuthConfig     `yaml:"auth"`
	CAS                CASConfig      `yaml:"cas"`
	LDAP               LDAPConfig     `yaml:"ldap"`
	PublicBaseURL      string         `yaml:"public_base_url"`
	Worker             WorkerConfig   `yaml:"worker"`
	CommandWaitTimeout time.Duration  `yaml:"command_wait_timeout"`
	LoginLogFile       string         `yaml:"login_log_file"`
}

// DefaultConfig returns a Config populated with sensible defaults. Callers are
// expected to override fields from YAML and environment variables via Load.
func DefaultConfig() *Config {
	return &Config{
		AppRole:    RoleAll,
		InstanceID: "node-1",
		Server: ServerConfig{
			Port: 8080,
			Mode: "debug",
		},
		Database: DatabaseConfig{
			Host:         "127.0.0.1",
			Port:         3306,
			MaxIdleConns: 10,
			MaxOpenConns: 100,
		},
		Redis: RedisConfig{
			Addr:     "127.0.0.1:6379",
			DB:       0,
			PoolSize: 10,
		},
		JWT: JWTConfig{
			Expire: 24,
		},
		Worker: WorkerConfig{
			LeaseTTL:      15 * time.Second,
			RenewInterval: 5 * time.Second,
		},
		CommandWaitTimeout: 30 * time.Second,
		LoginLogFile:       "",
	}
}

// Load reads the YAML file at yamlPath and then applies environment variable
// overrides. Missing files are tolerated when env vars supply the required
// values; this keeps container deployments that mount no config file working.
func Load(yamlPath string) (*Config, error) {
	cfg := DefaultConfig()

	if yamlPath != "" {
		data, err := os.ReadFile(yamlPath)
		if err == nil {
			if err := yaml.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("parse yaml %s: %w", yamlPath, err)
			}
		} else if !os.IsNotExist(err) {
			return nil, fmt.Errorf("read config %s: %w", yamlPath, err)
		}
	}

	applyEnvOverrides(cfg)

	return cfg, nil
}

// applyEnvOverrides overlays environment variables on top of the YAML-loaded
// config. Empty env vars are ignored so they cannot blank out YAML values.
func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("APP_ROLE"); v != "" {
		cfg.AppRole = v
	}
	if v := os.Getenv("INSTANCE_ID"); v != "" {
		cfg.InstanceID = v
	}
	if v := os.Getenv("DATABASE_HOST"); v != "" {
		cfg.Database.Host = v
	}
	if v := os.Getenv("DATABASE_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.Database.Port = port
		}
	}
	if v := os.Getenv("DATABASE_USER"); v != "" {
		cfg.Database.User = v
	}
	if v := os.Getenv("DATABASE_PASSWORD"); v != "" {
		cfg.Database.Password = v
	}
	if v := os.Getenv("DATABASE_NAME"); v != "" {
		cfg.Database.Name = v
	}
	if v := os.Getenv("REDIS_ADDR"); v != "" {
		cfg.Redis.Addr = v
	}
	if v := os.Getenv("REDIS_PASSWORD"); v != "" {
		cfg.Redis.Password = v
	}
	if v := os.Getenv("JWT_SECRET"); v != "" {
		cfg.JWT.Secret = v
	}
	if v := os.Getenv("PUBLIC_BASE_URL"); v != "" {
		cfg.PublicBaseURL = v
	}
	if v := os.Getenv("CAS_PUBLIC_URL"); v != "" {
		cfg.CAS.PublicURL = v
	}
	if v := os.Getenv("CAS_SERVICE_URL"); v != "" {
		cfg.CAS.ServiceURL = v
	}
	if v := os.Getenv("WORKER_LEASE_TTL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Worker.LeaseTTL = d
		}
	}
	if v := os.Getenv("WORKER_RENEW_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Worker.RenewInterval = d
		}
	}
	if v := os.Getenv("COMMAND_WAIT_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.CommandWaitTimeout = d
		}
	}
	if v := os.Getenv("LOGIN_LOG_FILE"); v != "" {
		cfg.LoginLogFile = v
	}
}

// Validate enforces invariants required for safe operation:
//   - APP_ROLE must be one of api, worker, all;
//   - production deployments (Server.Mode != "debug") must set a JWT secret;
//   - worker/all roles must set a non-empty InstanceID for lease ownership.
func (c *Config) Validate() error {
	role := strings.ToLower(c.AppRole)
	switch role {
	case RoleAPI, RoleWorker, RoleAll:
	default:
		return fmt.Errorf("app_role %q is invalid; must be one of api, worker, all", c.AppRole)
	}

	if strings.ToLower(c.Server.Mode) != "debug" && c.JWT.Secret == "" {
		return fmt.Errorf("jwt secret must not be empty in non-debug mode (set JWT_SECRET)")
	}

	if (role == RoleWorker || role == RoleAll) && c.InstanceID == "" {
		return fmt.Errorf("instance_id must not be empty for app_role %q (set INSTANCE_ID)", role)
	}

	return nil
}

// IsWorker reports whether the configured role runs the Worker loop.
func (c *Config) IsWorker() bool {
	role := strings.ToLower(c.AppRole)
	return role == RoleWorker || role == RoleAll
}

// IsAPI reports whether the configured role serves the public HTTP API.
func (c *Config) IsAPI() bool {
	role := strings.ToLower(c.AppRole)
	return role == RoleAPI || role == RoleAll
}

// RedisAddress resolves the effective Redis address, preferring the combined
// REDIS_ADDR form and falling back to host:port from YAML.
func (c *Config) RedisAddress() string {
	if c.Redis.Addr != "" {
		return c.Redis.Addr
	}
	if c.Redis.Host != "" {
		if c.Redis.Port > 0 {
			return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
		}
		return c.Redis.Host
	}
	return ""
}
