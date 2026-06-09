package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/yaml.v3"
)

// 用法: go run ./cmd/migrate-schema/main.go
// 从 configs/config.yaml 读取数据库连接配置

type Config struct {
	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Database string `yaml:"database"`
	} `yaml:"database"`
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
			fmt.Printf("配置文件加载成功: %s\n", path)
			break
		}
	}
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	dbCfg := cfg.Database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&multiStatements=true",
		dbCfg.User, dbCfg.Password, dbCfg.Host, dbCfg.Port, dbCfg.Database,
	)

	dbConn, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer dbConn.Close()

	if err := dbConn.Ping(); err != nil {
		log.Fatalf("数据库 ping 失败: %v", err)
	}
	fmt.Println("数据库连接成功")

	migrations := []struct {
		name string
		sql  string
	}{
		// 1. drill_template_step: 添加缺失列（兼容旧表结构）
		{
			name: "drill_template_step: 添加 phase 列",
			sql:  "ALTER TABLE `drill_template_step` ADD COLUMN `phase` VARCHAR(64) DEFAULT NULL COMMENT '阶段名称'",
		},
		{
			name: "drill_template_step: 添加 phase_step 列",
			sql:  "ALTER TABLE `drill_template_step` ADD COLUMN `phase_step` VARCHAR(64) DEFAULT NULL COMMENT '环节/子阶段'",
		},
		{
			name: "drill_template_step: 添加 execution_mode 列",
			sql:  "ALTER TABLE `drill_template_step` ADD COLUMN `execution_mode` VARCHAR(16) DEFAULT 'serial' COMMENT '执行模式'",
		},
		{
			name: "drill_template_step: 添加 estimated_duration_minutes 列",
			sql:  "ALTER TABLE `drill_template_step` ADD COLUMN `estimated_duration_minutes` INT DEFAULT NULL COMMENT '预计耗时(分钟)'",
		},
		{
			name: "drill_template_step: 添加 estimated_start_offset 列",
			sql:  "ALTER TABLE `drill_template_step` ADD COLUMN `estimated_start_offset` INT DEFAULT NULL COMMENT '预计开始时间偏移(分钟)'",
		},
		{
			name: "drill_template_step: 添加 attributes 列",
			sql:  "ALTER TABLE `drill_template_step` ADD COLUMN `attributes` JSON DEFAULT NULL COMMENT '动态扩展属性'",
		},
		{
			name: "drill_template_step: 添加 parent_step_id 列",
			sql:  "ALTER TABLE `drill_template_step` ADD COLUMN `parent_step_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '父步骤 ID'",
		},

		// 2. drill_template_step: timeout_minutes 默认值 5 → 120
		{
			name: "drill_template_step: timeout_minutes 默认值改为 120",
			sql:  "ALTER TABLE `drill_template_step` ALTER COLUMN `timeout_minutes` SET DEFAULT 120",
		},

		// 3. drill_instance_step: 列重命名
		{
			name: "drill_instance_step: step_template_id → template_step_id",
			sql:  "ALTER TABLE `drill_instance_step` CHANGE COLUMN `step_template_id` `template_step_id` BIGINT UNSIGNED NOT NULL COMMENT '来源步骤模板 ID'",
		},
		{
			name: "drill_instance_step: attributes → action_params",
			sql:  "ALTER TABLE `drill_instance_step` CHANGE COLUMN `attributes` `action_params` JSON DEFAULT NULL COMMENT '动态扩展属性'",
		},

		// 4. drill_instance_step: timeout_minutes 默认值 5 → 120
		{
			name: "drill_instance_step: timeout_minutes 默认值改为 120",
			sql:  "ALTER TABLE `drill_instance_step` ALTER COLUMN `timeout_minutes` SET DEFAULT 120",
		},

		// 5. drill_instance: 列重命名
		{
			name: "drill_instance: current_step_id → current_task_id",
			sql:  "ALTER TABLE `drill_instance` CHANGE COLUMN `current_step_id` `current_task_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '当前激活步骤 ID'",
		},

		// 6. drill_instance_step_log: 列重命名
		{
			name: "drill_instance_step_log: step_instance_id → task_instance_id",
			sql:  "ALTER TABLE `drill_instance_step_log` CHANGE COLUMN `step_instance_id` `task_instance_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '步骤实例 ID'",
		},
	}

	successCount := 0
	skipCount := 0
	failCount := 0

	for _, m := range migrations {
		fmt.Printf("\n[%s] 执行中...\n", m.name)
		_, err := dbConn.Exec(m.sql)
		if err != nil {
			errMsg := err.Error()
			// MySQL Error 1060 = Duplicate column name (列已存在)
			// MySQL Error 1054 = Unknown column (CHANGE 时旧列不存在，说明已迁移)
			// MySQL Error 1091 = Can't DROP (列不存在)
			if strings.Contains(errMsg, "Error 1060") ||
				strings.Contains(errMsg, "Duplicate column name") {
				fmt.Printf("  ✓ 跳过 (列已存在)\n")
				skipCount++
			} else if strings.Contains(errMsg, "Error 1054") && strings.Contains(m.sql, "CHANGE COLUMN") {
				fmt.Printf("  ✓ 跳过 (旧列不存在，可能已迁移)\n")
				skipCount++
			} else {
				fmt.Printf("  ✗ 失败: %v\n", err)
				failCount++
			}
		} else {
			fmt.Printf("  ✓ 成功\n")
			successCount++
		}
	}

	fmt.Printf("\n===== 迁移完成 =====\n")
	fmt.Printf("成功: %d, 跳过: %d, 失败: %d\n", successCount, skipCount, failCount)

	if failCount > 0 {
		os.Exit(1)
	}
}
