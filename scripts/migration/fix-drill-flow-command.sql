-- 快速修复脚本：在 drill_platform 数据库中创建 drill_flow_command 表
-- 使用方法: mysql -h <host> -P <port> -u <user> -p<pass> drill_platform < fix-drill-flow-command.sql

USE drill_platform;

-- 检查表是否已存在
SET @table_exists := (
    SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES 
    WHERE TABLE_SCHEMA = 'drill_platform' 
      AND TABLE_NAME = 'drill_flow_command'
);

-- 如果表不存在则创建
SET @sql := IF(@table_exists = 0,
    'CREATE TABLE `drill_flow_command` (
        `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
        `command_type` VARCHAR(64) NOT NULL,
        `drill_instance_id` BIGINT UNSIGNED NOT NULL,
        `step_instance_id` BIGINT UNSIGNED DEFAULT NULL,
        `operator_id` BIGINT UNSIGNED NOT NULL,
        `idempotency_key` VARCHAR(128) NOT NULL,
        `payload` JSON NOT NULL,
        `status` VARCHAR(20) NOT NULL,
        `worker_id` VARCHAR(128) DEFAULT NULL,
        `worker_epoch` BIGINT UNSIGNED NOT NULL DEFAULT 0,
        `lease_token` VARCHAR(128) NOT NULL DEFAULT \'\',
        `lease_until` DATETIME DEFAULT NULL,
        `attempts` INT NOT NULL DEFAULT 0,
        `attempt_count` INT NOT NULL DEFAULT 0,
        `result` JSON DEFAULT NULL,
        `error_code` VARCHAR(64) DEFAULT NULL,
        `error_message` VARCHAR(500) DEFAULT NULL,
        `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
        `started_at` DATETIME DEFAULT NULL,
        `finished_at` DATETIME DEFAULT NULL,
        `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
        PRIMARY KEY (`id`),
        UNIQUE KEY `uk_flow_command_idempotency` (`idempotency_key`),
        KEY `idx_flow_command_pending` (`status`, `created_at`),
        KEY `idx_flow_command_lease` (`lease_until`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT=\'流程命令表\'',
    'SELECT \'drill_flow_command 表已存在，跳过创建\' AS message'
);

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 验证表是否存在
SELECT 
    CASE 
        WHEN COUNT(*) = 1 THEN '✓ drill_flow_command 表创建成功'
        ELSE '✗ drill_flow_command 表创建失败'
    END AS result
FROM INFORMATION_SCHEMA.TABLES 
WHERE TABLE_SCHEMA = 'drill_platform' 
  AND TABLE_NAME = 'drill_flow_command';