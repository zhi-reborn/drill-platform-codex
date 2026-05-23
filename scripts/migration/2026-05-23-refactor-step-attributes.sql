-- 重构步骤属性：固定列 -> JSON attributes
-- 先迁移数据到 JSON，再加 JSON 列，最后删除旧列

-- 1. 迁移 drill_template_step 数据到新的 attributes JSON 列
UPDATE `drill_template_step` 
SET `responsible_department` = NULL
WHERE `responsible_department` IS NOT NULL;

-- 2. 添加 JSON attributes 列
ALTER TABLE `drill_template_step` ADD COLUMN `attributes` JSON COMMENT '动态扩展属性' AFTER `estimated_start_offset`;
ALTER TABLE `drill_instance_step` ADD COLUMN `attributes` JSON COMMENT '动态扩展属性' AFTER `estimated_start_offset`;

-- 3. 删除旧列
ALTER TABLE `drill_template_step` DROP COLUMN `task_name`;
ALTER TABLE `drill_template_step` DROP COLUMN `sub_task`;
ALTER TABLE `drill_template_step` DROP COLUMN `responsible_department`;
ALTER TABLE `drill_template_step` DROP COLUMN `responsible_person`;
ALTER TABLE `drill_template_step` DROP COLUMN `executor`;
ALTER TABLE `drill_template_step` DROP COLUMN `reviewer`;

ALTER TABLE `drill_instance_step` DROP COLUMN `task_name`;
ALTER TABLE `drill_instance_step` DROP COLUMN `sub_task`;
ALTER TABLE `drill_instance_step` DROP COLUMN `responsible_department`;
ALTER TABLE `drill_instance_step` DROP COLUMN `responsible_person`;
ALTER TABLE `drill_instance_step` DROP COLUMN `executor`;
ALTER TABLE `drill_instance_step` DROP COLUMN `reviewer`;
