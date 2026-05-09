# AGENTS.md - 智能体工作指南

## 项目概述

生产演练流程管理系统 - 基于 Go + Vue3 的 IT 故障演练流程管理平台

**核心架构**: 指挥中心大屏 (展示) + 流程引擎 (驱动) + WebSocket (实时通信)

---

## 快速命令参考


### 开发模式

```bash
# 后端开发
cd cmd/server && go run main.go

# 前端开发
cd web && npm install && npm run dev
```

### 访问地址

- 前端界面：http://localhost
- API 服务：http://localhost:8080
- WebSocket: ws://localhost:8081
- 默认账户：`admin` / `admin123`

---

## 项目结构边界

```
drill-platform/
├── cmd/server/           # 后端入口 (Go 1.23)
├── internal/             # 私有包 (不可外部引用)
│   ├── api/              # API 层 (handler/middleware/router)
│   ├── domain/           # 领域模型 (entity/dto)
│   ├── service/          # 业务逻辑层
│   ├── repository/       # 数据访问层 (GORM)
│   ├── infrastructure/   # 基础设施 (websocket/redis/oss)
│   └── pkg/              # 内部工具包 (flowengine/validator/response)
├── web/                  # 前端项目 (Vue3 + Vite + TypeScript)
├── scripts/              # 部署脚本 (docker-deploy.sh, init-db.sql)
├── nginx/                # Nginx 反向代理配置
└── docker-compose.yml    # Docker 编排
```

**重要**: `internal/` 目录遵循 Go 私有包约定，不可被外部模块引用。

---

## 开发约束与规范

### 核心设计原则

1. **轻量优先**: MVP 阶段用自研轻量状态机引擎，不引入重型 BPMN 引擎
2. **实时优先**: WebSocket 端到端延迟 < 1s，状态变更秒级同步
3. **扩展优先**: 模块化设计，后续可接入监控指标/故障注入

### 技术选型约束

| 层级 | 技术 | 版本 | 说明 |
|------|------|------|------|
| 后端 | Go | 1.23+ | 不可降级 |
| ORM | GORM | v2 | 不用 gopkg.in 旧版 |
| WebSocket | Gorilla | - | 不用标准库 websocket |
| 前端 | Vue | 3.4+ | 组合式 API，不用 Options API |
| UI 库 | Element Plus | 2.5+ | 不用 Element UI (Vue2) |
| 图表 | ECharts | 5.4+ | 不用 Chart.js |

### 禁止行为

- ❌ 用 `as any` / `@ts-ignore` / `@ts-expect-error` 绕过类型检查
- ❌ 引入重型工作流引擎 (Flowable/Camunda)
- ❌ 前端用 Options API (必须用 Composition API)
- ❌ 提交时包含 `.env` 文件 (已在 .gitignore)

---

## 测试规范

### 后端测试

```bash
# 单文件测试
go test -v ./internal/pkg/flowengine/engine_test.go

# 单包测试
go test -v ./internal/pkg/flowengine/

# 带覆盖率
go test -cover ./...
```

### 前端测试

```bash
# 组件测试 (Vitest)
cd web && npm run test

```

### 集成测试 prerequisites

运行集成测试前需确保:
1. MySQL 容器运行中 (`docker-compose ps | grep mysql`)
3. 数据库已初始化 (`scripts/init-db.sql` 已执行)

---

## OpenSpec 工作流

本项目使用 OpenSpec 规范驱动开发流程。

### 命令速查

```bash
# 创建新变更
openspec new change "<kebab-case-name>"

# 查看变更状态
openspec status --change "<name>"

# 查看 artifact 指令
openspec instructions <artifact-id> --change "<name>"

# 实施变更
/openspec-apply-change 
```

### 变更位置

所有变更位于 `openspec/changes/<name>/` 目录，包含:
- `proposal.md` - 变更概述
- `design.md` - 技术设计
- `specs/` - 规范文档
- `tasks.md` - 任务清单

---

## 数据库规范

### 核心表

| 表名 | 说明 | 关键字段 |
|------|------|----------|
| `users` | 用户表 | `role` (admin/director/executor/viewer) |
| `drill_template` | 演练模板 | `category` (灾备/降级/发布/安全) |
| `step_template` | 步骤模板 | `step_type` (serial/parallel/any_of/condition) |
| `drill_instance` | 演练实例 | `status` (pending/running/paused/completed/terminated) |
| `step_instance` | 步骤实例 | `status` (pending/running/completed/timeout/skipped/issue) |
| `step_instance_log` | 操作日志 | `action` (complete/issue/force_complete/skip) |
| `drill_assignee` | 人员分配 | 唯一索引 `uk_drill_step_user` |

### 初始化流程

```bash
# 手动初始化数据库
docker-compose exec mysql mysql -uroot -p drill_platform < scripts/init-db.sql
```

---

## WebSocket 协议

### 连接路径

- 大屏端：`/ws/display/:drillId` - 订阅演练状态
- 工作台：`/ws/tasks` - 订阅个人任务
- 指挥台：`/ws/control/:drillId` - 控制事件

### 消息类型

```json
{
  "event_type": "step_complete",
  "drill_id": 100,
  "payload": { /* 步骤详情 */ },
  "timestamp": 1715750732
}
```

---

## 环境配置

### 配置文件

- 后端：`configs/config.yaml`
- 前端：`web/.env` (开发), `web/.env.production` (生产)
- Docker: `.env` (根目录)

---

## 相关文档

- [总体设计文档](docs/生产演练流程管理系统_总体设计文档_v1.0.md)
