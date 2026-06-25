package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"drill-platform/internal/api/handler/health"
	"drill-platform/internal/api/router"
	"drill-platform/internal/infrastructure/events"
	"drill-platform/internal/infrastructure/redis"
	"drill-platform/internal/infrastructure/websocket"
	"drill-platform/internal/pkg/appconfig"
	"drill-platform/internal/pkg/loginlog"
	"drill-platform/internal/repository"
	"drill-platform/internal/service"
	"drill-platform/internal/worker"

	"gorm.io/gorm"
)

// dbHealthChecker adapts *gorm.DB to the health.HealthChecker interface.
type dbHealthChecker struct {
	db *gorm.DB
}

func (d *dbHealthChecker) Ping(ctx context.Context) error {
	if d.db == nil {
		return errors.New("database not initialized")
	}
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

// redisHostPort resolves the Redis host and port from the combined Addr field
// when present, falling back to the legacy Host/Port YAML fields.
func redisHostPort(cfg *appconfig.Config) (string, int) {
	addr := cfg.RedisAddress()
	if addr == "" {
		return cfg.Redis.Host, cfg.Redis.Port
	}
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return addr, cfg.Redis.Port
	}
	p, err := strconv.Atoi(port)
	if err != nil {
		return host, cfg.Redis.Port
	}
	return host, p
}

// toServiceCAS converts appconfig.CASConfig to service.CASConfig.
func toServiceCAS(c appconfig.CASConfig) service.CASConfig {
	return service.CASConfig{
		Enabled:    c.Enabled,
		ServerURL:  c.ServerURL,
		PublicURL:  c.PublicURL,
		ServiceURL: c.ServiceURL,
	}
}

// toServiceLDAP converts appconfig.LDAPConfig to service.LDAPConfig.
func toServiceLDAP(l appconfig.LDAPConfig) service.LDAPConfig {
	return service.LDAPConfig{
		Enabled:             l.Enabled,
		URL:                 l.URL,
		BindDN:              l.BindDN,
		BindPassword:        l.BindPassword,
		BaseDN:              l.BaseDN,
		UserFilter:          l.UserFilter,
		UsernameAttribute:   l.UsernameAttribute,
		RealNameAttribute:   l.RealNameAttribute,
		EmailAttribute:      l.EmailAttribute,
		PhoneAttribute:      l.PhoneAttribute,
		DepartmentAttribute: l.DepartmentAttribute,
		GroupBaseDN:         l.GroupBaseDN,
		GroupFilter:         l.GroupFilter,
		GroupNameAttribute:  l.GroupNameAttribute,
		TimeoutSeconds:      l.TimeoutSeconds,
	}
}

func main() {
	// 1. Load and validate configuration.
	configPaths := []string{
		"configs/config.yaml",
		"/app/configs/config.yaml",
		"../../configs/config.yaml",
		"/data/opencode/drill-platform/configs/config.yaml",
	}

	var cfg *appconfig.Config
	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			c, err := appconfig.Load(path)
			if err != nil {
				log.Fatalf("加载配置失败：%v", err)
			}
			cfg = c
			log.Printf("配置文件加载成功：%s", path)
			break
		}
	}
	if cfg == nil {
		log.Println("未找到配置文件，使用默认配置")
		cfg, _ = appconfig.Load("")
	}
	if err := cfg.Validate(); err != nil {
		log.Fatalf("配置校验失败：%v", err)
	}
	log.Printf("应用角色: %s, 实例ID: %s", cfg.AppRole, cfg.InstanceID)

	// 2. Initialize database.
	if err := repository.InitDB(&repository.Config{
		Host:         cfg.Database.Host,
		Port:         cfg.Database.Port,
		User:         cfg.Database.User,
		Password:     cfg.Database.Password,
		Database:     cfg.Database.Name,
		MaxIdleConns: cfg.Database.MaxIdleConns,
		MaxOpenConns: cfg.Database.MaxOpenConns,
	}); err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	log.Println("数据库连接成功")

	// 3. Initialize Redis (optional - API can run without it).
	var redisClient *redis.Client
	if addr := cfg.RedisAddress(); addr != "" {
		host, port := redisHostPort(cfg)
		rc, err := redis.NewClient(&redis.Config{
			Host:     host,
			Port:     port,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
			PoolSize: cfg.Redis.PoolSize,
		})
		if err != nil {
			log.Printf("Redis连接失败 (可忽略): %v", err)
		} else {
			redisClient = rc
			log.Println("Redis连接成功")
		}
	}

	// 4. WebSocket manager.
	wsManager := websocket.NewManager()
	go wsManager.Run()

	// 5. Services.
	services := service.NewServices(wsManager, redisClient)
	services.AuthService.SetJWTConfig(cfg.JWT.Secret, cfg.JWT.Expire)
	services.AuthService.SetExternalAuthConfig(service.ExternalAuthConfig{
		AutoCreateUser: cfg.Auth.AutoCreateUser,
		DefaultRole:    cfg.Auth.DefaultRole,
		RoleMappings:   cfg.Auth.RoleMappings,
	})
	services.AuthService.SetCASConfig(toServiceCAS(cfg.CAS))
	services.AuthService.SetLDAPConfig(toServiceLDAP(cfg.LDAP))

	// 6. Login logger (empty file => stdout only).
	loginLogger, err := loginlog.New(cfg.LoginLogFile)
	if err != nil {
		log.Printf("登录日志初始化失败（回退到标准输出）: %v", err)
		loginLogger, _ = loginlog.New("")
	} else if cfg.LoginLogFile != "" {
		log.Printf("登录日志文件：%s", cfg.LoginLogFile)
	} else {
		log.Println("登录日志输出到标准输出")
	}

	// 7. Events bus (publisher + subscriber). The same RedisBus instance acts
	// as both; the subscriber half is started in a goroutine for API roles.
	var eventBus *events.RedisBus
	var subscriber events.Subscriber
	if redisClient != nil {
		eventBus = events.NewRedisBus(redisClient)
		if cfg.IsAPI() {
			subscriber = eventBus
		}
	}

	// 8. Start events subscriber (API/all roles). The subscriber reports
	// healthy once it has an active Redis subscription.
	subCtx, subCancel := context.WithCancel(context.Background())
	defer subCancel()
	if subscriber != nil {
		go func() {
			if err := eventBus.Subscribe(subCtx, func(event events.Event) {
				wsManager.DeliverEvent(event)
			}); err != nil && !errors.Is(err, context.Canceled) {
				log.Printf("事件订阅退出: %v", err)
			}
		}()
	}

	// 9. Worker (worker/all roles). Requires Redis for leader election.
	var flowWorker *worker.Worker
	if cfg.IsWorker() {
		if redisClient == nil {
			log.Printf("Worker 角色需要 Redis 进行领导选举，但 Redis 不可用")
		} else {
			flowCommandRepo := repository.NewFlowCommandRepo()
			drillRepo := repository.NewDrillRepo()
			stepRepo := repository.NewStepRepo()

			lease := redis.NewLease(redisClient, "drill:worker:leader", cfg.InstanceID, cfg.Worker.LeaseTTL)
			recovery := service.NewFlowRecovery(services.DrillService, drillRepo, stepRepo, flowCommandRepo)

			var executor worker.Executor
			if eventBus != nil {
				executor = service.NewFlowCommandExecutor(
					repository.DB,
					flowCommandRepo,
					services.DrillService,
					services.TaskService,
					eventBus,
					lease,
				)
			}

			flowWorker = worker.NewWorker(
				worker.Config{
					LeaseTTL:      cfg.Worker.LeaseTTL,
					RenewInterval: cfg.Worker.RenewInterval,
					CommandLease:  60 * time.Second,
					IdlePoll:      500 * time.Millisecond,
				},
				lease,
				flowCommandRepo,
				recovery,
				executor,
				cfg.InstanceID,
			)
			services.SetWorker(flowWorker)

			workerCtx, workerCancel := context.WithCancel(context.Background())
			defer workerCancel()
			go func() {
				if err := flowWorker.Run(workerCtx); err != nil {
					log.Printf("Worker 退出: %v", err)
				}
			}()
			log.Println("Worker 已启动")
		}
	}

	// 10. Health handler. The readiness flag starts false so /ready returns
	// 503 until SetReady(true) is called after the server starts listening.
	var dbChecker health.HealthChecker
	if repository.DB != nil {
		dbChecker = &dbHealthChecker{db: repository.DB}
	}
	var redisChecker health.HealthChecker
	if redisClient != nil {
		redisChecker = redisClient
	}
	// Avoid the typed-nil interface pitfall: a nil *worker.Worker assigned to
	// an interface is non-nil, so guard the conversion explicitly.
	var workerArg health.WorkerStatus
	if flowWorker != nil {
		workerArg = flowWorker
	}
	healthHandler := health.NewHandler(
		dbChecker,
		redisChecker,
		subscriber,
		workerArg,
		cfg.AppRole,
		cfg.InstanceID,
	)

	// 11. Router and HTTP server.
	r := router.SetupRouter(services, wsManager, cfg.JWT.Secret, loginLogger, healthHandler)
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// 12. Start HTTP server in a goroutine.
	serverErr := make(chan error, 1)
	go func() {
		log.Printf("服务启动在 %s (mode=%s, role=%s)", addr, cfg.Server.Mode, cfg.AppRole)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	// 13. Mark ready once the server is listening.
	healthHandler.SetReady(true)
	log.Println("服务已就绪")

	// 14. Wait for SIGINT/SIGTERM or a server startup error.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case <-ctx.Done():
		log.Println("收到关闭信号，开始优雅关闭...")
	case err := <-serverErr:
		log.Fatalf("服务启动失败: %v", err)
	}

	// 15. Graceful shutdown sequence:
	//   a. mark readiness false so load balancers drain traffic;
	//   b. stop the Worker and release its lease;
	//   c. cancel the events subscriber;
	//   d. shut down the HTTP server with a timeout (closes WebSocket conns);
	//   e. close the login logger;
	//   f. close Redis;
	//   g. close MySQL.
	healthHandler.SetReady(false)
	log.Println("已标记为未就绪")

	if flowWorker != nil {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		if err := flowWorker.Shutdown(shutdownCtx); err != nil {
			log.Printf("Worker 关闭失败: %v", err)
		}
		shutdownCancel()
		log.Println("Worker 已停止")
	}

	subCancel()
	log.Println("事件订阅已停止")

	httpCtx, httpCancel := context.WithTimeout(context.Background(), 30*time.Second)
	if err := srv.Shutdown(httpCtx); err != nil {
		log.Printf("HTTP 服务关闭失败: %v", err)
	}
	httpCancel()
	log.Println("HTTP 服务已停止")

	if loginLogger != nil {
		loginLogger.Close()
	}

	if redisClient != nil {
		if err := redisClient.Close(); err != nil {
			log.Printf("Redis 关闭失败: %v", err)
		}
		log.Println("Redis 已关闭")
	}

	if err := repository.Close(); err != nil {
		log.Printf("数据库关闭失败: %v", err)
	}
	log.Println("数据库已关闭")

	log.Println("优雅关闭完成")
}
