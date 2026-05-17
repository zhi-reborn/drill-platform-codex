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

## 用户偏好与工作习惯

### 开发模式偏好

1. **联调优先**：前后端修改后必须立即测试验证
   - 后端修改 → `docker restart drill-backend-dev`
   - 前端修改 → 刷新浏览器验证
   - 数据库修改 → 直接执行 SQL 验证

2. **快速迭代**：MVP 阶段以功能可用为优先，代码可后续优化

3. **验证驱动**：修改后必须通过 API 测试或页面验证才能交付

### 数据库规范

1. **表命名**：统一使用 `drill_模块 _实体_子实体` 格式
   - `drill_template` - 演练模板
   - `drill_template_step` - 模板步骤
   - `drill_instance` - 演练实例
   - `drill_instance_step` - 实例步骤
   - `drill_instance_step_log` - 统一日志表（演练 + 步骤）

2. **级联删除**：删除主表时必须同步删除关联表
   - 删除演练 → 删除日志、人员分配、步骤实例

3. **字段扩展**：步骤相关表需要从模板继承字段
   - `step_type`, `timeout_minutes`, `default_assignee_role`, `executor_team`

### 代码风格

1. **简洁优先**：避免冗余按钮和功能
   - 功能重复的按钮只保留一个（如查看/监控）
   - 操作列宽度根据按钮数量调整

2. **注释精简**：只保留必要的注释
   - 复杂逻辑需要注释说明
   - 自解释的代码不需要注释

3. **错误处理**：区分不同类型的错误
   - 参数错误 → 400
   - 资源不存在 → 404
   - 服务器错误 → 500

