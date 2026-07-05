// =============================================================================
// deploy-ha: 一站式多活高可用部署工具
//
// 用法（项目根目录下执行）:
//
//   # 全流程：迁移数据库 + 构建并启动 HA 集群 + 健康检查
//   go run ./cmd/deploy-ha
//
//   # 仅执行数据库迁移到多活 schema（幂等，可重复执行）
//   go run ./cmd/deploy-ha migrate
//
//   # 仅启动 docker-compose.ha.yml（需先完成 migrate）
//   go run ./cmd/deploy-ha deploy
//
//   # 仅做部署后健康检查
//   go run ./cmd/deploy-ha verify
//
//   # 跳过 docker 构建（用已有镜像）
//   go run ./cmd/deploy-ha --no-build
//
//   # 指定 .env 文件路径（默认 ./ 或 ./.env）
//   go run ./cmd/deploy-ha --env /etc/drill/.env
//
// 前置条件:
//   - 已上传最新代码到服务器
//   - 服务器已安装 Go 1.23+ 与 docker / docker compose v2
//   - 外部 MySQL HA 与 Redis Cluster 已就绪
//   - 项目根目录存在 .env 文件（首次执行会从 .env.example 复制并提示填写）
//
// 行为:
//   1. 解析 .env 加载外部依赖连接信息
//   2. 连接 MySQL，幂等执行所有 schema 变更（业务字段重命名 + 多活核心表）
//   3. 校验多活 schema 关键对象已就位（失败则中止）
//   4. 调用 docker compose -f docker-compose.ha.yml up -d --build
//   5. 轮询各 backend 节点 /ready，等待 leader 选举完成（最长 90s）
// =============================================================================
package main

import (
	"bufio"
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/yaml.v3"
)

// -----------------------------------------------------------------------------
// 配置加载
// -----------------------------------------------------------------------------

// Config 携带部署相关的最小字段集合。
type Config struct {
	DatabaseHost      string
	DatabasePort      int
	DatabaseUser      string
	DatabasePassword  string
	DatabaseName      string
	RedisClusterAddrs string
	RedisPassword     string
	RedisTLS          bool
	JWTSecret         string
	PublicBaseURL     string

	// 配置来源描述（用于日志）
	source string
	// 原始 env 切片，用于传递给 docker compose 子进程
	rawEnv []string
}

// yamlConfig 映射 configs/config.yaml 的结构，只读部署需要的字段。
type yamlConfig struct {
	AppRole       string `yaml:"app_role"`
	InstanceID    string `yaml:"instance_id"`
	PublicBaseURL string `yaml:"public_base_url"`
	Server        struct {
		Port int    `yaml:"port"`
		Mode string `yaml:"mode"`
	} `yaml:"server"`
	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Name     string `yaml:"name"`
	} `yaml:"database"`
	Redis struct {
		Addr         string `yaml:"addr"`
		Host         string `yaml:"host"`
		Port         int    `yaml:"port"`
		Password     string `yaml:"password"`
		ClusterAddrs string `yaml:"cluster_addrs"`
		TLS          bool   `yaml:"tls"`
	} `yaml:"redis"`
	JWT struct {
		Secret string `yaml:"secret"`
		Expire int    `yaml:"expire"`
	} `yaml:"jwt"`
}

func loadEnvFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	out := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		eq := strings.IndexByte(line, '=')
		if eq < 0 {
			continue
		}
		key := strings.TrimSpace(line[:eq])
		val := strings.TrimSpace(line[eq+1:])
		// 去掉两端引号
		if len(val) >= 2 {
			if (val[0] == '"' && val[len(val)-1] == '"') ||
				(val[0] == '\'' && val[len(val)-1] == '\'') {
				val = val[1 : len(val)-1]
			}
		}
		out[key] = val
	}
	return out, scanner.Err()
}

// resolveEnvFile 找到 .env 文件路径；不存在则从 .env.example 复制并提示用户填写。
func resolveEnvFile(explicit string) (string, error) {
	if explicit != "" {
		if _, err := os.Stat(explicit); err != nil {
			return "", fmt.Errorf("--env 指定的文件不存在: %s\n请先 cp .env.example %s 并填写真实值", explicit, explicit)
		}
		return explicit, nil
	}

	// 默认查找顺序: ./.env → ../../.env
	candidates := []string{".env", "../../.env"}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			abs, _ := filepath.Abs(c)
			return abs, nil
		}
	}

	// 不存在则从 .env.example 复制
	if _, err := os.Stat(".env.example"); err == nil {
		if err := copyFile(".env.example", ".env"); err != nil {
			return "", fmt.Errorf("从 .env.example 复制 .env 失败: %v", err)
		}
		fmt.Println("⚠ 未找到 .env，已从 .env.example 复制一份。")
		fmt.Println("  请编辑 .env 填写 DATABASE_HOST / REDIS_CLUSTER_ADDRS / JWT_SECRET 后重新执行本工具。")
		os.Exit(1)
	}
	return "", fmt.Errorf("未找到 .env 文件，且当前目录无 .env.example 可作为模板")
}

func copyFile(src, dst string) error {
	in, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, in, 0o644)
}

// loadYAMLConfig 从 YAML 文件读取配置。
func loadYAMLConfig(path string) (*yamlConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var yc yamlConfig
	if err := yaml.Unmarshal(data, &yc); err != nil {
		return nil, fmt.Errorf("解析 YAML 失败: %v", err)
	}
	return &yc, nil
}

// buildConfigFromYAML 从 YAML 配置构建 Config。
func buildConfigFromYAML(yc *yamlConfig) *Config {
	cfg := &Config{
		DatabaseHost:      yc.Database.Host,
		DatabasePort:      yc.Database.Port,
		DatabaseUser:      yc.Database.User,
		DatabasePassword:  yc.Database.Password,
		DatabaseName:      yc.Database.Name,
		RedisClusterAddrs: yc.Redis.ClusterAddrs,
		RedisPassword:    yc.Redis.Password,
		RedisTLS:         yc.Redis.TLS,
		JWTSecret:        yc.JWT.Secret,
		PublicBaseURL:    yc.PublicBaseURL,
	}
	if cfg.DatabasePort == 0 {
		cfg.DatabasePort = 3306
	}
	// Redis: cluster_addrs 优先于 addr
	if cfg.RedisClusterAddrs == "" && yc.Redis.Addr != "" {
		cfg.RedisClusterAddrs = yc.Redis.Addr
	}
	if cfg.RedisClusterAddrs == "" && yc.Redis.Host != "" {
		port := yc.Redis.Port
		if port == 0 {
			port = 6379
		}
		cfg.RedisClusterAddrs = fmt.Sprintf("%s:%d", yc.Redis.Host, port)
	}
	return cfg
}

// resolveYAMLFile 查找 configs/config.yaml。
func resolveYAMLFile(explicit string) (string, error) {
	if explicit != "" {
		if _, err := os.Stat(explicit); err != nil {
			return "", fmt.Errorf("--config 指定的文件不存在: %s", explicit)
		}
		return explicit, nil
	}
	candidates := []string{
		"configs/config.yaml",
		"configs/config.yml",
		"config.yaml",
		"config.yml",
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			abs, _ := filepath.Abs(c)
			return abs, nil
		}
	}
	return "", nil // 不存在时返回空串，调用方决定是否 fallback
}

// buildConfig 从 env map 解析 Config。
func buildConfig(env map[string]string) (*Config, error) {
	cfg := &Config{
		DatabaseHost:      env["DATABASE_HOST"],
		DatabaseUser:      env["DATABASE_USER"],
		DatabasePassword:  env["DATABASE_PASSWORD"],
		DatabaseName:      env["DATABASE_NAME"],
		RedisClusterAddrs: env["REDIS_CLUSTER_ADDRS"],
		RedisPassword:    env["REDIS_PASSWORD"],
		JWTSecret:         env["JWT_SECRET"],
		PublicBaseURL:     env["PUBLIC_BASE_URL"],
	}

	portStr := env["DATABASE_PORT"]
	if portStr == "" {
		portStr = "3306"
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("DATABASE_PORT 非法: %q", portStr)
	}
	cfg.DatabasePort = port

	if v, ok := env["REDIS_TLS"]; ok {
		b, _ := strconv.ParseBool(strings.ToLower(v))
		cfg.RedisTLS = b
	}

	// 传递原始 env 给 docker compose（保留全部变量）
	for k, v := range env {
		cfg.rawEnv = append(cfg.rawEnv, fmt.Sprintf("%s=%s", k, v))
	}
	return cfg, nil
}

// validateForMigrate 检查迁移所需的 DB 连接字段。
func (c *Config) validateForMigrate() error {
	var missing []string
	if c.DatabaseHost == "" {
		missing = append(missing, "database.host (或 DATABASE_HOST)")
	}
	if c.DatabaseUser == "" {
		missing = append(missing, "database.user (或 DATABASE_USER)")
	}
	if c.DatabaseName == "" {
		missing = append(missing, "database.name (或 DATABASE_NAME)")
	}
	if len(missing) > 0 {
		return fmt.Errorf("缺少必填字段: %s\n请在 configs/config.yaml 或 .env 中填写 MySQL 连接信息", strings.Join(missing, ", "))
	}
	return nil
}

// validateForDeploy 检查启动 HA 集群所需的全部字段。
func (c *Config) validateForDeploy() error {
	if err := c.validateForMigrate(); err != nil {
		return err
	}
	var missing []string
	if c.RedisClusterAddrs == "" {
		missing = append(missing, "redis.cluster_addrs (或 REDIS_CLUSTER_ADDRS)")
	}
	if c.JWTSecret == "" || c.JWTSecret == "your-secret-key-change-in-production" {
		missing = append(missing, "jwt.secret (或 JWT_SECRET)")
	}
	if len(missing) > 0 {
		return fmt.Errorf("缺少必填字段: %s\n请在 configs/config.yaml 或 .env 中填写", strings.Join(missing, ", "))
	}
	return nil
}

// -----------------------------------------------------------------------------
// 数据库连接
// -----------------------------------------------------------------------------

func buildDSN(c *Config) string {
	// go-sql-driver/mysql DSN: user:pass@tcp(host:port)/dbname?params
	// 密码中可能含 @ : / 等特殊字符，使用 url.QueryEscape 转义
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&multiStatements=true&timeout=10s",
		url.QueryEscape(c.DatabaseUser),
		url.QueryEscape(c.DatabasePassword),
		c.DatabaseHost,
		c.DatabasePort,
		c.DatabaseName,
	)
}

func connectDB(c *Config) (*sql.DB, error) {
	db, err := sql.Open("mysql", buildDSN(c))
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(2)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("无法连接 MySQL %s:%d/%s: %v\n请确认 configs/config.yaml 或 .env 中数据库配置正确，且 MySQL 已允许远端访问",
			c.DatabaseHost, c.DatabasePort, c.DatabaseName, err)
	}
	return db, nil
}

// -----------------------------------------------------------------------------
// Schema 检查辅助
// -----------------------------------------------------------------------------

func tableExists(db *sql.DB, table string) (bool, error) {
	var n int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?
	`, table).Scan(&n)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func columnExists(db *sql.DB, table, column string) (bool, error) {
	var n int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?
	`, table, column).Scan(&n)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func indexExists(db *sql.DB, table, index string) (bool, error) {
	var n int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND INDEX_NAME = ?
	`, table, index).Scan(&n)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// addColumnIfMissing 仅当列不存在时执行 ALTER ADD COLUMN。
func addColumnIfMissing(db *sql.DB, table, column, ddl string) error {
	ok, err := columnExists(db, table, column)
	if err != nil {
		return fmt.Errorf("检查 %s.%s 失败: %v", table, column, err)
	}
	if ok {
		fmt.Printf("  ✓ 跳过 %s.%s (已存在)\n", table, column)
		return nil
	}
	if _, err := db.Exec(ddl); err != nil {
		return fmt.Errorf("执行失败 %s: %v", ddl, err)
	}
	fmt.Printf("  ✓ 添加 %s.%s\n", table, column)
	return nil
}

// addIndexIfMissing 仅当索引不存在时执行 ALTER ADD KEY/UNIQUE KEY。
func addIndexIfMissing(db *sql.DB, table, index, ddl string) error {
	ok, err := indexExists(db, table, index)
	if err != nil {
		return fmt.Errorf("检查 %s.%s 索引失败: %v", table, index, err)
	}
	if ok {
		fmt.Printf("  ✓ 跳过 %s.%s 索引 (已存在)\n", table, index)
		return nil
	}
	if _, err := db.Exec(ddl); err != nil {
		return fmt.Errorf("执行失败 %s: %v", ddl, err)
	}
	fmt.Printf("  ✓ 添加 %s.%s 索引\n", table, index)
	return nil
}

// renameColumnIfNeeded 仅当旧列存在且新列不存在时执行 CHANGE COLUMN。
func renameColumnIfNeeded(db *sql.DB, table, oldCol, newCol, ddl string) error {
	hasOld, err := columnExists(db, table, oldCol)
	if err != nil {
		return err
	}
	if !hasOld {
		fmt.Printf("  ✓ 跳过 %s.%s→%s (旧列不存在，可能已重命名)\n", table, oldCol, newCol)
		return nil
	}
	hasNew, err := columnExists(db, table, newCol)
	if err != nil {
		return err
	}
	if hasNew {
		fmt.Printf("  ✓ 跳过 %s.%s→%s (新列已存在)\n", table, oldCol, newCol)
		return nil
	}
	if _, err := db.Exec(ddl); err != nil {
		return fmt.Errorf("重命名失败 %s: %v", ddl, err)
	}
	fmt.Printf("  ✓ 重命名 %s.%s→%s\n", table, oldCol, newCol)
	return nil
}

// -----------------------------------------------------------------------------
// 迁移步骤
// -----------------------------------------------------------------------------

// runMigrations 执行所有 schema 变更。完全幂等。
func runMigrations(db *sql.DB) error {
	fmt.Println("\n[1/4] 业务字段重命名（当前代码已使用新字段名，必须迁移）")

	// B1. drill_instance.current_step_id → current_task_id
	if err := renameColumnIfNeeded(db, "drill_instance", "current_step_id", "current_task_id",
		"ALTER TABLE `drill_instance` CHANGE COLUMN `current_step_id` `current_task_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '当前激活步骤 ID'"); err != nil {
		return err
	}

	// B2. drill_instance_step.step_template_id → template_step_id + idx_step_template_id
	if err := renameColumnIfNeeded(db, "drill_instance_step", "step_template_id", "template_step_id",
		"ALTER TABLE `drill_instance_step` CHANGE COLUMN `step_template_id` `template_step_id` BIGINT UNSIGNED NOT NULL COMMENT '来源步骤模板 ID'"); err != nil {
		return err
	}
	if err := addIndexIfMissing(db, "drill_instance_step", "idx_step_template_id",
		"ALTER TABLE `drill_instance_step` ADD KEY `idx_step_template_id` (`template_step_id`)"); err != nil {
		return err
	}

	// B3. drill_instance_step.attributes → action_params
	if err := renameColumnIfNeeded(db, "drill_instance_step", "attributes", "action_params",
		"ALTER TABLE `drill_instance_step` CHANGE COLUMN `attributes` `action_params` JSON DEFAULT NULL COMMENT '动态扩展属性'"); err != nil {
		return err
	}

	// B4. drill_instance_step_log.step_instance_id → task_instance_id
	if err := renameColumnIfNeeded(db, "drill_instance_step_log", "step_instance_id", "task_instance_id",
		"ALTER TABLE `drill_instance_step_log` CHANGE COLUMN `step_instance_id` `task_instance_id` BIGINT UNSIGNED DEFAULT NULL"); err != nil {
		return err
	}

	// B5. drill_template.phase_order
	if err := addColumnIfMissing(db, "drill_template", "phase_order",
		"ALTER TABLE `drill_template` ADD COLUMN `phase_order` JSON DEFAULT NULL COMMENT '阶段顺序列表' AFTER `created_by`"); err != nil {
		return err
	}

	// B6. 索引补齐
	if err := addIndexIfMissing(db, "user", "idx_role",
		"ALTER TABLE `user` ADD KEY `idx_role` (`role`)"); err != nil {
		return err
	}
	if err := addIndexIfMissing(db, "user", "idx_department",
		"ALTER TABLE `user` ADD KEY `idx_department` (`department`)"); err != nil {
		return err
	}
	if err := addIndexIfMissing(db, "drill_template", "idx_created_by",
		"ALTER TABLE `drill_template` ADD KEY `idx_created_by` (`created_by`)"); err != nil {
		return err
	}

	fmt.Println("\n[2/4] 创建多活核心表（drill_flow_command + drill_worker_epoch）")

	// A1. drill_flow_command（CREATE TABLE IF NOT EXISTS 保证幂等）
	if _, err := db.Exec(`
CREATE TABLE IF NOT EXISTS ` + "`drill_flow_command`" + ` (
    ` + "`id`" + ` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    ` + "`command_type`" + ` VARCHAR(64) NOT NULL,
    ` + "`drill_instance_id`" + ` BIGINT UNSIGNED NOT NULL,
    ` + "`step_instance_id`" + ` BIGINT UNSIGNED DEFAULT NULL,
    ` + "`operator_id`" + ` BIGINT UNSIGNED NOT NULL,
    ` + "`idempotency_key`" + ` VARCHAR(128) NOT NULL,
    ` + "`payload`" + ` JSON NOT NULL,
    ` + "`status`" + ` VARCHAR(20) NOT NULL,
    ` + "`worker_id`" + ` VARCHAR(128) DEFAULT NULL,
    ` + "`worker_epoch`" + ` BIGINT UNSIGNED NOT NULL DEFAULT 0,
    ` + "`lease_token`" + ` VARCHAR(128) NOT NULL DEFAULT '',
    ` + "`lease_until`" + ` DATETIME DEFAULT NULL,
    ` + "`attempts`" + ` INT NOT NULL DEFAULT 0,
    ` + "`attempt_count`" + ` INT NOT NULL DEFAULT 0,
    ` + "`result`" + ` JSON DEFAULT NULL,
    ` + "`error_code`" + ` VARCHAR(64) DEFAULT NULL,
    ` + "`error_message`" + ` VARCHAR(500) DEFAULT NULL,
    ` + "`created_at`" + ` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ` + "`started_at`" + ` DATETIME DEFAULT NULL,
    ` + "`finished_at`" + ` DATETIME DEFAULT NULL,
    ` + "`updated_at`" + ` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (` + "`id`" + `),
    UNIQUE KEY ` + "`uk_flow_command_idempotency`" + ` (` + "`idempotency_key`" + `),
    KEY ` + "`idx_flow_command_pending`" + ` (` + "`status`" + `, ` + "`created_at`" + `),
    KEY ` + "`idx_flow_command_lease`" + ` (` + "`lease_until`" + `)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='流程命令表'`); err != nil {
		return fmt.Errorf("创建 drill_flow_command 失败: %v", err)
	}
	fmt.Println("  ✓ drill_flow_command 表已就绪")

	// 兼容已存在旧表（缺 worker_epoch/lease_token/attempt_count 列）
	if err := addColumnIfMissing(db, "drill_flow_command", "worker_epoch",
		"ALTER TABLE `drill_flow_command` ADD COLUMN `worker_epoch` BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `worker_id`"); err != nil {
		return err
	}
	if err := addColumnIfMissing(db, "drill_flow_command", "lease_token",
		"ALTER TABLE `drill_flow_command` ADD COLUMN `lease_token` VARCHAR(128) NOT NULL DEFAULT '' AFTER `worker_epoch`"); err != nil {
		return err
	}
	if err := addColumnIfMissing(db, "drill_flow_command", "attempt_count",
		"ALTER TABLE `drill_flow_command` ADD COLUMN `attempt_count` INT NOT NULL DEFAULT 0 AFTER `attempts`"); err != nil {
		return err
	}

	// A2. drill_worker_epoch
	if _, err := db.Exec(`
CREATE TABLE IF NOT EXISTS ` + "`drill_worker_epoch`" + ` (
    ` + "`id`" + ` BIGINT UNSIGNED NOT NULL,
    ` + "`worker_id`" + ` VARCHAR(128) NOT NULL COMMENT '当前持有 epoch 的 worker',
    ` + "`epoch`" + ` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '单调递增 epoch，每次领导切换 +1',
    ` + "`lease_until`" + ` DATETIME DEFAULT NULL COMMENT 'epoch 租约到期时间',
    ` + "`created_at`" + ` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ` + "`updated_at`" + ` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (` + "`id`" + `)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Worker epoch 单例行表'`); err != nil {
		return fmt.Errorf("创建 drill_worker_epoch 失败: %v", err)
	}
	fmt.Println("  ✓ drill_worker_epoch 表已就绪")

	fmt.Println("\n[3/4] 添加 command_id 字段 + 幂等唯一键")

	// A3. drill_instance_step_log.command_id + uk_log_command_action_step
	if err := addColumnIfMissing(db, "drill_instance_step_log", "command_id",
		"ALTER TABLE `drill_instance_step_log` ADD COLUMN `command_id` BIGINT UNSIGNED DEFAULT NULL AFTER `task_instance_id`"); err != nil {
		return err
	}
	if err := addLogUniqueKey(db); err != nil {
		return err
	}

	// A4. notification.command_id + uk_notification_command_user_type_step
	if err := addColumnIfMissing(db, "notification", "command_id",
		"ALTER TABLE `notification` ADD COLUMN `command_id` BIGINT UNSIGNED DEFAULT NULL AFTER `type`"); err != nil {
		return err
	}
	if err := addIndexIfMissing(db, "notification", "uk_notification_command_user_type_step",
		"ALTER TABLE `notification` ADD UNIQUE KEY `uk_notification_command_user_type_step` (`command_id`, `user_id`, `type`, `step_id`)"); err != nil {
		return err
	}

	fmt.Println("\n[4/4] 校验多活 schema 关键对象")
	return verifySchema(db)
}

// addLog_unique_key 自动适配列名（task_instance_id 或 step_instance_id）。
func addLogUniqueKey(db *sql.DB) error {
	ok, err := indexExists(db, "drill_instance_step_log", "uk_log_command_action_step")
	if err != nil {
		return err
	}
	if ok {
		fmt.Println("  ✓ 跳过 drill_instance_step_log.uk_log_command_action_step (已存在)")
		return nil
	}

	hasTask, err := columnExists(db, "drill_instance_step_log", "task_instance_id")
	if err != nil {
		return err
	}
	col := "step_instance_id"
	if hasTask {
		col = "task_instance_id"
	}
	ddl := fmt.Sprintf(
		"ALTER TABLE `drill_instance_step_log` ADD UNIQUE KEY `uk_log_command_action_step` (`command_id`, `action`, `%s`)",
		col,
	)
	if _, err := db.Exec(ddl); err != nil {
		return fmt.Errorf("添加 uk_log_command_action_step 失败: %v", err)
	}
	fmt.Printf("  ✓ 添加 drill_instance_step_log.uk_log_command_action_step (列=%s)\n", col)
	return nil
}

// verifySchema 校验所有多活核心对象已就位。
func verifySchema(db *sql.DB) error {
	checks := []struct {
		desc string
		ok  bool
	}{
		{"drill_flow_command 表", mustTableExist(db, "drill_flow_command")},
		{"drill_worker_epoch 表", mustTableExist(db, "drill_worker_epoch")},
		{"drill_flow_command.worker_epoch 列", mustColumnExist(db, "drill_flow_command", "worker_epoch")},
		{"drill_flow_command.lease_token 列", mustColumnExist(db, "drill_flow_command", "lease_token")},
		{"drill_flow_command.attempt_count 列", mustColumnExist(db, "drill_flow_command", "attempt_count")},
		{"drill_instance_step_log.command_id 列", mustColumnExist(db, "drill_instance_step_log", "command_id")},
		{"notification.command_id 列", mustColumnExist(db, "notification", "command_id")},
	}
	allOK := true
	for _, c := range checks {
		if c.ok {
			fmt.Printf("  ✓ %s\n", c.desc)
		} else {
			fmt.Printf("  ✗ %s\n", c.desc)
			allOK = false
		}
	}
	if !allOK {
		return fmt.Errorf("schema 校验失败，多活表结构不完整")
	}
	fmt.Println("  → 多活表结构校验通过")
	return nil
}

func mustTableExist(db *sql.DB, table string) bool {
	ok, err := tableExists(db, table)
	return err == nil && ok
}

func mustColumnExist(db *sql.DB, table, column string) bool {
	ok, err := columnExists(db, table, column)
	return err == nil && ok
}

// -----------------------------------------------------------------------------
// Docker Compose 部署
// -----------------------------------------------------------------------------

func runDockerCompose(cfg *Config, noBuild bool) error {
	composeFile := "docker-compose.ha.yml"
	if _, err := os.Stat(composeFile); err != nil {
		return fmt.Errorf("未找到 %s，请在项目根目录执行本工具", composeFile)
	}

	// 检查 docker / docker compose 可用
	if err := exec.Command("docker", "version").Run(); err != nil {
		return fmt.Errorf("docker 命令不可用: %v\n请先安装 Docker Engine", err)
	}
	if err := exec.Command("docker", "compose", "version").Run(); err != nil {
		return fmt.Errorf("docker compose v2 不可用: %v\n请安装 docker compose v2 插件", err)
	}

	args := []string{"compose", "-f", composeFile}
	// .env 文件由 docker compose 自动加载，但显式指定更稳妥
	args = append(args, "--env-file", ".env")
	args = append(args, "up", "-d")
	if !noBuild {
		args = append(args, "--build")
	}

	fmt.Printf("\n$ docker %s\n", strings.Join(args, " "))
	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), cfg.rawEnv...)
	return cmd.Run()
}

// -----------------------------------------------------------------------------
// 健康检查
// -----------------------------------------------------------------------------

// waitForReady 轮询各 backend 节点的 /ready 端点。
// 容器名 backend-a/b/c 内部端口 8080，但宿主通过 nginx 80 端口访问。
// 这里直接 docker exec 进每个容器探活，最可靠。
func waitForReady(timeout time.Duration) error {
	fmt.Println("\n等待各 backend 节点就绪...")

	containers := []string{"drill-backend-a", "drill-backend-b", "drill-backend-c"}
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		allReady := true
		for _, c := range containers {
			status := probeContainer(c)
			fmt.Printf("  %s: %s\n", c, status)
			if !strings.Contains(status, "ready") || strings.Contains(status, "unready") {
				allReady = false
			}
		}
		if allReady {
			fmt.Println("  ✓ 所有节点 /ready 返回正常")
			return nil
		}
		time.Sleep(3 * time.Second)
		fmt.Println("  ... 重试中")
	}
	return fmt.Errorf("等待节点就绪超时（%v）", timeout)
}

// probeContainer 用 docker exec 在容器内执行 wget 探活。
func probeContainer(container string) string {
	// 先检查容器是否存在且 running
	ps, err := exec.Command("docker", "ps", "--filter", "name="+container,
		"--format", "{{.Status}}").Output()
	if err != nil || strings.TrimSpace(string(ps)) == "" {
		return "未运行"
	}

	// 探 /ready
	out, err := exec.Command("docker", "exec", container,
		"wget", "-q", "-O-", "-T", "2", "http://localhost:8080/ready").Output()
	if err != nil {
		return "unready (wget 失败)"
	}
	s := string(out)
	if strings.Contains(s, "\"ready\":true") || strings.Contains(s, "leader-ready") ||
		strings.Contains(s, "standby-ready") || strings.Contains(s, "\"ready\": true") {
		// 提取 worker_status 字段
		if idx := strings.Index(s, "\"worker_status\""); idx >= 0 {
			rest := s[idx:]
			end := strings.Index(rest, "\",")
			if end > 0 {
				return strings.TrimSpace(rest[:end]) + "\""
			}
		}
		return "ready"
	}
	return "unready: " + strings.TrimSpace(s)
}

// reportStatus 汇报集群拓扑与后续操作建议。
func reportStatus() {
	fmt.Println("\n============================================================")
	fmt.Println("✓ 多活高可用集群部署完成")
	fmt.Println("============================================================")
	fmt.Println("架构概览:")
	fmt.Println("  nginx (80) → backend-a / backend-b / backend-c (8080)")
	fmt.Println("  共享 MySQL + Redis Cluster (由外部提供 HA)")
	fmt.Println()
	fmt.Println("常用运维命令:")
	fmt.Println("  # 查看日志")
	fmt.Println("  docker compose -f docker-compose.ha.yml logs -f")
	fmt.Println()
	fmt.Println("  # 查看节点状态")
	fmt.Println("  docker exec drill-backend-a wget -qO- http://localhost:8080/ready")
	fmt.Println("  docker exec drill-backend-b wget -qO- http://localhost:8080/ready")
	fmt.Println("  docker exec drill-backend-c wget -qO- http://localhost:8080/ready")
	fmt.Println()
	fmt.Println("  # 验证故障切换")
	fmt.Println("  bash scripts/test-ha.sh")
	fmt.Println()
	fmt.Println("访问入口:")
	fmt.Println("  应用:   http://localhost/")
	fmt.Println("  API:    http://localhost/api/v1/...")
	fmt.Println("  健康检查: http://localhost/health")
	fmt.Println("============================================================")
}

// -----------------------------------------------------------------------------
// 命令入口
// -----------------------------------------------------------------------------

func printHelp() {
	fmt.Println(`deploy-ha: 一站式多活高可用部署工具

用法:
  go run ./cmd/deploy-ha [子命令] [参数]

子命令:
  all       (默认) 迁移 + 部署 + 健康检查
  migrate   仅执行数据库迁移到多活 schema (幂等)
  deploy    仅启动 docker-compose.ha.yml
  verify    仅做部署后健康检查

配置来源（优先级从高到低）:
  1. --config PATH   显式指定 YAML 配置文件
  2. configs/config.yaml / config.yaml  自动查找
  3. --env PATH      显式指定 .env 文件
  4. .env            自动查找

参数:
  --config PATH   指定 YAML 配置文件路径
  --env PATH      指定 .env 文件路径
  --no-build      启动集群时跳过镜像构建
  --timeout DUR   健康检查超时 (默认 90s)

示例:
  go run ./cmd/deploy-ha                        # 自动找 configs/config.yaml
  go run ./cmd/deploy-ha migrate                # 只迁移，从 YAML 读 DB 配置
  go run ./cmd/deploy-ha --config /etc/drill/config.yaml migrate
  go run ./cmd/deploy-ha deploy --no-build
  go run ./cmd/deploy-ha verify --timeout 120s`)
}

func parseFlags(args []string) (configFile string, envFile string, noBuild bool, timeout time.Duration) {
	fs := flag.NewFlagSet("deploy-ha", flag.ExitOnError)
	fs.StringVar(&configFile, "config", "", "YAML 配置文件路径 (如 configs/config.yaml)")
	fs.StringVar(&envFile, "env", "", ".env 文件路径")
	fs.BoolVar(&noBuild, "no-build", false, "跳过 docker build")
	timeout = 90 * time.Second
	var timeoutStr string
	fs.StringVar(&timeoutStr, "timeout", "90s", "健康检查超时")
	_ = fs.Parse(args)
	if d, err := time.ParseDuration(timeoutStr); err == nil {
		timeout = d
	}
	return
}

func runMigrateCmd(args []string) {
	configFile, envFile, _, _ := parseFlags(args)
	cfg, err := loadConfig(configFile, envFile)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	if err := cfg.validateForMigrate(); err != nil {
		log.Fatalf("配置校验失败: %v", err)
	}
	db, err := connectDB(cfg)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer db.Close()

	if err := runMigrations(db); err != nil {
		log.Fatalf("迁移失败: %v", err)
	}
	fmt.Println("\n✓ 数据库迁移完成")
}

func runDeployCmd(args []string) {
	configFile, envFile, noBuild, _ := parseFlags(args)
	cfg, err := loadConfig(configFile, envFile)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	if err := cfg.validateForDeploy(); err != nil {
		log.Fatalf("配置校验失败: %v", err)
	}
	if err := runDockerCompose(cfg, noBuild); err != nil {
		log.Fatalf("docker compose 启动失败: %v", err)
	}
	fmt.Println("\n✓ docker compose 已启动")
}

func runVerifyCmd(args []string) {
	_, _, _, timeout := parseFlags(args)
	if err := waitForReady(timeout); err != nil {
		log.Fatalf("健康检查失败: %v", err)
	}
	reportStatus()
}

func runAllCmd(args []string) {
	configFile, envFile, noBuild, timeout := parseFlags(args)
	cfg, err := loadConfig(configFile, envFile)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	if err := cfg.validateForDeploy(); err != nil {
		log.Fatalf("配置校验失败: %v", err)
	}

	// Step 1: 迁移
	db, err := connectDB(cfg)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	if err := runMigrations(db); err != nil {
		db.Close()
		log.Fatalf("迁移失败: %v", err)
	}
	db.Close()
	fmt.Println("\n✓ Step 1/3 数据库迁移完成")

	// Step 2: 部署
	if err := runDockerCompose(cfg, noBuild); err != nil {
		log.Fatalf("docker compose 启动失败: %v", err)
	}
	fmt.Println("\n✓ Step 2/3 docker compose 已启动")

	// Step 3: 健康检查
	if err := waitForReady(timeout); err != nil {
		log.Fatalf("健康检查失败: %v\n可手动重试: go run ./cmd/deploy-ha verify", err)
	}
	fmt.Println("\n✓ Step 3/3 健康检查通过")
	reportStatus()
}

// loadConfig 按优先级加载配置：YAML → .env → 报错。
// 环境变量始终可以覆盖文件中的值（与 server 行为一致）。
func loadConfig(configFile string, envFile string) (*Config, error) {
	// 1. 尝试从 YAML 加载
	yamlPath, err := resolveYAMLFile(configFile)
	if err != nil {
		return nil, err
	}

	var cfg *Config

	if yamlPath != "" {
		yc, err := loadYAMLConfig(yamlPath)
		if err != nil {
			return nil, fmt.Errorf("读取 %s 失败: %v", yamlPath, err)
		}
		cfg = buildConfigFromYAML(yc)
		cfg.source = yamlPath
		fmt.Printf("已加载配置: %s\n", yamlPath)
	}

	// 2. 尝试从 .env 加载（仅在 YAML 不存在时作为主配置源）
	// 如果 YAML 已加载，不再读 .env（YAML 优先级高于 .env）
	if cfg == nil {
		envPath, err := resolveEnvFile(envFile)
		if err != nil {
			return nil, err
		}
		if envPath != "" {
			env, err := loadEnvFile(envPath)
			if err != nil {
				return nil, fmt.Errorf("读取 %s 失败: %v", envPath, err)
			}
			envCfg, err := buildConfig(env)
			if err != nil {
				return nil, err
			}
			cfg = envCfg
			cfg.source = envPath
			fmt.Printf("已加载配置: %s\n", envPath)
		}
	}

	if cfg == nil {
		return nil, fmt.Errorf("未找到任何配置文件\n请准备 configs/config.yaml 或 .env（参考 .env.example）")
	}

	// 3. 环境变量始终可以覆盖文件配置（与 drill-server 行为一致）
	applyOSEnviron(cfg)

	fmt.Printf("  MySQL: %s:%d/%s\n  Redis: %s\n",
		cfg.DatabaseHost, cfg.DatabasePort, cfg.DatabaseName,
		maskAddr(cfg.RedisClusterAddrs))
	return cfg, nil
}

// applyOSEnviron 用当前进程的环境变量覆盖配置（与 drill-server 的 applyEnvOverrides 行为一致）。
func applyOSEnviron(cfg *Config) {
	if v := os.Getenv("DATABASE_HOST"); v != "" {
		cfg.DatabaseHost = v
	}
	if v := os.Getenv("DATABASE_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.DatabasePort = port
		}
	}
	if v := os.Getenv("DATABASE_USER"); v != "" {
		cfg.DatabaseUser = v
	}
	if v := os.Getenv("DATABASE_PASSWORD"); v != "" {
		cfg.DatabasePassword = v
	}
	if v := os.Getenv("DATABASE_NAME"); v != "" {
		cfg.DatabaseName = v
	}
	if v := os.Getenv("REDIS_CLUSTER_ADDRS"); v != "" {
		cfg.RedisClusterAddrs = v
	}
	if v := os.Getenv("REDIS_PASSWORD"); v != "" {
		cfg.RedisPassword = v
	}
	if v := os.Getenv("JWT_SECRET"); v != "" {
		cfg.JWTSecret = v
	}
}

// maskAddr 隐藏密码细节，仅显示地址。
func maskAddr(s string) string {
	if s == "" {
		return "<未配置>"
	}
	return s
}

func main() {
	log.SetFlags(0)

	args := os.Args[1:]
	if len(args) == 0 {
		runAllCmd(args)
		return
	}

	switch args[0] {
	case "migrate":
		runMigrateCmd(args[1:])
	case "deploy":
		runDeployCmd(args[1:])
	case "verify":
		runVerifyCmd(args[1:])
	case "all":
		runAllCmd(args[1:])
	case "-h", "--help", "help":
		printHelp()
	default:
		// 兼容直接传 flag（无子命令）的情况
		runAllCmd(args)
	}
}
