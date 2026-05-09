# 生产演练流程管理系统

基于 Go + Vue3 的 IT 生产系统故障演练流程管理平台，以指挥中心大屏为核心展示载体，以流程引擎为驱动核心。

## 技术栈

**后端**
- Go 1.23
- Gin (Web 框架)
- GORM v2 (ORM)
- MySQL 8.0 (数据库)
- Gorilla WebSocket (实时通信)

**前端**
- Vue 3
- Vite
- Element Plus (UI 组件库)
- TypeScript
- ECharts 5 (图表库)

**部署**
- Docker
- Docker Compose
- Nginx (反向代理)

## 项目结构

```
drill-platform/
├── cmd/                    # 应用入口
│   └── server/             # 后端服务
├── internal/               # 私有包
│   ├── api/                # API 层
│   ├── domain/             # 领域模型
│   ├── service/            # 业务逻辑
│   ├── repository/         # 数据访问
│   ├── infrastructure/     # 基础设施
│   └── pkg/                # 内部工具包
├── web/                    # 前端项目
│   ├── src/
│   ├── public/
│   └── Dockerfile
├── scripts/                # 脚本
│   ├── init-db.sql         # 数据库初始化
│   └── docker-deploy.sh    # 部署脚本
├── nginx/                  # Nginx 配置
│   └── nginx.conf
├── docker-compose.yml      # Docker 编排
├── Dockerfile              # 后端镜像
└── .env.example            # 环境变量模板
```

## 快速开始

### 1. 环境准备

- Docker 20.10+
- Docker Compose 2.0+

### 2. 配置环境变量

```bash
cp .env.example .env
# 编辑 .env 文件，修改必要的配置
```

### 3. 一键部署

```bash
# 赋予执行权限
chmod +x scripts/docker-deploy.sh

# 启动所有服务
./scripts/docker-deploy.sh deploy
```

### 4. 访问系统

- 前端界面：http://localhost
- API 服务：http://localhost:8080
- WebSocket: ws://localhost:8081

### 5. 默认账户

- 用户名：`admin`
- 密码：`admin123`

## 常用命令

```bash
# 部署所有服务
./scripts/docker-deploy.sh deploy

# 构建镜像
./scripts/docker-deploy.sh build

# 启动服务
./scripts/docker-deploy.sh start

# 停止服务
./scripts/docker-deploy.sh stop

# 重启服务
./scripts/docker-deploy.sh restart

# 查看日志
./scripts/docker-deploy.sh logs

# 健康检查
./scripts/docker-deploy.sh health

# 清理资源 (包括数据卷)
./scripts/docker-deploy.sh clean
```

## Docker Compose 服务

| 服务名 | 容器名 | 端口 | 说明 |
|--------|--------|------|------|
| mysql | drill-platform-mysql | 3306 | MySQL 8.0 数据库 |
| redis | drill-platform-redis | 6379 | Redis 7 缓存 |
| backend | drill-platform-backend | 8080, 8081 | Go 后端服务 |
| frontend | drill-platform-frontend | - | Vue3 前端 (内部) |
| nginx | drill-platform-nginx | 80, 443 | Nginx 反向代理 |

## 开发模式

### 后端开发

```bash
cd cmd/server
go run main.go
```

### 前端开发

```bash
cd web
npm install
npm run dev
```

## 数据库表

| 表名 | 说明 |
|------|------|
| users | 用户表 |
| drill_processes | 演练流程定义表 |
| process_steps | 流程步骤表 |
| drill_instances | 演练实例表 |
| step_executions | 步骤执行记录表 |
| operation_logs | 操作日志表 |
| system_configs | 系统配置表 |

## 核心设计原则

- **轻量优先**: MVP 阶段自研轻量级状态机引擎，不引入重型工作流引擎
- **实时优先**: WebSocket 全双工通信，状态变更端到端延迟 < 1s
- **扩展优先**: 模块化设计，后续可无缝接入真实监控指标、故障注入能力

## 注意事项

1. **生产环境**: 务必修改 `.env` 中的默认密码和 JWT_SECRET
2. **数据持久化**: MySQL 和 Redis 数据存储在 Docker volume 中
3. **日志目录**: 后端日志输出到 `./logs` 目录
4. **前端构建**: 生产环境需先执行 `npm run build` 构建前端


