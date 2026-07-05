-- =============================================================================
-- TiDB-compatible migration from 960a06859b2a8115be6e819614c49c4f90e4f96a
-- to the multi-active schema.
--
-- Use this script when the target database is TiDB, because TiDB does not
-- support the stored procedures used by 2026-07-05-migrate-to-multi-active.sql.
--
-- Usage:
--   mysql -h <host> -P <port> -u <user> -p<pass> drill_platform \
--     < scripts/migration/2026-07-05-migrate-to-multi-active-tidb.sql
--
-- The script is idempotent. Existing columns, tables, and indexes are skipped.
-- =============================================================================

SET NAMES utf8mb4;
SET @schema := DATABASE();

-- -----------------------------------------------------------------------------
-- 1. Business column alignment.
-- -----------------------------------------------------------------------------

-- drill_instance.current_step_id -> current_task_id
SET @sql := (
    SELECT IF(
        EXISTS (
            SELECT 1 FROM INFORMATION_SCHEMA.COLUMNS
            WHERE TABLE_SCHEMA = @schema
              AND TABLE_NAME = 'drill_instance'
              AND COLUMN_NAME = 'current_step_id'
        )
        AND NOT EXISTS (
            SELECT 1 FROM INFORMATION_SCHEMA.COLUMNS
            WHERE TABLE_SCHEMA = @schema
              AND TABLE_NAME = 'drill_instance'
              AND COLUMN_NAME = 'current_task_id'
        ),
        'ALTER TABLE `drill_instance` CHANGE COLUMN `current_step_id` `current_task_id` BIGINT UNSIGNED DEFAULT NULL COMMENT ''当前激活步骤 ID''',
        'SELECT ''skip drill_instance.current_task_id'' AS message'
    )
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- drill_instance_step.step_template_id -> template_step_id
SET @sql := (
    SELECT IF(
        EXISTS (
            SELECT 1 FROM INFORMATION_SCHEMA.COLUMNS
            WHERE TABLE_SCHEMA = @schema
              AND TABLE_NAME = 'drill_instance_step'
              AND COLUMN_NAME = 'step_template_id'
        )
        AND NOT EXISTS (
            SELECT 1 FROM INFORMATION_SCHEMA.COLUMNS
            WHERE TABLE_SCHEMA = @schema
              AND TABLE_NAME = 'drill_instance_step'
              AND COLUMN_NAME = 'template_step_id'
        ),
        'ALTER TABLE `drill_instance_step` CHANGE COLUMN `step_template_id` `template_step_id` BIGINT UNSIGNED NOT NULL COMMENT ''来源步骤模板 ID''',
        'SELECT ''skip drill_instance_step.template_step_id'' AS message'
    )
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql := (
    SELECT IF(
        NOT EXISTS (
            SELECT 1 FROM INFORMATION_SCHEMA.STATISTICS
            WHERE TABLE_SCHEMA = @schema
              AND TABLE_NAME = 'drill_instance_step'
              AND INDEX_NAME = 'idx_step_template_id'
        ),
        'ALTER TABLE `drill_instance_step` ADD KEY `idx_step_template_id` (`template_step_id`)',
        'SELECT ''skip drill_instance_step.idx_step_template_id'' AS message'
    )
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- drill_instance_step.attributes -> action_params
SET @sql := (
    SELECT IF(
        EXISTS (
            SELECT 1 FROM INFORMATION_SCHEMA.COLUMNS
            WHERE TABLE_SCHEMA = @schema
              AND TABLE_NAME = 'drill_instance_step'
              AND COLUMN_NAME = 'attributes'
        )
        AND NOT EXISTS (
            SELECT 1 FROM INFORMATION_SCHEMA.COLUMNS
            WHERE TABLE_SCHEMA = @schema
              AND TABLE_NAME = 'drill_instance_step'
              AND COLUMN_NAME = 'action_params'
        ),
        'ALTER TABLE `drill_instance_step` CHANGE COLUMN `attributes` `action_params` JSON DEFAULT NULL COMMENT ''动态扩展属性''',
        'SELECT ''skip drill_instance_step.action_params'' AS message'
    )
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- drill_instance_step_log.step_instance_id -> task_instance_id
SET @sql := (
    SELECT IF(
        EXISTS (
            SELECT 1 FROM INFORMATION_SCHEMA.COLUMNS
            WHERE TABLE_SCHEMA = @schema
              AND TABLE_NAME = 'drill_instance_step_log'
              AND COLUMN_NAME = 'step_instance_id'
        )
        AND NOT EXISTS (
            SELECT 1 FROM INFORMATION_SCHEMA.COLUMNS
            WHERE TABLE_SCHEMA = @schema
              AND TABLE_NAME = 'drill_instance_step_log'
              AND COLUMN_NAME = 'task_instance_id'
        ),
        'ALTER TABLE `drill_instance_step_log` CHANGE COLUMN `step_instance_id` `task_instance_id` BIGINT UNSIGNED DEFAULT NULL',
        'SELECT ''skip drill_instance_step_log.task_instance_id'' AS message'
    )
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- drill_template.phase_order
SET @sql := (
    SELECT IF(
        NOT EXISTS (
            SELECT 1 FROM INFORMATION_SCHEMA.COLUMNS
            WHERE TABLE_SCHEMA = @schema
              AND TABLE_NAME = 'drill_template'
              AND COLUMN_NAME = 'phase_order'
        ),
        'ALTER TABLE `drill_template` ADD COLUMN `phase_order` JSON DEFAULT NULL COMMENT ''阶段顺序列表'' AFTER `created_by`',
        'SELECT ''skip drill_template.phase_order'' AS message'
    )
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Supplemental indexes aligned with current init-db.sql.
SET @sql := (
    SELECT IF(
        NOT EXISTS (
            SELECT 1 FROM INFORMATION_SCHEMA.STATISTICS
            WHERE TABLE_SCHEMA = @schema
              AND TABLE_NAME = 'user'
              AND INDEX_NAME = 'idx_role'
        ),
        'ALTER TABLE `user` ADD KEY `idx_role` (`role`)',
        'SELECT ''skip user.idx_role'' AS message'
    )
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql := (
    SELECT IF(
        NOT EXISTS (
            SELECT 1 FROM INFORMATION_SCHEMA.STATISTICS
            WHERE TABLE_SCHEMA = @schema
              AND TABLE_NAME = 'user'
              AND INDEX_NAME = 'idx_department'
        ),
        'ALTER TABLE `user` ADD KEY `idx_department` (`department`)',
        'SELECT ''skip user.idx_department'' AS message'
    )
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql := (
    SELECT IF(
        NOT EXISTS (
            SELECT 1 FROM INFORMATION_SCHEMA.STATISTICS
            WHERE TABLE_SCHEMA = @schema
              AND TABLE_NAME = 'drill_template'
              AND INDEX_NAME = 'idx_created_by'
        ),
        'ALTER TABLE `drill_template` ADD KEY `idx_created_by` (`created_by`)',
        'SELECT ''skip drill_template.idx_created_by'' AS message'
    )
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- -----------------------------------------------------------------------------
-- 2. Multi-active command tables.
-- -----------------------------------------------------------------------------

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

SET @sql := (
    SELECT IF(
        NOT EXISTS (
            SELECT 1 FROM INFORMATION_SCHEMA.COLUMNS
            WHERE TABLE_SCHEMA = @schema
              AND TABLE_NAME = 'drill_flow_command'
              AND COLUMN_NAME = 'worker_epoch'
        ),
        'ALTER TABLE `drill_flow_command` ADD COLUMN `worker_epoch` BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `worker_id`',
        'SELECT ''skip drill_flow_command.worker_epoch'' AS message'
    )
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql := (
    SELECT IF(
        NOT EXISTS (
            SELECT 1 FROM INFORMATION_SCHEMA.COLUMNS
            WHERE TABLE_SCHEMA = @schema
              AND TABLE_NAME = 'drill_flow_command'
              AND COLUMN_NAME = 'lease_token'
        ),
        'ALTER TABLE `drill_flow_command` ADD COLUMN `lease_token` VARCHAR(128) NOT NULL DEFAULT '''' AFTER `worker_epoch`',
        'SELECT ''skip drill_flow_command.lease_token'' AS message'
    )
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql := (
    SELECT IF(
        NOT EXISTS (
            SELECT 1 FROM INFORMATION_SCHEMA.COLUMNS
            WHERE TABLE_SCHEMA = @schema
              AND TABLE_NAME = 'drill_flow_command'
              AND COLUMN_NAME = 'attempt_count'
        ),
        'ALTER TABLE `drill_flow_command` ADD COLUMN `attempt_count` INT NOT NULL DEFAULT 0 AFTER `attempts`',
        'SELECT ''skip drill_flow_command.attempt_count'' AS message'
    )
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

CREATE TABLE IF NOT EXISTS `drill_worker_epoch` (
    `id` BIGINT UNSIGNED NOT NULL,
    `worker_id` VARCHAR(128) NOT NULL COMMENT '当前持有 epoch 的 worker',
    `epoch` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '单调递增 epoch，每次领导切换 +1',
    `lease_until` DATETIME DEFAULT NULL COMMENT 'epoch 租约到期时间',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Worker epoch 单例行表';

-- -----------------------------------------------------------------------------
-- 3. Command idempotency columns and indexes.
-- -----------------------------------------------------------------------------

SET @log_command_after := (
    SELECT CASE
        WHEN EXISTS (
            SELECT 1 FROM INFORMATION_SCHEMA.COLUMNS
            WHERE TABLE_SCHEMA = @schema
              AND TABLE_NAME = 'drill_instance_step_log'
              AND COLUMN_NAME = 'task_instance_id'
        ) THEN 'task_instance_id'
        ELSE 'step_instance_id'
    END
);

SET @sql := (
    SELECT IF(
        NOT EXISTS (
            SELECT 1 FROM INFORMATION_SCHEMA.COLUMNS
            WHERE TABLE_SCHEMA = @schema
              AND TABLE_NAME = 'drill_instance_step_log'
              AND COLUMN_NAME = 'command_id'
        ),
        CONCAT('ALTER TABLE `drill_instance_step_log` ADD COLUMN `command_id` BIGINT UNSIGNED DEFAULT NULL AFTER `', @log_command_after, '`'),
        'SELECT ''skip drill_instance_step_log.command_id'' AS message'
    )
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql := (
    SELECT IF(
        NOT EXISTS (
            SELECT 1 FROM INFORMATION_SCHEMA.STATISTICS
            WHERE TABLE_SCHEMA = @schema
              AND TABLE_NAME = 'drill_instance_step_log'
              AND INDEX_NAME = 'uk_log_command_action_step'
        ),
        CONCAT('ALTER TABLE `drill_instance_step_log` ADD UNIQUE KEY `uk_log_command_action_step` (`command_id`, `action`, `', @log_command_after, '`)'),
        'SELECT ''skip drill_instance_step_log.uk_log_command_action_step'' AS message'
    )
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql := (
    SELECT IF(
        NOT EXISTS (
            SELECT 1 FROM INFORMATION_SCHEMA.COLUMNS
            WHERE TABLE_SCHEMA = @schema
              AND TABLE_NAME = 'notification'
              AND COLUMN_NAME = 'command_id'
        ),
        'ALTER TABLE `notification` ADD COLUMN `command_id` BIGINT UNSIGNED DEFAULT NULL AFTER `type`',
        'SELECT ''skip notification.command_id'' AS message'
    )
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql := (
    SELECT IF(
        NOT EXISTS (
            SELECT 1 FROM INFORMATION_SCHEMA.STATISTICS
            WHERE TABLE_SCHEMA = @schema
              AND TABLE_NAME = 'notification'
              AND INDEX_NAME = 'uk_notification_command_user_type_step'
        ),
        'ALTER TABLE `notification` ADD UNIQUE KEY `uk_notification_command_user_type_step` (`command_id`, `user_id`, `type`, `step_id`)',
        'SELECT ''skip notification.uk_notification_command_user_type_step'' AS message'
    )
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- -----------------------------------------------------------------------------
-- 4. Verification.
-- -----------------------------------------------------------------------------

SELECT
    SUM(TABLE_NAME = 'drill_flow_command') AS drill_flow_command_exists,
    SUM(TABLE_NAME = 'drill_worker_epoch') AS drill_worker_epoch_exists
FROM INFORMATION_SCHEMA.TABLES
WHERE TABLE_SCHEMA = @schema
  AND TABLE_NAME IN ('drill_flow_command', 'drill_worker_epoch');

SELECT TABLE_NAME, COLUMN_NAME
FROM INFORMATION_SCHEMA.COLUMNS
WHERE TABLE_SCHEMA = @schema
  AND (
      (TABLE_NAME = 'drill_flow_command' AND COLUMN_NAME IN ('worker_epoch', 'lease_token', 'attempt_count'))
      OR (TABLE_NAME = 'drill_instance_step_log' AND COLUMN_NAME = 'command_id')
      OR (TABLE_NAME = 'notification' AND COLUMN_NAME = 'command_id')
  )
ORDER BY TABLE_NAME, COLUMN_NAME;
