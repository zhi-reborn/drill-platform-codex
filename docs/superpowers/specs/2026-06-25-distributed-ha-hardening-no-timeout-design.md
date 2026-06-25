# 分布式高可用加固与取消超时设计

## 1. 背景与边界

当前代码已经具备 MySQL 流程命令表、Redis Worker 租约、Redis Pub/Sub、健康检查和 Nginx upstream 等基础，但生产执行路径仍绕过事务性命令执行器，Redis 租约不能严格阻止旧主写入，API 读取会推进流程，WebSocket 主路径仍可能只广播到本机。

本次加固基于以下已确认前提：

- 内网 MySQL、Redis、Nginx 均由基础设施团队以集群或多活方式提供。
- 应用仓库不负责实现 MySQL、Redis、Nginx 集群本身，只负责连接配置、应用节点部署和正确使用这些集群。
- 所有任务取消超时，不论执行多久都不会再自动进入 `timeout`。
- 为兼容历史数据，数据库中的超时字段和历史 `timeout` 状态保留。

## 2. 目标

- API/WebSocket 节点可以无会话粘滞地 Active-Active 部署。
- Redis 选主提供快速故障切换，MySQL generation fencing 严格阻止旧主提交。
- 所有流程修改必须通过 MySQL 命令队列执行。
- 业务状态、日志、通知和命令结果在同一事务内提交。
- 同一演练的命令严格串行，不同演练可以并行。
- API 查询操作完全只读，不恢复、不修正、不推进流程。
- 所有实时事件经 Redis 跨节点分发，并使用统一 WebSocket 消息协议。
- 新任务永不产生超时；历史超时数据仍可读取。
- 新机器可以通过外部 MySQL/Redis/Nginx 地址完成应用节点部署。

## 3. 最终运行架构

每台应用机器运行同一镜像，通过 `APP_ROLE` 区分：

- `api`：HTTP、WebSocket、命令提交和查询；不创建可写流程引擎。
- `worker`：参与 Redis 选主并执行流程命令；不暴露公共业务 API。
- `all`：兼具两种职责，适合节点数量较少的部署。

所有节点连接外部高可用 MySQL 和 Redis。Nginx upstream 使用机器内网地址、服务发现结果或部署平台提供的后端池，不依赖 Docker Compose 内部 DNS。

## 4. Redis 选主与 MySQL Generation Fencing

### 4.1 Worker epoch 表

新增单行表 `drill_worker_epoch`：

| 字段 | 说明 |
| --- | --- |
| `id` | 固定值 `1` |
| `generation` | 单调递增的 Worker 代次 |
| `worker_id` | 当前获取代次的 Worker |
| `lease_token` | 当前 Redis 租约 token 的摘要或完整值 |
| `updated_at` | 更新时间 |

`drill_flow_command` 增加：

- `worker_generation`
- `worker_id`

### 4.2 获取领导权

1. Worker 使用现有 token-fenced Redis 租约竞争领导权。
2. 获得 Redis 租约后，在 MySQL 事务中 `SELECT ... FOR UPDATE` 锁定 epoch 单行。
3. 将 `generation + 1` 写回，并记录 `worker_id` 和当前租约 token。
4. Worker 只有成功取得 MySQL generation 后才进入 `leader-ready`。
5. 恢复演练和接管过期命令均使用该 generation。

### 4.3 严格 fencing

领取命令时写入当前 `worker_id` 和 `worker_generation`。

每次业务事务提交前必须在同一事务中锁定 epoch 行并校验：

```text
command.worker_generation == epoch.generation
command.worker_id == epoch.worker_id
command.status == processing
```

任意条件不满足时事务回滚，旧主不能写入业务状态、日志、通知或命令结果。

Redis 租约负责快速发现领导权变化；MySQL generation 是最终写入栅栏。即使旧主因暂停、网络分区或取消不及时而继续运行，也无法在新主取得更高 generation 后提交。

## 5. Worker 执行与续租

- Worker 续租循环与命令执行解耦，命令执行不能阻塞 Redis 续租。
- 续租失败立即取消当前 leader context，停止领取新命令。
- 已开始命令收到取消信号后应尽快退出。
- 即使命令未及时响应取消，MySQL fencing 仍阻止它提交。
- 命令处理租约可周期延长，防止正常长任务被提前重新排队。
- Worker 切换回 standby 前不主动删除其他 token 的 Redis 租约。

## 6. 命令串行与事务边界

### 6.1 同演练串行

执行命令前，在独占数据库连接上获取：

```sql
SELECT GET_LOCK(CONCAT('drill-flow:', ?), 5);
```

该命名锁覆盖一次命令的完整业务事务。命令结束后显式 `RELEASE_LOCK`；连接断开时 MySQL 自动释放。

不同演练使用不同锁名，可以并行执行。

### 6.2 统一生产执行路径

删除生产环境中“依赖非空则调用旧 Service/内存引擎直接写库”的分支。

每种命令必须由事务性执行器处理：

- 读取数据库当前状态；
- 校验 generation；
- 执行条件状态转换；
- 计算并写入后续步骤状态和演练进度；
- 写日志和通知，并设置 `command_id`；
- 将命令写为 `succeeded` 或稳定业务失败；
- 提交事务；
- 提交成功后发布事件。

内存流程引擎只允许作为纯计算组件，不能直接写数据库、创建通知、写日志或广播 WebSocket。

### 6.3 幂等

- 相同 `Idempotency-Key` 必须对应相同的操作人、命令类型、演练、步骤和 payload。
- 同一键被用于不同请求时返回冲突错误，不复用旧命令。
- 状态已达到目标时按幂等成功处理，不重复日志、通知和事件。
- 日志、通知和其他副作用均携带 `command_id`。

## 7. API 节点无状态化

- API 节点不维护可写流程实例。
- 所有修改 handler 只做鉴权、参数校验和命令提交。
- `GET /drills/:id/steps` 等查询只读取数据库或 Redis 缓存。
- 查询路径不得调用 `Recover`、`AdvanceFlow`、父步骤协调或任何数据库修正方法。
- 缓存更新只由事务提交后的失效事件触发。
- Redis 不可用时，修改命令仍可写入 MySQL并返回 `202`，但节点 `/ready` 返回 `503`，避免继续承接生产流量。

## 8. WebSocket 跨节点事件

### 8.1 唯一事件出口

业务服务和流程引擎不得直接调用本机 `wsManager`。

事务执行器在提交后发布标准事件：

```json
{
  "id": "event-uuid",
  "event_type": "step_completed",
  "drill_id": 72,
  "user_id": 10,
  "timestamp": "2026-06-25T10:00:00+08:00",
  "payload": {}
}
```

Redis 事件 payload 必须是前端现有 WebSocket 客户端可直接解析的完整消息，而不是原始命令 payload。

### 8.2 节点广播

- 每个 API 节点订阅 Redis 事件频道。
- 收到事件后只广播给连接在本机且 drill/user 匹配的客户端。
- WebSocket 每个 frame 只发送一条 JSON，不拼接多条 JSON。
- 关闭连接时，先从 Manager 移除客户端，再关闭发送通道，消除 send-on-closed-channel 竞态。
- 事件包含唯一 ID；节点使用有界短期集合去重。

### 8.3 最终一致

Redis Pub/Sub 不承担历史重放。客户端每次连接或重连成功后重新拉取：

- 演练详情；
- 步骤状态；
- 我的任务；
- 通知列表。

事件发布失败写入可观测错误日志和指标。数据库仍是最终状态来源。

## 9. 取消新任务超时

### 9.1 删除运行逻辑

- 删除 `TimeoutScheduler` 及其启动、注册、取消和事件逻辑。
- 删除 `step_timeout` 内部命令。
- 恢复流程时不扫描 `timeout_at`，不生成超时命令。
- 新步骤启动时不写 `timeout_at`。
- 流程不会因执行时间自动改变状态。

### 9.2 历史兼容

保留数据库字段：

- `timeout_minutes`
- `timeout_at`

保留历史状态：

- `timeout`

历史 `timeout` 仍被视为终态，以免旧演练重新推进或改变历史结果。接口可以继续返回这些历史字段和状态，但新记录写入 `timeout_at = NULL`。

### 9.3 前端

- 模板编辑器移除超时分钟配置。
- 任务详情和任务列表不展示超时限制、截止时间和超时提醒。
- 不再处理新的 `step_timeout` WebSocket 事件。
- 历史演练页面仍能将历史 `timeout` 显示为“历史超时”。
- 屏幕中的进行时间只按 `start_time` 计算，不与超时分钟比较。

## 10. 健康检查

`/ready` 对各角色执行严格依赖检查：

### API/all

- MySQL 可访问；
- Redis 可访问；
- Redis 订阅已建立；
- 配置和数据库 schema 版本正确。

### worker/all

- MySQL 可访问；
- Redis 可访问；
- Worker 正常参与选主；
- leader 必须持有有效 Redis 租约和当前 MySQL generation。

Redis 初始化失败不能被当作可忽略依赖。缺失必需依赖时 `/ready` 返回 `503`。

standby Worker 可以 ready，但响应必须明确返回 `standby`。

## 11. 数据库初始化与迁移

- 修正 `scripts/init-db.sql`，使字段名、默认值和实体一致。
- 新增 migration ledger 和有序迁移执行入口，应用部署前或独立 migration job 执行。
- 启动服务不自动修改生产 schema，但 `/ready` 检查预期 schema 版本。
- 支持现有旧库升级，迁移必须可重复判断、不可清空现有业务数据。
- 超时字段不删除，避免破坏历史数据。

## 12. 部署

应用节点配置至少包括：

```text
APP_ROLE
INSTANCE_ID
DATABASE_HOST
DATABASE_PORT
DATABASE_USER
DATABASE_PASSWORD
DATABASE_NAME
REDIS_ADDR
REDIS_PASSWORD
JWT_SECRET
PUBLIC_BASE_URL
CAS_PUBLIC_URL
CAS_SERVICE_URL
WORKER_LEASE_TTL
WORKER_RENEW_INTERVAL
COMMAND_WAIT_TIMEOUT
```

仓库中的 Docker Compose 仅作为单机双节点验收环境。生产文档使用以下逻辑拓扑：

```text
Nginx 集群/VIP
  -> application-node-1:8080
  -> application-node-2:8080
  -> application-node-N:8080

application nodes
  -> MySQL cluster endpoint
  -> Redis cluster endpoint
```

通用 Nginx 示例必须使用可替换的内网地址或模板变量，不依赖 `backend-a`、`backend-b` Compose DNS。

基础设施负责依据 `/ready` 从 upstream 摘除节点。

## 13. 验证标准

### 自动化

- 旧主 generation 低于当前 generation 时无法提交。
- 命令执行超过 Redis 租约周期时，续租不被阻塞。
- 同一演练命令严格串行，不同演练并行。
- 生产执行路径状态、日志、通知和命令结果同事务提交。
- 重放相同命令不产生重复副作用。
- API GET 不产生任何数据库写入。
- 未来任意时长运行步骤不会产生 timeout。
- 历史 timeout 状态仍正常读取并视为终态。
- Redis 事件在两个真实 API 节点间传播，客户端能解析完整消息。
- WebSocket 重连后主动补拉状态。
- Redis 或订阅异常时 `/ready` 返回 `503`。
- 新库初始化后 schema 与实体一致。

### 双节点故障验收

1. 启动两个真实应用节点并连接外部或测试 MySQL/Redis。
2. WebSocket 连接节点 B，命令请求发送到节点 A，节点 B 收到标准事件。
3. 创建一个长时间 processing 命令，停止或隔离 Leader。
4. Standby 获得更高 generation 并接管。
5. 旧 Leader 恢复后尝试提交，MySQL fencing 拒绝。
6. 新 Leader 完成命令，状态、日志、通知各只有一份。
7. 步骤持续运行超过原 timeout_minutes，状态仍为 running。
8. 停止任一 API 节点，Nginx 集群仍可提供 HTTP 和 WebSocket 服务。

只有以上验证全部通过，才认为应用代码满足本次多机器分布式部署要求。
