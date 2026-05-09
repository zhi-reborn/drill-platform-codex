-- 演练平台数据库初始化脚本
-- 基于《生产演练流程管理系统_总体设计文档_v1.0》

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- 创建数据库
CREATE DATABASE IF NOT EXISTS drill_platform DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE drill_platform;

-- 用户表
DROP TABLE IF EXISTS `user`;
CREATE TABLE `user` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `username` VARCHAR(64) NOT NULL COMMENT '用户名',
    `real_name` VARCHAR(64) NOT NULL COMMENT '真实姓名',
    `password_hash` VARCHAR(256) NOT NULL COMMENT '密码哈希 (bcrypt)',
    `role` VARCHAR(32) NOT NULL COMMENT '角色：admin/director/executor/viewer',
    `department` VARCHAR(64) DEFAULT NULL COMMENT '部门',
    `phone` VARCHAR(20) DEFAULT NULL COMMENT '手机号',
    `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1-启用，0-禁用',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_username` (`username`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

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
DROP TABLE IF EXISTS `step_template`;
CREATE TABLE `step_template` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `drill_template_id` BIGINT UNSIGNED NOT NULL COMMENT '所属模板 ID',
    `name` VARCHAR(128) NOT NULL COMMENT '步骤名称',
    `seq` INT NOT NULL COMMENT '排序序号',
    `step_type` VARCHAR(32) NOT NULL COMMENT '类型：serial/parallel/any_of/condition',
    `timeout_minutes` INT NOT NULL DEFAULT 5 COMMENT '超时时间 (分钟)',
    `pre_step_ids` JSON DEFAULT NULL COMMENT '前置步骤 ID 列表',
    `guide_content` TEXT COMMENT '操作指引',
    `is_blocking` TINYINT NOT NULL DEFAULT 1 COMMENT '是否阻塞：0-非阻塞，1-阻塞',
    `default_assignee_role` VARCHAR(64) DEFAULT NULL COMMENT '默认分配角色',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_drill_template_id` (`drill_template_id`),
    CONSTRAINT `fk_step_template_drill` FOREIGN KEY (`drill_template_id`) REFERENCES `drill_template` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='步骤模板表';

-- 演练实例表
DROP TABLE IF EXISTS `drill_instance`;
CREATE TABLE `drill_instance` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `template_id` BIGINT UNSIGNED NOT NULL COMMENT '来源模板 ID',
    `name` VARCHAR(128) NOT NULL COMMENT '演练名称',
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
DROP TABLE IF EXISTS `step_instance`;
CREATE TABLE `step_instance` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `drill_instance_id` BIGINT UNSIGNED NOT NULL COMMENT '所属演练 ID',
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
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_drill_step` (`drill_instance_id`, `status`),
    CONSTRAINT `fk_step_instance_drill` FOREIGN KEY (`drill_instance_id`) REFERENCES `drill_instance` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='步骤实例表';

-- 步骤操作日志表
DROP TABLE IF EXISTS `step_instance_log`;
CREATE TABLE `step_instance_log` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `step_instance_id` BIGINT UNSIGNED NOT NULL COMMENT '步骤实例 ID',
    `action` VARCHAR(32) NOT NULL COMMENT '操作类型：complete/issue/force_complete/skip',
    `operator_id` BIGINT UNSIGNED NOT NULL COMMENT '操作人 ID',
    `operator_name` VARCHAR(64) NOT NULL COMMENT '操作人姓名',
    `content` TEXT COMMENT '操作内容/备注',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_step_instance` (`step_instance_id`),
    CONSTRAINT `fk_step_log_step` FOREIGN KEY (`step_instance_id`) REFERENCES `step_instance` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='步骤操作日志表';

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

SET FOREIGN_KEY_CHECKS = 1;
