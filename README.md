# 演练流程管理系统

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


## 核心设计原则

- **轻量优先**: MVP 阶段自研轻量级状态机引擎，不引入重型工作流引擎
- **实时优先**: WebSocket 全双工通信，状态变更端到端延迟 < 1s
- **扩展优先**: 模块化设计，后续可无缝接入真实监控指标、故障注入能力



