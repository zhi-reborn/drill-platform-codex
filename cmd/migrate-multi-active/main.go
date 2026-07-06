package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"drill-platform/internal/pkg/appconfig"

	"github.com/go-sql-driver/mysql"
)

type dbConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
}

func main() {
	log.SetFlags(0)

	var configPath string
	var timeout time.Duration
	flag.StringVar(&configPath, "config", "configs/config.yaml", "配置文件路径")
	flag.DurationVar(&timeout, "timeout", 30*time.Second, "数据库连接超时")
	flag.Parse()

	cfg, err := loadDBConfig(configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	db, err := connect(cfg, timeout)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()

	fmt.Printf("已连接数据库: %s:%d/%s\n", cfg.Host, cfg.Port, cfg.Name)
	if err := runMultiActiveMigration(db); err != nil {
		log.Fatalf("迁移失败: %v", err)
	}
	fmt.Println("✓ 多活 schema 迁移完成")
}

func loadDBConfig(configPath string) (*dbConfig, error) {
	cfg, err := appconfig.Load(configPath)
	if err != nil {
		return nil, err
	}
	db := &dbConfig{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		Name:     cfg.Database.Name,
	}
	if v := os.Getenv("DATABASE_HOST"); v != "" {
		db.Host = v
	}
	if v := os.Getenv("DATABASE_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			db.Port = port
		}
	}
	if v := os.Getenv("DATABASE_USER"); v != "" {
		db.User = v
	}
	if v := os.Getenv("DATABASE_PASSWORD"); v != "" {
		db.Password = v
	}
	if v := os.Getenv("DATABASE_NAME"); v != "" {
		db.Name = v
	}
	if db.Host == "" || db.Port == 0 || db.User == "" || db.Name == "" {
		return nil, fmt.Errorf("缺少数据库配置，请设置 database.host/port/user/name 或 DATABASE_* 环境变量")
	}
	return db, nil
}

func connect(cfg *dbConfig, timeout time.Duration) (*sql.DB, error) {
	mysqlCfg := mysql.NewConfig()
	mysqlCfg.User = cfg.User
	mysqlCfg.Passwd = cfg.Password
	mysqlCfg.Net = "tcp"
	mysqlCfg.Addr = fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	mysqlCfg.DBName = cfg.Name
	mysqlCfg.Params = map[string]string{
		"charset":         "utf8mb4",
		"parseTime":       "True",
		"multiStatements": "true",
		"timeout":         "10s",
	}
	dsn := mysqlCfg.FormatDSN()

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func runMultiActiveMigration(db *sql.DB) error {
	fmt.Println("[1/4] 业务字段对齐")
	if err := renameColumnIfNeeded(db, "drill_instance", "current_step_id", "current_task_id",
		"ALTER TABLE `drill_instance` CHANGE COLUMN `current_step_id` `current_task_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '当前激活步骤 ID'"); err != nil {
		return err
	}
	if err := renameColumnIfNeeded(db, "drill_instance_step", "step_template_id", "template_step_id",
		"ALTER TABLE `drill_instance_step` CHANGE COLUMN `step_template_id` `template_step_id` BIGINT UNSIGNED NOT NULL COMMENT '来源步骤模板 ID'"); err != nil {
		return err
	}
	if err := addIndexIfMissing(db, "drill_instance_step", "idx_step_template_id",
		"ALTER TABLE `drill_instance_step` ADD KEY `idx_step_template_id` (`template_step_id`)"); err != nil {
		return err
	}
	if err := renameColumnIfNeeded(db, "drill_instance_step", "attributes", "action_params",
		"ALTER TABLE `drill_instance_step` CHANGE COLUMN `attributes` `action_params` JSON DEFAULT NULL COMMENT '动态扩展属性'"); err != nil {
		return err
	}
	if err := renameColumnIfNeeded(db, "drill_instance_step_log", "step_instance_id", "task_instance_id",
		"ALTER TABLE `drill_instance_step_log` CHANGE COLUMN `step_instance_id` `task_instance_id` BIGINT UNSIGNED DEFAULT NULL"); err != nil {
		return err
	}
	if err := addColumnIfMissing(db, "drill_template", "phase_order",
		"ALTER TABLE `drill_template` ADD COLUMN `phase_order` JSON DEFAULT NULL COMMENT '阶段顺序列表' AFTER `created_by`"); err != nil {
		return err
	}
	if err := addIndexIfMissing(db, "user", "idx_role", "ALTER TABLE `user` ADD KEY `idx_role` (`role`)"); err != nil {
		return err
	}
	if err := addIndexIfMissing(db, "user", "idx_department", "ALTER TABLE `user` ADD KEY `idx_department` (`department`)"); err != nil {
		return err
	}
	if err := addIndexIfMissing(db, "drill_template", "idx_created_by", "ALTER TABLE `drill_template` ADD KEY `idx_created_by` (`created_by`)"); err != nil {
		return err
	}

	fmt.Println("[2/4] 多活核心表")
	if _, err := db.Exec(createFlowCommandTableSQL); err != nil {
		return fmt.Errorf("创建 drill_flow_command 失败: %w", err)
	}
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
	if _, err := db.Exec(createWorkerEpochTableSQL); err != nil {
		return fmt.Errorf("创建 drill_worker_epoch 失败: %w", err)
	}

	fmt.Println("[3/4] 命令幂等字段")
	logCol, err := detectLogStepColumn(db)
	if err != nil {
		return err
	}
	if err := addColumnIfMissing(db, "drill_instance_step_log", "command_id",
		fmt.Sprintf("ALTER TABLE `drill_instance_step_log` ADD COLUMN `command_id` BIGINT UNSIGNED DEFAULT NULL AFTER `%s`", logCol)); err != nil {
		return err
	}
	if err := addIndexIfMissing(db, "drill_instance_step_log", "uk_log_command_action_step",
		fmt.Sprintf("ALTER TABLE `drill_instance_step_log` ADD UNIQUE KEY `uk_log_command_action_step` (`command_id`, `action`, `%s`)", logCol)); err != nil {
		return err
	}
	if err := addColumnIfMissing(db, "notification", "command_id",
		"ALTER TABLE `notification` ADD COLUMN `command_id` BIGINT UNSIGNED DEFAULT NULL AFTER `type`"); err != nil {
		return err
	}
	if err := addIndexIfMissing(db, "notification", "uk_notification_command_user_type_step",
		"ALTER TABLE `notification` ADD UNIQUE KEY `uk_notification_command_user_type_step` (`command_id`, `user_id`, `type`, `step_id`)"); err != nil {
		return err
	}

	fmt.Println("[4/4] 校验 schema")
	return verifySchema(db)
}

func tableExists(db *sql.DB, table string) (bool, error) {
	var n int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?
	`, table).Scan(&n)
	return n > 0, err
}

func columnExists(db *sql.DB, table, column string) (bool, error) {
	var n int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?
	`, table, column).Scan(&n)
	return n > 0, err
}

func indexExists(db *sql.DB, table, index string) (bool, error) {
	var n int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND INDEX_NAME = ?
	`, table, index).Scan(&n)
	return n > 0, err
}

func addColumnIfMissing(db *sql.DB, table, column, ddl string) error {
	ok, err := columnExists(db, table, column)
	if err != nil {
		return fmt.Errorf("检查 %s.%s 失败: %w", table, column, err)
	}
	if ok {
		fmt.Printf("  ✓ 跳过 %s.%s\n", table, column)
		return nil
	}
	if _, err := db.Exec(ddl); err != nil {
		return fmt.Errorf("执行失败 %s: %w", ddl, err)
	}
	fmt.Printf("  ✓ 添加 %s.%s\n", table, column)
	return nil
}

func addIndexIfMissing(db *sql.DB, table, index, ddl string) error {
	ok, err := indexExists(db, table, index)
	if err != nil {
		return fmt.Errorf("检查 %s.%s 索引失败: %w", table, index, err)
	}
	if ok {
		fmt.Printf("  ✓ 跳过 %s.%s 索引\n", table, index)
		return nil
	}
	if _, err := db.Exec(ddl); err != nil {
		return fmt.Errorf("执行失败 %s: %w", ddl, err)
	}
	fmt.Printf("  ✓ 添加 %s.%s 索引\n", table, index)
	return nil
}

func renameColumnIfNeeded(db *sql.DB, table, oldCol, newCol, ddl string) error {
	hasOld, err := columnExists(db, table, oldCol)
	if err != nil {
		return err
	}
	hasNew, err := columnExists(db, table, newCol)
	if err != nil {
		return err
	}
	if !hasOld || hasNew {
		fmt.Printf("  ✓ 跳过 %s.%s -> %s\n", table, oldCol, newCol)
		return nil
	}
	if _, err := db.Exec(ddl); err != nil {
		return fmt.Errorf("重命名失败 %s: %w", ddl, err)
	}
	fmt.Printf("  ✓ 重命名 %s.%s -> %s\n", table, oldCol, newCol)
	return nil
}

func detectLogStepColumn(db *sql.DB) (string, error) {
	hasTask, err := columnExists(db, "drill_instance_step_log", "task_instance_id")
	if err != nil {
		return "", err
	}
	return logStepColumn(hasTask), nil
}

func logStepColumn(hasTaskInstanceID bool) string {
	if hasTaskInstanceID {
		return "task_instance_id"
	}
	return "step_instance_id"
}

func verifyRequirements() map[string]struct{} {
	keys := []string{
		"table:drill_flow_command",
		"table:drill_worker_epoch",
		"column:drill_flow_command.worker_epoch",
		"column:drill_flow_command.lease_token",
		"column:drill_flow_command.attempt_count",
		"column:drill_instance_step_log.command_id",
		"column:notification.command_id",
	}
	out := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		out[key] = struct{}{}
	}
	return out
}

func verifySchema(db *sql.DB) error {
	missing := make([]string, 0)
	for key := range verifyRequirements() {
		kind, name, _ := strings.Cut(key, ":")
		switch kind {
		case "table":
			ok, err := tableExists(db, name)
			if err != nil {
				return err
			}
			if !ok {
				missing = append(missing, key)
			}
		case "column":
			table, column, _ := strings.Cut(name, ".")
			ok, err := columnExists(db, table, column)
			if err != nil {
				return err
			}
			if !ok {
				missing = append(missing, key)
			}
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("schema 校验失败，缺少: %s", strings.Join(missing, ", "))
	}
	fmt.Println("  ✓ 多活 schema 校验通过")
	return nil
}

const createFlowCommandTableSQL = `
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='流程命令表'`

const createWorkerEpochTableSQL = `
CREATE TABLE IF NOT EXISTS ` + "`drill_worker_epoch`" + ` (
    ` + "`id`" + ` BIGINT UNSIGNED NOT NULL,
    ` + "`worker_id`" + ` VARCHAR(128) NOT NULL COMMENT '当前持有 epoch 的 worker',
    ` + "`epoch`" + ` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '单调递增 epoch，每次领导切换 +1',
    ` + "`lease_until`" + ` DATETIME DEFAULT NULL COMMENT 'epoch 租约到期时间',
    ` + "`created_at`" + ` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ` + "`updated_at`" + ` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (` + "`id`" + `)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Worker epoch 单例行表'`
