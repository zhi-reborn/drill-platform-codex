-- =============================================================================
-- 从 960a06859b2a8115be6e819614c49c4f90e4f96a 升级到多活部署所需的全部表结构变更
--
-- 用法:
--   mysql -h <host> -P <port> -u <user> -p<pass> drill_platform < 2026-07-05-migrate-to-multi-active.sql
--
-- 特性:
--   1. 完全幂等：可重复执行，已存在的对象/列/索引会自动跳过
--   2. 自适应：兼容列已重命名/未重命名两种状态
--   3. 包含两类变更:
--      A) 多活必需：drill_flow_command / drill_worker_epoch 表 + command_id + 幂等唯一键
--      B) 业务对齐：字段重命名 / 索引调整 / phase_order 字段（当前代码已使用新字段名，
--         必须迁移，否则应用启动查询会失败）
--
-- 前置条件:
--   - 数据库已存在 (drill_platform)
--   - schema 处于 960a068 commit 后的状态（drill_template_step 已是 node-tree 模型）
--   - MySQL 8.0+ 推荐（CHANGE COLUMN 重命名，5.7+ 均支持）
-- =============================================================================

SET NAMES utf8mb4;
SET @schema := DATABASE();

-- =============================================================================
-- 1. 集中定义所有辅助存储过程
-- =============================================================================
DELIMITER //

DROP PROCEDURE IF EXISTS exec_sql_if_exists //
CREATE PROCEDURE exec_sql_if_exists(IN p_table VARCHAR(64), IN p_column VARCHAR(64), IN p_ddl TEXT)
BEGIN
    DECLARE col_exists INT DEFAULT 0;

    SELECT COUNT(*) INTO col_exists
    FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE()
      AND TABLE_NAME = p_table
      AND COLUMN_NAME = p_column;

    IF col_exists = 0 THEN
        SET @sql := p_ddl;
        PREPARE stmt FROM @sql;
        EXECUTE stmt;
        DEALLOCATE PREPARE stmt;
    END IF;
END //

DROP PROCEDURE IF EXISTS exec_sql_if_exists_idx //
CREATE PROCEDURE exec_sql_if_exists_idx(IN p_table VARCHAR(64), IN p_index VARCHAR(128), IN p_ddl TEXT)
BEGIN
    DECLARE idx_exists INT DEFAULT 0;

    SELECT COUNT(*) INTO idx_exists
    FROM INFORMATION_SCHEMA.STATISTICS
    WHERE TABLE_SCHEMA = DATABASE()
      AND TABLE_NAME = p_table
      AND INDEX_NAME = p_index;

    IF idx_exists = 0 THEN
        SET @sql := p_ddl;
        PREPARE stmt FROM @sql;
        EXECUTE stmt;
        DEALLOCATE PREPARE stmt;
    END IF;
END //

DROP PROCEDURE IF EXISTS rename_column_if_needed //
-- 通用列重命名：仅当 old 列存在且 new 列不存在时执行
CREATE PROCEDURE rename_column_if_needed(IN p_table VARCHAR(64), IN p_old VARCHAR(64), IN p_new VARCHAR(64), IN p_ddl TEXT)
BEGIN
    DECLARE has_old INT DEFAULT 0;
    DECLARE has_new INT DEFAULT 0;

    SELECT COUNT(*) INTO has_old
    FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = p_table AND COLUMN_NAME = p_old;

    SELECT COUNT(*) INTO has_new
    FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = p_table AND COLUMN_NAME = p_new;

    IF has_old = 1 AND has_new = 0 THEN
        SET @sql := p_ddl;
        PREPARE stmt FROM @sql;
        EXECUTE stmt;
        DEALLOCATE PREPARE stmt;
    END IF;
END //

DROP PROCEDURE IF EXISTS add_log_unique_key //
-- drill_instance_step_log 添加 uk_log_command_action_step，自动适配列名
CREATE PROCEDURE add_log_unique_key()
BEGIN
    DECLARE idx_exists INT DEFAULT 0;
    DECLARE step_col VARCHAR(64);

    SELECT COUNT(*) INTO idx_exists
    FROM INFORMATION_SCHEMA.STATISTICS
    WHERE TABLE_SCHEMA = DATABASE()
      AND TABLE_NAME = 'drill_instance_step_log'
      AND INDEX_NAME = 'uk_log_command_action_step';

    IF idx_exists = 0 THEN
        SELECT CASE
            WHEN EXISTS (
                SELECT 1 FROM INFORMATION_SCHEMA.COLUMNS
                WHERE TABLE_SCHEMA = DATABASE()
                  AND TABLE_NAME = 'drill_instance_step_log'
                  AND COLUMN_NAME = 'task_instance_id'
            ) THEN 'task_instance_id'
            ELSE 'step_instance_id'
        END INTO step_col;

        SET @ddl := CONCAT(
            'ALTER TABLE `drill_instance_step_log` ADD UNIQUE KEY `uk_log_command_action_step` (`command_id`, `action`, `', step_col, '`)'
        );
        PREPARE stmt FROM @ddl;
        EXECUTE stmt;
        DEALLOCATE PREPARE stmt;
    END IF;
END //

DROP PROCEDURE IF EXISTS verify_ha_schema //
CREATE PROCEDURE verify_ha_schema()
BEGIN
    DECLARE v INT DEFAULT 0;

    SELECT COUNT(*) INTO v FROM INFORMATION_SCHEMA.TABLES
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'drill_flow_command';
    IF v = 0 THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = '校验失败: drill_flow_command 表未创建';
    END IF;

    SELECT COUNT(*) INTO v FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'drill_flow_command'
      AND COLUMN_NAME IN ('worker_epoch', 'lease_token', 'attempt_count');
    IF v < 3 THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = '校验失败: drill_flow_command 缺少 worker_epoch/lease_token/attempt_count 列';
    END IF;

    SELECT COUNT(*) INTO v FROM INFORMATION_SCHEMA.TABLES
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'drill_worker_epoch';
    IF v = 0 THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = '校验失败: drill_worker_epoch 表未创建';
    END IF;

    SELECT COUNT(*) INTO v FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'drill_instance_step_log' AND COLUMN_NAME = 'command_id';
    IF v = 0 THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = '校验失败: drill_instance_step_log.command_id 列未添加';
    END IF;

    SELECT COUNT(*) INTO v FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'notification' AND COLUMN_NAME = 'command_id';
    IF v = 0 THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = '校验失败: notification.command_id 列未添加';
    END IF;

    SELECT '多活表结构校验通过' AS result;
END //

DELIMITER ;

-- =============================================================================
-- 2. 业务字段重命名（当前代码已使用新字段名，必须迁移）
-- =============================================================================

-- B1. drill_instance.current_step_id → current_task_id
CALL rename_column_if_needed(
    'drill_instance', 'current_step_id', 'current_task_id',
    'ALTER TABLE `drill_instance` CHANGE COLUMN `current_step_id` `current_task_id` BIGINT UNSIGNED DEFAULT NULL COMMENT \'当前激活步骤 ID\''
);

-- B2. drill_instance_step.step_template_id → template_step_id
CALL rename_column_if_needed(
    'drill_instance_step', 'step_template_id', 'template_step_id',
    'ALTER TABLE `drill_instance_step` CHANGE COLUMN `step_template_id` `template_step_id` BIGINT UNSIGNED NOT NULL COMMENT \'来源步骤模板 ID\''
);
CALL exec_sql_if_exists_idx('drill_instance_step', 'idx_step_template_id',
    'ALTER TABLE `drill_instance_step` ADD KEY `idx_step_template_id` (`template_step_id`)');

-- B3. drill_instance_step.attributes → action_params
--     （drill_template_step 表的 attributes 仍保留，不重命名）
CALL rename_column_if_needed(
    'drill_instance_step', 'attributes', 'action_params',
    'ALTER TABLE `drill_instance_step` CHANGE COLUMN `attributes` `action_params` JSON DEFAULT NULL COMMENT \'动态扩展属性\''
);

-- B4. drill_instance_step_log.step_instance_id → task_instance_id
CALL rename_column_if_needed(
    'drill_instance_step_log', 'step_instance_id', 'task_instance_id',
    'ALTER TABLE `drill_instance_step_log` CHANGE COLUMN `step_instance_id` `task_instance_id` BIGINT UNSIGNED DEFAULT NULL'
);

-- B5. drill_template.phase_order
CALL exec_sql_if_exists('drill_template', 'phase_order',
    'ALTER TABLE `drill_template` ADD COLUMN `phase_order` JSON DEFAULT NULL COMMENT \'阶段顺序列表\' AFTER `created_by`');

-- B6. 索引补齐（与最新 init-db.sql 对齐）
CALL exec_sql_if_exists_idx('user', 'idx_role',
    'ALTER TABLE `user` ADD KEY `idx_role` (`role`)');
CALL exec_sql_if_exists_idx('user', 'idx_department',
    'ALTER TABLE `user` ADD KEY `idx_department` (`department`)');
CALL exec_sql_if_exists_idx('drill_template', 'idx_created_by',
    'ALTER TABLE `drill_template` ADD KEY `idx_created_by` (`created_by`)');

-- =============================================================================
-- 3. 【多活核心】新增表 drill_flow_command + drill_worker_epoch
--    CREATE TABLE IF NOT EXISTS 保证幂等；后续 ALTER 补齐列兼容老版本
-- =============================================================================

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

-- 兼容已存在旧表（缺 worker_epoch/lease_token/attempt_count 列）
CALL exec_sql_if_exists('drill_flow_command', 'worker_epoch',
    'ALTER TABLE `drill_flow_command` ADD COLUMN `worker_epoch` BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `worker_id`');
CALL exec_sql_if_exists('drill_flow_command', 'lease_token',
    'ALTER TABLE `drill_flow_command` ADD COLUMN `lease_token` VARCHAR(128) NOT NULL DEFAULT \'\' AFTER `worker_epoch`');
CALL exec_sql_if_exists('drill_flow_command', 'attempt_count',
    'ALTER TABLE `drill_flow_command` ADD COLUMN `attempt_count` INT NOT NULL DEFAULT 0 AFTER `attempts`');

CREATE TABLE IF NOT EXISTS `drill_worker_epoch` (
    `id` BIGINT UNSIGNED NOT NULL,
    `worker_id` VARCHAR(128) NOT NULL COMMENT '当前持有 epoch 的 worker',
    `epoch` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '单调递增 epoch，每次领导切换 +1',
    `lease_until` DATETIME DEFAULT NULL COMMENT 'epoch 租约到期时间',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Worker epoch 单例行表';

-- =============================================================================
-- 4. 【多活核心】command_id 字段 + 幂等唯一键
-- =============================================================================

-- drill_instance_step_log.command_id + uk_log_command_action_step
CALL exec_sql_if_exists('drill_instance_step_log', 'command_id',
    'ALTER TABLE `drill_instance_step_log` ADD COLUMN `command_id` BIGINT UNSIGNED DEFAULT NULL AFTER `task_instance_id`');
CALL add_log_unique_key();

-- notification.command_id + uk_notification_command_user_type_step
CALL exec_sql_if_exists('notification', 'command_id',
    'ALTER TABLE `notification` ADD COLUMN `command_id` BIGINT UNSIGNED DEFAULT NULL AFTER `type`');
CALL exec_sql_if_exists_idx('notification', 'uk_notification_command_user_type_step',
    'ALTER TABLE `notification` ADD UNIQUE KEY `uk_notification_command_user_type_step` (`command_id`, `user_id`, `type`, `step_id`)');

-- =============================================================================
-- 5. 校验：确认所有多活核心对象已就位（不通过会中断）
-- =============================================================================
CALL verify_ha_schema();

-- =============================================================================
-- 6. 清理辅助过程
-- =============================================================================
DROP PROCEDURE IF EXISTS exec_sql_if_exists;
DROP PROCEDURE IF EXISTS exec_sql_if_exists_idx;
DROP PROCEDURE IF EXISTS rename_column_if_needed;
DROP PROCEDURE IF EXISTS add_log_unique_key;
DROP PROCEDURE IF EXISTS verify_ha_schema;

-- =============================================================================
-- 迁移完成
-- 下一步：
--   1. 部署外部 MySQL HA（MGR/RDS/ProxySQL）+ Redis Cluster
--   2. cp .env.example .env 填写 DATABASE_HOST / REDIS_CLUSTER_ADDRS / JWT_SECRET
--   3. docker compose -f docker-compose.ha.yml up -d --build
--   4. 运行 scripts/test-ha.sh 验证故障切换
-- =============================================================================
