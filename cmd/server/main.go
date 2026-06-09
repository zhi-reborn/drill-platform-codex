package main

import (
	"fmt"
	"log"
	"os"

	"drill-platform/internal/api/router"
	"drill-platform/internal/infrastructure/redis"
	"drill-platform/internal/infrastructure/websocket"
	"drill-platform/internal/pkg/loginlog"
	"drill-platform/internal/repository"
	"drill-platform/internal/service"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig       `yaml:"server"`
	Database DatabaseConfig     `yaml:"database"`
	Redis    RedisConfig        `yaml:"redis"`
	JWT      JWTConfig          `yaml:"jwt"`
	Auth     AuthConfig         `yaml:"auth"`
	CAS      service.CASConfig  `yaml:"cas"`
	LDAP     service.LDAPConfig `yaml:"ldap"`
	Log      LogConfig          `yaml:"log"`
}

type AuthConfig struct {
	Mode           string            `yaml:"mode"`
	AutoCreateUser bool              `yaml:"autoCreateUser"`
	DefaultRole    string            `yaml:"defaultRole"`
	RoleMappings   map[string]string `yaml:"roleMappings"`
}

type ServerConfig struct {
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"`
}

type DatabaseConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	User         string `yaml:"user"`
	Password     string `yaml:"password"`
	Database     string `yaml:"database"`
	MaxIdleConns int    `yaml:"maxIdleConns"`
	MaxOpenConns int    `yaml:"maxOpenConns"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	PoolSize int    `yaml:"poolSize"`
}

type JWTConfig struct {
	Secret string `yaml:"secret"`
	Expire int    `yaml:"expire"`
}

type LogConfig struct {
	LoginLogFile string `yaml:"loginLogFile"`
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	return &cfg, yaml.Unmarshal(data, &cfg)
}

func main() {
	configPaths := []string{
		"configs/config.yaml",
		"../../configs/config.yaml",
		"/data/opencode/drill-platform/configs/config.yaml",
	}

	var cfg *Config
	var err error

	for _, path := range configPaths {
		cfg, err = loadConfig(path)
		if err == nil {
			log.Printf("配置文件加载成功：%s", path)
			break
		}
	}

	if err != nil {
		log.Fatalf("加载配置失败：%v", err)
	}

	if err := repository.InitDB(&repository.Config{
		Host:         cfg.Database.Host,
		Port:         cfg.Database.Port,
		User:         cfg.Database.User,
		Password:     cfg.Database.Password,
		Database:     cfg.Database.Database,
		MaxIdleConns: cfg.Database.MaxIdleConns,
		MaxOpenConns: cfg.Database.MaxOpenConns,
	}); err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer repository.Close()
	log.Println("数据库连接成功")

	redisClient, err := redis.NewClient(&redis.Config{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})
	if err != nil {
		log.Printf("Redis连接失败 (可忽略): %v", err)
	} else {
		log.Println("Redis连接成功")
		defer redisClient.Close()
	}

	wsManager := websocket.NewManager()
	go wsManager.Run()

	services := service.NewServices(wsManager, redisClient)
	services.AuthService.SetJWTConfig(cfg.JWT.Secret, cfg.JWT.Expire)
	services.AuthService.SetExternalAuthConfig(service.ExternalAuthConfig{
		AutoCreateUser: cfg.Auth.AutoCreateUser,
		DefaultRole:    cfg.Auth.DefaultRole,
		RoleMappings:   cfg.Auth.RoleMappings,
	})
	services.AuthService.SetCASConfig(cfg.CAS)
	services.AuthService.SetLDAPConfig(cfg.LDAP)

	loginLogger, err := loginlog.New(cfg.Log.LoginLogFile)
	if err != nil {
		log.Printf("登录日志初始化失败（将仅输出到标准输出）: %v", err)
	} else {
		defer loginLogger.Close()
		log.Printf("登录日志文件：%s", cfg.Log.LoginLogFile)
	}

	r := router.SetupRouter(services, wsManager, cfg.JWT.Secret, loginLogger)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("服务启动在 %s (mode=%s)", addr, cfg.Server.Mode)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
