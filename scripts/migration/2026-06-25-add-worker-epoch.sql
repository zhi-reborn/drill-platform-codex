-- Migration: add worker epoch singleton table and fence command ownership.
-- Adds worker_epoch, lease_token, attempt_count columns to drill_flow_command
-- and creates the drill_worker_epoch singleton table used to fence command
-- ownership across leadership transitions.

SET @schema := DATABASE();

-- Ensure drill_flow_command exists for fresh databases that run migrations
-- before init-db.sql. The CREATE TABLE matches the canonical definition in
-- init-db.sql; existing tables are left untouched.
CREATE TABLE IF NOT EXISTS `drill_flow_command` (
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
    `lease_token` VARCHAR(128) NOT NULL DEFAULT '',
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='流程命令表';

-- Add worker_epoch column if missing.
SET @sql := (
    SELECT IF(COUNT(*) = 0,
        'ALTER TABLE `drill_flow_command` ADD COLUMN `worker_epoch` BIGINT UNSIGNED NOT NULL DEFAULT 0',
        'SELECT 1'
    )
    FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = @schema
      AND TABLE_NAME = 'drill_flow_command'
      AND COLUMN_NAME = 'worker_epoch'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add lease_token column if missing.
SET @sql := (
    SELECT IF(COUNT(*) = 0,
        'ALTER TABLE `drill_flow_command` ADD COLUMN `lease_token` VARCHAR(128) NOT NULL DEFAULT ''''',
        'SELECT 1'
    )
    FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = @schema
      AND TABLE_NAME = 'drill_flow_command'
      AND COLUMN_NAME = 'lease_token'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add attempt_count column if missing.
SET @sql := (
    SELECT IF(COUNT(*) = 0,
        'ALTER TABLE `drill_flow_command` ADD COLUMN `attempt_count` INT NOT NULL DEFAULT 0',
        'SELECT 1'
    )
    FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = @schema
      AND TABLE_NAME = 'drill_flow_command'
      AND COLUMN_NAME = 'attempt_count'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Create the singleton worker epoch table.
CREATE TABLE IF NOT EXISTS `drill_worker_epoch` (
    `id` BIGINT UNSIGNED NOT NULL,
    `worker_id` VARCHAR(128) NOT NULL COMMENT '当前持有 epoch 的 worker',
    `epoch` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '单调递增 epoch，每次领导切换 +1',
    `lease_until` DATETIME DEFAULT NULL COMMENT 'epoch 租约到期时间',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Worker epoch 单例行表';
