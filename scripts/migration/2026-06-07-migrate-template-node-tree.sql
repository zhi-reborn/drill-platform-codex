-- Migrate the legacy phase/stage/task/step template schema into the current
-- drill_template_step tree model.
--
-- Run this only when drill_template_step still uses template_task_id.
-- Node ID mapping:
--   phase     -> original id
--   stage     -> 100000 + original id
--   task      -> 200000 + original id
--   operation -> 300000 + original id

SET FOREIGN_KEY_CHECKS = 0;

DROP TABLE IF EXISTS `drill_template_step_old_20260607`;
RENAME TABLE `drill_template_step` TO `drill_template_step_old_20260607`;

CREATE TABLE `drill_template_step` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `drill_template_id` BIGINT UNSIGNED NOT NULL COMMENT '所属模板 ID',
    `parent_step_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '父步骤 ID，NULL 表示根节点',
    `name` VARCHAR(128) NOT NULL COMMENT '步骤名称',
    `seq` INT NOT NULL COMMENT '排序序号',
    `step_type` VARCHAR(32) NOT NULL COMMENT '类型：serial/parallel',
    `timeout_minutes` INT NOT NULL DEFAULT 5 COMMENT '超时时间 (分钟)',
    `pre_step_ids` JSON DEFAULT NULL COMMENT '前置步骤 ID 列表',
    `guide_content` TEXT COMMENT '操作指引',
    `is_blocking` TINYINT NOT NULL DEFAULT 1 COMMENT '是否阻塞：0-非阻塞，1-阻塞',
    `default_assignee_role` VARCHAR(64) DEFAULT NULL COMMENT '默认分配角色',
    `executor_team` VARCHAR(64) DEFAULT NULL COMMENT '执行团队',
    `phase` VARCHAR(64) DEFAULT NULL COMMENT '阶段名称',
    `phase_step` VARCHAR(64) DEFAULT NULL COMMENT '环节/子阶段',
    `execution_mode` VARCHAR(16) DEFAULT 'serial' COMMENT '执行模式：serial/parallel',
    `estimated_duration_minutes` INT DEFAULT NULL COMMENT '预计耗时(分钟)',
    `estimated_start_offset` INT DEFAULT NULL COMMENT '预计开始时间偏移(分钟)',
    `attributes` JSON DEFAULT NULL COMMENT '动态扩展属性',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_drill_template_id` (`drill_template_id`),
    KEY `idx_parent_step` (`parent_step_id`),
    CONSTRAINT `fk_step_template_drill` FOREIGN KEY (`drill_template_id`) REFERENCES `drill_template` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_step_parent_template` FOREIGN KEY (`parent_step_id`) REFERENCES `drill_template_step` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='步骤模板表';

INSERT INTO `drill_template_step` (
    `id`, `drill_template_id`, `parent_step_id`, `name`, `seq`, `step_type`,
    `timeout_minutes`, `pre_step_ids`, `guide_content`, `is_blocking`,
    `default_assignee_role`, `executor_team`, `phase`, `phase_step`,
    `execution_mode`, `estimated_duration_minutes`, `estimated_start_offset`,
    `attributes`, `created_at`
)
SELECT
    p.`id`,
    p.`drill_template_id`,
    NULL,
    p.`name`,
    p.`seq`,
    COALESCE(NULLIF(p.`step_type`, ''), 'serial'),
    5,
    JSON_ARRAY(),
    NULL,
    1,
    NULL,
    NULL,
    p.`name`,
    NULL,
    COALESCE(NULLIF(p.`step_type`, ''), 'serial'),
    NULL,
    NULL,
    JSON_OBJECT('level', 'phase', 'legacy_table', 'drill_template_phase', 'legacy_id', p.`id`),
    p.`created_at`
FROM `drill_template_phase` p;

INSERT INTO `drill_template_step` (
    `id`, `drill_template_id`, `parent_step_id`, `name`, `seq`, `step_type`,
    `timeout_minutes`, `pre_step_ids`, `guide_content`, `is_blocking`,
    `default_assignee_role`, `executor_team`, `phase`, `phase_step`,
    `execution_mode`, `estimated_duration_minutes`, `estimated_start_offset`,
    `attributes`, `created_at`
)
SELECT
    100000 + s.`id`,
    p.`drill_template_id`,
    p.`id`,
    s.`name`,
    s.`seq`,
    COALESCE(NULLIF(s.`step_type`, ''), 'serial'),
    5,
    JSON_ARRAY(),
    NULL,
    1,
    NULL,
    NULL,
    p.`name`,
    s.`name`,
    COALESCE(NULLIF(s.`step_type`, ''), 'serial'),
    NULL,
    NULL,
    JSON_OBJECT('level', 'stage', 'legacy_table', 'drill_template_stage', 'legacy_id', s.`id`),
    s.`created_at`
FROM `drill_template_stage` s
JOIN `drill_template_phase` p ON p.`id` = s.`template_phase_id`;

INSERT INTO `drill_template_step` (
    `id`, `drill_template_id`, `parent_step_id`, `name`, `seq`, `step_type`,
    `timeout_minutes`, `pre_step_ids`, `guide_content`, `is_blocking`,
    `default_assignee_role`, `executor_team`, `phase`, `phase_step`,
    `execution_mode`, `estimated_duration_minutes`, `estimated_start_offset`,
    `attributes`, `created_at`
)
SELECT
    200000 + t.`id`,
    p.`drill_template_id`,
    100000 + s.`id`,
    t.`name`,
    t.`seq`,
    COALESCE(NULLIF(t.`step_type`, ''), NULLIF(t.`execution_mode`, ''), 'serial'),
    COALESCE(t.`timeout_minutes`, 5),
    COALESCE(t.`pre_step_ids`, JSON_ARRAY()),
    t.`guide_content`,
    COALESCE(t.`is_blocking`, 1),
    t.`default_assignee_role`,
    t.`executor_team`,
    COALESCE(t.`phase`, p.`name`),
    COALESCE(t.`phase_step`, s.`name`),
    COALESCE(NULLIF(t.`execution_mode`, ''), NULLIF(t.`step_type`, ''), 'serial'),
    t.`estimated_duration_minutes`,
    t.`estimated_start_offset`,
    COALESCE(t.`attributes`, JSON_OBJECT('level', 'task', 'legacy_table', 'drill_template_task', 'legacy_id', t.`id`)),
    t.`created_at`
FROM `drill_template_task` t
JOIN `drill_template_stage` s ON s.`id` = t.`template_stage_id`
JOIN `drill_template_phase` p ON p.`id` = s.`template_phase_id`;

INSERT INTO `drill_template_step` (
    `id`, `drill_template_id`, `parent_step_id`, `name`, `seq`, `step_type`,
    `timeout_minutes`, `pre_step_ids`, `guide_content`, `is_blocking`,
    `default_assignee_role`, `executor_team`, `phase`, `phase_step`,
    `execution_mode`, `estimated_duration_minutes`, `estimated_start_offset`,
    `attributes`, `created_at`
)
SELECT
    300000 + os.`id`,
    p.`drill_template_id`,
    200000 + t.`id`,
    os.`name`,
    os.`seq`,
    COALESCE(NULLIF(os.`step_type`, ''), 'serial'),
    5,
    JSON_ARRAY(),
    NULL,
    1,
    NULL,
    t.`executor_team`,
    COALESCE(t.`phase`, p.`name`),
    COALESCE(t.`phase_step`, s.`name`),
    COALESCE(NULLIF(os.`step_type`, ''), 'serial'),
    NULL,
    NULL,
    JSON_OBJECT(
        'level', 'operation',
        'legacy_table', 'drill_template_step',
        'legacy_id', os.`id`,
        'action_type', os.`action_type`,
        'action_params', COALESCE(os.`action_params`, JSON_OBJECT())
    ),
    os.`created_at`
FROM `drill_template_step_old_20260607` os
JOIN `drill_template_task` t ON t.`id` = os.`template_task_id`
JOIN `drill_template_stage` s ON s.`id` = t.`template_stage_id`
JOIN `drill_template_phase` p ON p.`id` = s.`template_phase_id`;

SET @next_id := (
    SELECT COALESCE(MAX(`id`), 0) + 1 FROM `drill_template_step`
);
SET @sql := CONCAT('ALTER TABLE `drill_template_step` AUTO_INCREMENT = ', @next_id);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET FOREIGN_KEY_CHECKS = 1;
