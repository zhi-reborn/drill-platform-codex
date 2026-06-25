# 高可用部署指南

## 架构概述

```
                    ┌─────────┐
                    │  Nginx  │
                    │ (LB)    │
                    └────┬────┘
              ┌──────────┼──────────┐
              ▼                     ▼
        ┌──────────┐         ┌──────────┐
        │backend-a │         │backend-b │
        │(all)     │         │(all)     │
        └────┬─────┘         └────┬─────┘
              │                     │
              ▼                     ▼
        ┌──────────┐         ┌──────────┐
        │  MySQL   │         │  Redis   │
        │ (共享)    │         │ (共享)    │
        └──────────┘         └──────────┘
```

## 关键组件

### 1. 流程命令化
所有流程变更（开始/暂停/终止演练、步骤操作）通过 durable command 提交，存储在 `drill_flow_command` 表中。API handler 不再直接执行流程变更。

### 2. Worker 单例领导选举
- 使用 Redis SetNX 实现租约式领导选举
- 同一时刻只有一个节点作为 Leader 处理命令
- Leader 通过 token-fenced Lua 脚本保护租约，防止跨节点误释放
- Leader 故障后，备用节点在租约过期后自动接管

### 3. 事件分发
- WebSocket 事件通过 Redis Pub/Sub 跨节点分发
- 每个节点订阅 `drill:events` 频道，将事件投递到本地 WebSocket 客户端

### 4. 流程恢复
- 新 Leader 上线后执行 `FlowRecovery.RecoverAll`，重建运行中演练的引擎状态
- 过期超时步骤自动生成 `step_timeout` 内部命令
- 幂等键格式：`timeout:<drill-id>:<step-id>:<timeout-unix>`

## 部署步骤

### 1. 准备环境

```bash
cp .env.ha.example .env.ha
# 编辑 .env.ha，设置 JWT_SECRET 和数据库密码
```

### 2. 构建和启动

```bash
docker compose -f docker-compose.ha.yml up -d
```

### 3. 验证健康

```bash
# 检查两个节点就绪
curl http://localhost:18080/ready
curl http://localhost:18081/ready

# 通过 Nginx 访问
curl http://localhost/health
```

### 4. 验证故障切换

```bash
# 停止 Leader 节点
docker compose -f docker-compose.ha.yml stop backend-a

# 确认服务可用（备用节点接管）
curl http://localhost/health
```

## 生产环境注意事项

### MySQL 高可用
本部署使用单实例 MySQL。生产环境必须使用 MySQL 自身的高可用方案：
- MySQL Group Replication / InnoDB Cluster
- 或托管数据库服务（RDS、PolarDB 等）

### Redis 高可用
本部署使用单实例 Redis。生产环境必须使用 Redis 高可用方案：
- Redis Sentinel
- Redis Cluster
- 或托管 Redis 服务

### 配置要点
- 每个节点的 `INSTANCE_ID` 必须唯一
- `JWT_SECRET` 在所有节点间必须一致
- `WORKER_LEASE_TTL` 建议保持 15s，过短会导致频繁切换
- Nginx 的 `max_fails=3 fail_timeout=10s` 配置确保故障节点被快速摘除

### 监控
- 监控 `/ready` 端点确认节点健康
- 监控 Worker 状态（standby/recovering/leader-ready/stopping）
- 监控 `drill_flow_command` 表中 pending/processing 命令数量
- 监控 Redis Pub/Sub 连接状态
