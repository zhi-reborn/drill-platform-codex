#!/bin/bash
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PROJECT_NAME="drill-platform"
COMPOSE_CMD="docker compose"

log_info()    { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn()    { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error()   { echo -e "${RED}[ERROR]${NC} $1"; }
log_step()    { echo -e "${BLUE}[STEP]${NC} $1"; }

check_deps() {
    log_info "检查依赖..."
    if ! command -v docker &>/dev/null; then
        log_error "需要安装 Docker"
        exit 1
    fi
    if docker compose version &>/dev/null; then
        COMPOSE_CMD="docker compose"
    elif command -v docker-compose &>/dev/null; then
        COMPOSE_CMD="docker-compose"
    else
        log_error "需要安装 Docker Compose (docker compose 或 docker-compose)"
        exit 1
    fi
    log_info "使用: $COMPOSE_CMD"
}

load_env() {
    if [ ! -f .env ]; then
        log_warn "未找到 .env，从 .env.example 复制"
        cp .env.example .env
        log_warn "默认值已启用，编辑 .env 可自定义配置"
    fi
}

build() {
    log_step "构建所有服务镜像..."
    $COMPOSE_CMD build --pull "$@"
    log_info "构建完成"
}

deploy() {
    check_deps
    load_env

    log_step "启动基础设施 (MySQL, Redis)..."
    $COMPOSE_CMD up -d mysql redis

    log_info "等待数据库就绪..."
    local max_wait=60
    local waited=0
    while [ $waited -lt $max_wait ]; do
        if $COMPOSE_CMD exec -T mysql mysqladmin ping -h localhost --silent 2>/dev/null; then
            log_info "MySQL 就绪 ($waited s)"
            break
        fi
        sleep 2
        waited=$((waited + 2))
    done

    if [ $waited -ge $max_wait ]; then
        log_error "MySQL 启动超时"
        $COMPOSE_CMD logs mysql
        exit 1
    fi

    log_step "初始化数据库..."
    if [ -f scripts/init-db.sql ]; then
        $COMPOSE_CMD exec -T mysql mysql -u"${MYSQL_USER:-drill_user}" -p"${MYSQL_PASSWORD:-drill_password123}" "${MYSQL_DATABASE:-drill_platform}" < scripts/init-db.sql 2>/dev/null || \
        log_warn "数据库可能已初始化"
    fi

    log_step "启动应用服务..."
    $COMPOSE_CMD up -d backend frontend nginx

    log_step "健康检查..."
    sleep 5
    health

    log_info "部署完成！"
    log_info "前端: http://localhost:${HTTP_PORT:-80}"
    log_info "API:   http://localhost:${API_PORT:-8080}"
}

start() {
    load_env
    log_step "启动所有服务..."
    $COMPOSE_CMD up -d
    log_info "启动完成"
}

stop() {
    log_step "停止所有服务..."
    $COMPOSE_CMD stop
    log_info "已停止"
}

restart() {
    stop
    sleep 2
    start
}

logs() {
    local service="${1:--f}"
    if [ "$service" = "-f" ]; then
        $COMPOSE_CMD logs -f
    else
        $COMPOSE_CMD logs --tail=100 "$service"
    fi
}

health() {
    log_info "服务健康状态"
    echo "-------------------------------------------"

    check_service "MySQL" "$COMPOSE_CMD exec -T mysql mysqladmin ping -h localhost --silent"
    check_service "Redis" "$COMPOSE_CMD exec -T redis redis-cli ping"
    check_service "后端" "curl -sf http://localhost:${API_PORT:-8080}/health"
    check_service "Nginx" "curl -sf http://localhost:${HTTP_PORT:-80}/"

    echo "-------------------------------------------"
}

check_service() {
    local name="$1"
    local cmd="$2"
    if eval "$cmd" &>/dev/null; then
        echo -e "  ${GREEN}●${NC} $name: 正常"
    else
        echo -e "  ${RED}●${NC} $name: 异常"
    fi
}

clean() {
    log_warn "将删除所有容器、网络和匿名卷"
    read -p "确认继续？[y/N] " confirm
    if [ "$confirm" = "y" ] || [ "$confirm" = "Y" ]; then
        log_step "清理资源..."
        $COMPOSE_CMD down -v --remove-orphans
        log_info "清理完成"
    else
        log_info "已取消"
    fi
}

status() {
    $COMPOSE_CMD ps
}

usage() {
    cat << EOF
用法: $0 <command> [options]

命令:
  deploy      一键部署所有服务
  build       构建 Docker 镜像
  start       启动所有服务
  stop        停止所有服务
  restart     重启所有服务
  logs        查看服务日志 [--tail N] [service_name]
  health      检查服务健康状态
  status      查看服务运行状态
  clean       清理所有资源 (容器/网络/卷)
  help        显示此帮助信息

示例:
  $0 deploy
  $0 build --no-cache
  $0 logs backend
  $0 status
EOF
}

cd "$(dirname "$0")/.."

case "${1:-help}" in
    deploy)   deploy ;;
    build)    shift; build "$@" ;;
    start)    start ;;
    stop)     stop ;;
    restart)  restart ;;
    logs)     shift; logs "$@" ;;
    health)   health ;;
    status)   status ;;
    clean)    clean ;;
    help)     usage ;;
    *)        log_error "未知命令: $1"; usage; exit 1 ;;
esac
