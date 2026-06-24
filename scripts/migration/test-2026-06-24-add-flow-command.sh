#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
container="${MYSQL_CONTAINER:-drill-mysql-dev}"
root_password="${MYSQL_ROOT_PASSWORD:-rootpassword123}"
legacy_database="flow_command_legacy_test_$$"
current_database="flow_command_current_test_$$"
mixed_database="flow_command_mixed_test_$$"
databases=("$legacy_database" "$current_database" "$mixed_database")

mysql_exec() {
    docker exec -i "$container" mysql -uroot "-p${root_password}" "$@"
}

cleanup() {
    for database in "${databases[@]}"; do
        mysql_exec -e "DROP DATABASE IF EXISTS \`${database}\`" >/dev/null 2>&1 || true
    done
}
trap cleanup EXIT

initialize_database() {
    local database="$1"
    sed "s/drill_platform/${database}/g" "$repo_root/scripts/init-db.sql" | mysql_exec
}

run_migration_twice() {
    local database="$1"
    mysql_exec "$database" < "$repo_root/scripts/migration/2026-06-24-add-flow-command.sql" >/dev/null
    mysql_exec "$database" < "$repo_root/scripts/migration/2026-06-24-add-flow-command.sql" >/dev/null
}

assert_target_schema() {
    local database="$1"
    local column_count
    local index_columns

    column_count="$(
        mysql_exec -Nse "
            SELECT COUNT(*)
            FROM INFORMATION_SCHEMA.COLUMNS
            WHERE TABLE_SCHEMA = '${database}'
              AND TABLE_NAME = 'drill_instance_step_log'
              AND COLUMN_NAME = 'task_instance_id'
              AND COLUMN_TYPE = 'bigint unsigned'
              AND IS_NULLABLE = 'YES'
        "
    )"

    index_columns="$(
        mysql_exec -Nse "
            SELECT GROUP_CONCAT(COLUMN_NAME ORDER BY SEQ_IN_INDEX)
            FROM INFORMATION_SCHEMA.STATISTICS
            WHERE TABLE_SCHEMA = '${database}'
              AND TABLE_NAME = 'drill_instance_step_log'
              AND INDEX_NAME = 'uk_log_command_action_task'
              AND NON_UNIQUE = 0
        "
    )"

    test "$column_count" = "1"
    test "$index_columns" = "command_id,action,task_instance_id"
}

initialize_database "$legacy_database"
mysql_exec "$legacy_database" <<'SQL'
ALTER TABLE `drill_instance_step_log`
    DROP INDEX `uk_log_command_action_task`,
    DROP INDEX `idx_step_instance`,
    DROP COLUMN `command_id`,
    CHANGE COLUMN `task_instance_id` `step_instance_id` BIGINT UNSIGNED NULL,
    ADD INDEX `idx_step_instance` (`step_instance_id`);

ALTER TABLE `notification`
    DROP INDEX `uk_notification_command_user_type_step`,
    DROP COLUMN `command_id`;

DROP TABLE `drill_flow_command`;
SQL
run_migration_twice "$legacy_database"
assert_target_schema "$legacy_database"

initialize_database "$current_database"
run_migration_twice "$current_database"
assert_target_schema "$current_database"

initialize_database "$mixed_database"
mysql_exec "$mixed_database" <<'SQL'
ALTER TABLE `drill_instance_step_log`
    DROP INDEX `uk_log_command_action_task`,
    ADD COLUMN `step_instance_id` BIGINT UNSIGNED NULL AFTER `task_instance_id`,
    ADD UNIQUE INDEX `uk_log_command_action_task` (`command_id`, `action`, `step_instance_id`);

INSERT INTO `drill_instance_step_log` (
    `command_id`,
    `drill_instance_id`,
    `task_instance_id`,
    `step_instance_id`,
    `action`,
    `operator_id`,
    `operator_name`
) VALUES (1, 1, 10, 20, 'complete', 1, 'migration-test');
SQL
run_migration_twice "$mixed_database"
assert_target_schema "$mixed_database"

mixed_row="$(
    mysql_exec "$mixed_database" -Nse "
        SELECT CONCAT(task_instance_id, ',', step_instance_id)
        FROM drill_instance_step_log
        WHERE command_id = 1
    "
)"
test "$mixed_row" = "10,20"

echo "flow command migration compatibility test passed"
