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
    `lease_until` DATETIME(3) DEFAULT NULL,
    `attempts` INT NOT NULL DEFAULT 0,
    `result` JSON DEFAULT NULL,
    `error_code` VARCHAR(64) DEFAULT NULL,
    `error_message` VARCHAR(500) DEFAULT NULL,
    `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    `started_at` DATETIME(3) DEFAULT NULL,
    `finished_at` DATETIME(3) DEFAULT NULL,
    `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_flow_command_idempotency` (`idempotency_key`),
    KEY `idx_flow_command_pending` (`status`, `created_at`, `id`),
    KEY `idx_flow_command_lease` (`status`, `lease_until`),
    KEY `idx_flow_command_drill_status` (`drill_instance_id`, `status`, `id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='持久化流程命令表';

SET @sql = (
    SELECT IF(
        COUNT(*) = 0,
        'ALTER TABLE `drill_instance_step_log` ADD COLUMN `command_id` BIGINT UNSIGNED DEFAULT NULL AFTER `id`',
        'SELECT 1'
    )
    FROM `INFORMATION_SCHEMA`.`COLUMNS`
    WHERE `TABLE_SCHEMA` = DATABASE()
      AND `TABLE_NAME` = 'drill_instance_step_log'
      AND `COLUMN_NAME` = 'command_id'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql = (
    SELECT IF(
        COUNT(*) = 0,
        'ALTER TABLE `notification` ADD COLUMN `command_id` BIGINT UNSIGNED DEFAULT NULL AFTER `id`',
        'SELECT 1'
    )
    FROM `INFORMATION_SCHEMA`.`COLUMNS`
    WHERE `TABLE_SCHEMA` = DATABASE()
      AND `TABLE_NAME` = 'notification'
      AND `COLUMN_NAME` = 'command_id'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql = (
    SELECT IF(
        COUNT(*) = 0,
        'ALTER TABLE `drill_flow_command` ADD UNIQUE INDEX `uk_flow_command_idempotency` (`idempotency_key`)',
        'SELECT 1'
    )
    FROM `INFORMATION_SCHEMA`.`STATISTICS`
    WHERE `TABLE_SCHEMA` = DATABASE()
      AND `TABLE_NAME` = 'drill_flow_command'
      AND `INDEX_NAME` = 'uk_flow_command_idempotency'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql = (
    SELECT IF(
        COUNT(*) = 0,
        'ALTER TABLE `drill_flow_command` ADD INDEX `idx_flow_command_pending` (`status`, `created_at`, `id`)',
        'SELECT 1'
    )
    FROM `INFORMATION_SCHEMA`.`STATISTICS`
    WHERE `TABLE_SCHEMA` = DATABASE()
      AND `TABLE_NAME` = 'drill_flow_command'
      AND `INDEX_NAME` = 'idx_flow_command_pending'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql = (
    SELECT IF(
        COUNT(*) = 0,
        'ALTER TABLE `drill_flow_command` ADD INDEX `idx_flow_command_lease` (`status`, `lease_until`)',
        'SELECT 1'
    )
    FROM `INFORMATION_SCHEMA`.`STATISTICS`
    WHERE `TABLE_SCHEMA` = DATABASE()
      AND `TABLE_NAME` = 'drill_flow_command'
      AND `INDEX_NAME` = 'idx_flow_command_lease'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql = (
    SELECT IF(
        COUNT(*) = 0,
        'ALTER TABLE `drill_flow_command` ADD INDEX `idx_flow_command_drill_status` (`drill_instance_id`, `status`, `id`)',
        'SELECT 1'
    )
    FROM `INFORMATION_SCHEMA`.`STATISTICS`
    WHERE `TABLE_SCHEMA` = DATABASE()
      AND `TABLE_NAME` = 'drill_flow_command'
      AND `INDEX_NAME` = 'idx_flow_command_drill_status'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql = (
    SELECT IF(
        COUNT(*) = 0,
        'ALTER TABLE `drill_instance_step_log` ADD UNIQUE INDEX `uk_log_command_action_task` (`command_id`, `action`, `task_instance_id`)',
        'SELECT 1'
    )
    FROM `INFORMATION_SCHEMA`.`STATISTICS`
    WHERE `TABLE_SCHEMA` = DATABASE()
      AND `TABLE_NAME` = 'drill_instance_step_log'
      AND `INDEX_NAME` = 'uk_log_command_action_task'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql = (
    SELECT IF(
        COUNT(*) = 0,
        'ALTER TABLE `notification` ADD UNIQUE INDEX `uk_notification_command_user_type_step` (`command_id`, `user_id`, `type`, `step_id`)',
        'SELECT 1'
    )
    FROM `INFORMATION_SCHEMA`.`STATISTICS`
    WHERE `TABLE_SCHEMA` = DATABASE()
      AND `TABLE_NAME` = 'notification'
      AND `INDEX_NAME` = 'uk_notification_command_user_type_step'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
