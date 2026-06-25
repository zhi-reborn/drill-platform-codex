-- 演练平台数据库初始化脚本
-- 基于最新生产环境表结构

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

CREATE DATABASE IF NOT EXISTS drill_platform DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE drill_platform;

-- 用户表
DROP TABLE IF EXISTS `user`;
CREATE TABLE `user` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `username` VARCHAR(64) NOT NULL COMMENT '用户名',
    `real_name` VARCHAR(64) NOT NULL COMMENT '真实姓名',
    `password_hash` VARCHAR(256) NOT NULL COMMENT '密码哈希 (bcrypt)',
    `email` VARCHAR(128) DEFAULT NULL COMMENT '邮箱',
    `role` VARCHAR(32) NOT NULL COMMENT '角色：admin/director/executor/viewer',
    `department` VARCHAR(64) DEFAULT NULL COMMENT '部门',
    `phone` VARCHAR(20) DEFAULT NULL COMMENT '手机号',
    `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1-启用，0-禁用',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_username` (`username`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 模板分类表
DROP TABLE IF EXISTS `drill_template_category`;
CREATE TABLE `drill_template_category` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
    `value` VARCHAR(50) NOT NULL COMMENT '分类值（英文标识）',
    `label` VARCHAR(50) NOT NULL COMMENT '分类名称（中文显示）',
    `sort_order` INT NOT NULL DEFAULT 0 COMMENT '排序顺序',
    `tag_type` VARCHAR(20) NOT NULL DEFAULT 'info' COMMENT '标签类型：primary|success|warning|danger|info',
    `created_at` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_value` (`value`),
    KEY `idx_sort` (`sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='模板分类表';

-- 演练模板表
DROP TABLE IF EXISTS `drill_template`;
CREATE TABLE `drill_template` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `name` VARCHAR(128) NOT NULL COMMENT '模板名称',
    `category` VARCHAR(64) NOT NULL COMMENT '分类：灾备/降级/发布/安全',
    `description` TEXT COMMENT '模板描述',
    `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：0-禁用，1-启用',
    `created_by` BIGINT UNSIGNED NOT NULL COMMENT '创建人 ID',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` DATETIME DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_category` (`category`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='演练模板表';

-- 步骤模板表
DROP TABLE IF EXISTS `drill_template_step`;
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

-- 演练实例表
DROP TABLE IF EXISTS `drill_instance`;
CREATE TABLE `drill_instance` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `template_id` BIGINT UNSIGNED NOT NULL COMMENT '来源模板 ID',
    `name` VARCHAR(128) NOT NULL COMMENT '演练名称',
    `description` TEXT COMMENT '注意事项/演练描述',
    `status` VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT '状态：pending/running/paused/completed/terminated',
    `start_time` DATETIME DEFAULT NULL COMMENT '实际开始时间',
    `end_time` DATETIME DEFAULT NULL COMMENT '实际结束时间',
    `planned_start` DATETIME DEFAULT NULL COMMENT '计划开始时间',
    `current_step_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '当前激活步骤 ID',
    `progress_pct` INT NOT NULL DEFAULT 0 COMMENT '进度百分比 (0-100)',
    `created_by` BIGINT UNSIGNED NOT NULL COMMENT '创建人 ID',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_status` (`status`),
    KEY `idx_created_at` (`created_at`),
    KEY `idx_template_id` (`template_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='演练实例表';

-- 步骤实例表
DROP TABLE IF EXISTS `drill_instance_step`;
CREATE TABLE `drill_instance_step` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `drill_instance_id` BIGINT UNSIGNED NOT NULL COMMENT '所属演练 ID',
    `parent_step_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '父步骤 ID，NULL 表示根节点',
    `step_template_id` BIGINT UNSIGNED NOT NULL COMMENT '来源步骤模板 ID',
    `name` VARCHAR(128) NOT NULL COMMENT '步骤名称',
    `seq` INT NOT NULL COMMENT '排序序号',
    `status` VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT '状态：pending/running/completed/timeout/skipped/issue',
    `assignee_ids` JSON NOT NULL COMMENT '分配的执行 ID 列表',
    `actual_operator` BIGINT UNSIGNED DEFAULT NULL COMMENT '实际完成人 ID',
    `start_time` DATETIME DEFAULT NULL COMMENT '步骤开始时间',
    `end_time` DATETIME DEFAULT NULL COMMENT '步骤结束时间',
    `timeout_at` DATETIME DEFAULT NULL COMMENT '超时截止时间',
    `remark` TEXT COMMENT '完成备注',
    `issue_desc` TEXT COMMENT '问题描述',
    `step_type` VARCHAR(32) DEFAULT NULL COMMENT '步骤类型',
    `timeout_minutes` INT DEFAULT 5 COMMENT '超时时间 (分钟)',
    `default_assignee_role` VARCHAR(64) DEFAULT NULL COMMENT '默认分配角色',
    `executor_team` VARCHAR(64) DEFAULT NULL COMMENT '执行团队',
    `phase` VARCHAR(64) DEFAULT NULL COMMENT '阶段名称',
    `phase_step` VARCHAR(64) DEFAULT NULL COMMENT '环节/子阶段',
    `pre_step_ids` JSON DEFAULT NULL COMMENT '前置步骤 ID 列表（实例步骤ID）',
    `execution_mode` VARCHAR(16) DEFAULT 'serial' COMMENT '执行模式：serial/parallel',
    `estimated_duration_minutes` INT DEFAULT NULL COMMENT '预计耗时(分钟)',
    `estimated_start_offset` INT DEFAULT NULL COMMENT '预计开始时间偏移(分钟)',
    `attributes` JSON DEFAULT NULL COMMENT '动态扩展属性',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_drill_step` (`drill_instance_id`, `status`),
    KEY `idx_parent_step` (`parent_step_id`),
    CONSTRAINT `fk_step_instance_drill` FOREIGN KEY (`drill_instance_id`) REFERENCES `drill_instance` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_step_parent_instance` FOREIGN KEY (`parent_step_id`) REFERENCES `drill_instance_step` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='步骤实例表';

-- 流程命令表
DROP TABLE IF EXISTS `drill_flow_command`;
CREATE TABLE `drill_flow_command` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `command_type` VARCHAR(64) NOT NULL,
    `drill_instance_id` BIGINT UNSIGNED NOT NULL,
    `step_instance_id` BIGINT UNSIGNED DEFAULT NULL,
    `operator_id` BIGINT UNSIGNED NOT NULL,
    `idempotency_key` VARCHAR(128) NOT NULL,
    `payload` JSON NOT NULL,
    `status` VARCHAR(20) NOT NULL,
    `worker_id` VARCHAR(128) DEFAULT NULL,
    `lease_until` DATETIME DEFAULT NULL,
    `attempts` INT NOT NULL DEFAULT 0,
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

-- 演练操作日志表
DROP TABLE IF EXISTS `drill_instance_step_log`;
CREATE TABLE `drill_instance_step_log` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `drill_instance_id` BIGINT UNSIGNED NOT NULL,
    `task_instance_id` BIGINT UNSIGNED DEFAULT NULL,
    `command_id` BIGINT UNSIGNED DEFAULT NULL,
    `action` VARCHAR(32) NOT NULL COMMENT '操作类型：complete/issue/force_complete/skip',
    `operator_id` BIGINT UNSIGNED NOT NULL COMMENT '操作人 ID',
    `operator_name` VARCHAR(64) NOT NULL COMMENT '操作人姓名',
    `content` TEXT COMMENT '操作内容/备注',
    `created_at` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_drill_instance` (`drill_instance_id`),
    KEY `idx_created_at` (`created_at`),
    KEY `idx_step_instance` (`task_instance_id`),
    UNIQUE KEY `uk_log_command_action_step` (`command_id`, `action`, `task_instance_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='演练操作日志表';

-- 演练人员分配表
DROP TABLE IF EXISTS `drill_assignee`;
CREATE TABLE `drill_assignee` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `drill_instance_id` BIGINT UNSIGNED NOT NULL COMMENT '演练实例 ID',
    `step_instance_id` BIGINT UNSIGNED NOT NULL COMMENT '步骤实例 ID',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户 ID',
    `notify_sent` TINYINT NOT NULL DEFAULT 0 COMMENT '是否已通知：0-未通知，1-已通知',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_drill_step_user` (`drill_instance_id`, `step_instance_id`, `user_id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_drill_instance_id` (`drill_instance_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='演练人员分配表';

-- 通知表
DROP TABLE IF EXISTS `notification`;
CREATE TABLE `notification` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户 ID',
    `type` VARCHAR(50) NOT NULL COMMENT '通知类型',
    `command_id` BIGINT UNSIGNED DEFAULT NULL,
    `title` VARCHAR(200) NOT NULL COMMENT '通知标题',
    `content` TEXT COMMENT '通知内容',
    `drill_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '关联演练 ID',
    `drill_name` VARCHAR(200) DEFAULT NULL COMMENT '演练名称',
    `step_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '关联步骤 ID',
    `step_name` VARCHAR(200) DEFAULT NULL COMMENT '步骤名称',
    `is_read` TINYINT NOT NULL DEFAULT 0 COMMENT '是否已读：0-未读，1-已读',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_user_created` (`user_id`, `created_at`),
    KEY `idx_user_unread` (`user_id`, `is_read`),
    UNIQUE KEY `uk_notification_command_user_type_step` (`command_id`, `user_id`, `type`, `step_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='通知表';

-- 用户数据 (密码：admin123, bcrypt hash)
INSERT INTO `user` (`username`, `real_name`, `password_hash`, `role`, `department`, `status`) VALUES
('admin', '系统管理员', '$2a$10$iJN4iIelCFVrErNjcFHlWOM0DgZeR.9YOmL.LMDYIfLUrbYHkd/.S', 'admin', '技术部', 1),
('director1', '张指挥', '$2a$10$iJN4iIelCFVrErNjcFHlWOM0DgZeR.9YOmL.LMDYIfLUrbYHkd/.S', 'director', '运维部', 1),
('executor1', '李执行', '$2a$10$iJN4iIelCFVrErNjcFHlWOM0DgZeR.9YOmL.LMDYIfLUrbYHkd/.S', 'executor', '研发部', 1),
('viewer1', '王观察', '$2a$10$iJN4iIelCFVrErNjcFHlWOM0DgZeR.9YOmL.LMDYIfLUrbYHkd/.S', 'viewer', '测试部', 1),
('director2', '刘指挥', '$2a$10$iJN4iIelCFVrErNjcFHlWOM0DgZeR.9YOmL.LMDYIfLUrbYHkd/.S', 'director', '运维部', 1);

SET FOREIGN_KEY_CHECKS = 1;
