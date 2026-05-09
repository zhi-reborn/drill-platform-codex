#!/usr/bin/env bash
# 数据库迁移脚本
# 用法: ./scripts/migrate.sh [backup|restore|migrate|status]

set -euo pipefail

MYSQL_HOST="${MYSQL_HOST:-localhost}"
MYSQL_PORT="${MYSQL_PORT:-3306}"
MYSQL_USER="${MYSQL_USER:-root}"
MYSQL_PASSWORD="${MYSQL_PASSWORD:-}"
MYSQL_DB="${MYSQL_DB:-drill_platform}"
BACKUP_DIR="${BACKUP_DIR:-./backups}"
INIT_SQL="scripts/init-db.sql"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*"
}

mysql_exec() {
    MYSQL_PWD=$MYSQL_PASSWORD mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" "$MYSQL_DB" "$@"
}

backup() {
    mkdir -p "$BACKUP_DIR"
    local backup_file="$BACKUP_DIR/drill_platform_$(date '+%Y%m%d_%H%M%S').sql"
    log "执行数据库备份..."
    MYSQL_PWD=$MYSQL_PASSWORD mysqldump -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" "$MYSQL_DB" > "$backup_file"
    log "备份完成: $backup_file"
}

restore() {
    local backup_file="$1"
    if [ ! -f "$backup_file" ]; then
        echo "错误: 备份文件不存在: $backup_file"
        exit 1
    fi
    log "从备份文件恢复: $backup_file"
    MYSQL_PWD=$MYSQL_PASSWORD mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" "$MYSQL_DB" < "$backup_file"
    log "恢复完成"
}

migrate() {
    if [ ! -f "$INIT_SQL" ]; then
        echo "错误: 初始化脚本不存在: $INIT_SQL"
        exit 1
    fi
    log "执行迁移..."
    backup
    mysql_exec < "$INIT_SQL"
    log "迁移完成"
}

status() {
    log "检查数据库连接..."
    if mysql_exec -e "SELECT 1" > /dev/null 2>&1; then
        log "数据库连接正常"
        log "数据库: $MYSQL_DB"
        log "主机: $MYSQL_HOST:$MYSQL_PORT"
        log ""
        log "表列表:"
        mysql_exec -e "SHOW TABLES" 2>/dev/null | tail -n +2
        log ""
        log "表行数:"
        mysql_exec -e "
            SELECT TABLE_NAME, TABLE_ROWS 
            FROM information_schema.TABLES 
            WHERE TABLE_SCHEMA = '$MYSQL_DB' 
            ORDER BY TABLE_NAME;
        " 2>/dev/null
    else
        echo "错误: 无法连接到数据库"
        exit 1
    fi
}

case "${1:-status}" in
    backup)
        backup
        ;;
    restore)
        if [ -z "${2:-}" ]; then
            echo "用法: $0 restore <backup_file>"
            exit 1
        fi
        restore "$2"
        ;;
    migrate)
        migrate
        ;;
    status)
        status
        ;;
    *)
        echo "用法: $0 {backup|restore|migrate|status}"
        echo ""
        echo "命令:"
        echo "  backup    备份当前数据库"
        echo "  restore   从备份文件恢复 (需指定备份文件路径)"
        echo "  migrate   执行迁移 (自动备份后执行 init-db.sql)"
        echo "  status    查看数据库状态 (默认)"
        exit 1
        ;;
esac
