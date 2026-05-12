# AGENTS.md - 智能体工作指南

## 项目概述

演练流程管理系统 - 基于 Go + Vue3 的 IT 故障演练流程管理平台

**核心架构**: 指挥中心大屏 (展示) + 流程引擎 (驱动) + WebSocket (实时通信)

---

## 快速命令参考

### 开发模式

**推荐**：使用 docker-compose 进行开发环境联调（WSL 环境）

```bash
# 首次使用前 - 安装前端依赖（必须）
cd web && npm install && cd ..

# 启动所有服务（源码挂载模式）
docker-compose -f docker-compose.dev.yml up

# 后台运行
docker-compose -f docker-compose.dev.yml up -d

# 查看日志
docker-compose -f docker-compose.dev.yml logs -f

# 停止并清理
docker-compose -f docker-compose.dev.yml down

# 清理数据卷（重置数据库）
docker-compose -f docker-compose.dev.yml down -v
```


### 访问地址

- 前端界面：http://localhost:5173
- API 服务：http://localhost:8080
- WebSocket: ws://localhost:8081
- MySQL: localhost:3306
- Redis: localhost:6379
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

## 相关文档

- [总体设计文档](docs/生产演练流程管理系统_总体设计文档_v1.0.md)
