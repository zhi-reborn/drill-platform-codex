CREATE TABLE IF NOT EXISTS `notification` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '接收用户 ID',
  `type` VARCHAR(50) NOT NULL COMMENT '通知类型',
  `title` VARCHAR(200) NOT NULL COMMENT '标题',
  `content` TEXT COMMENT '内容',
  `drill_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '关联演练 ID',
  `drill_name` VARCHAR(200) DEFAULT NULL COMMENT '演练名称',
  `step_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '关联步骤 ID',
  `step_name` VARCHAR(200) DEFAULT NULL COMMENT '步骤名称',
  `is_read` TINYINT NOT NULL DEFAULT 0 COMMENT '是否已读 0=未读 1=已读',
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  INDEX `idx_user_created` (`user_id`, `created_at` DESC),
  INDEX `idx_user_unread` (`user_id`, `is_read`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户通知表';
