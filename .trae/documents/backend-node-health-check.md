# 后端节点健康状态检查 + 系统健康页面

## Context

系统支持多实例部署（`app_role: api/worker/all` + `instance_id`），但健康信息只散落在无鉴权的 `/ready` 探针里，管理员无法从一个页面总览所有节点的运行状态。本次新增：

- **零依赖**（用户确认）：组件连通性 + Go 进程运行时指标（runtime 标准库），不引入 gopsutil
- **多节点聚合**（用户确认）：各实例每 10s 向 Redis 心跳上报健康快照，API 聚合展示所有节点，心跳超 30s 判 offline
- **管理员专属**（用户确认）：前端 `/admin/health` 页面，仅 admin 可见

## 后端改动（4 文件）

### 1. 新建 `internal/infrastructure/nodereg/reg.go`（~80 行）

参考 presence/store.go 风格的小包。Redis key `drill:health:nodes`（Hash）。

- `Reporter{client redis.UniversalClient, instanceID string}`：
  - `Report(ctx, snapshot []byte) error` — 用 Redis `TIME` 命令取统一时钟，`HSET <id>=<json>, <id>:hb=<now_unix>`，key `Expire 24h` 兜底
  - `Deregister(ctx) error` — `HDEL` 两个字段（优雅关闭时调用）
- `Registry.List(ctx) ([]Record, error)` — `TIME` + `HGETALL`，按 `<id>`/`<id>:hb` 配对成 `Record{InstanceID, Data, HeartbeatAt}`；顺手清理心跳超 1h 的孤儿条目（crash 残留）
- 常量：`OfflineAfter = 30s`、`PruneAfter = 1h`；内部操作 1s ctx 超时
- **时钟策略**：offline 判定一律用 Redis TIME，规避跨节点时钟偏差

### 2. 修改 `internal/api/handler/health/handler.go`（+~160 行）

- 新类型（JSON tag 全 snake_case）：
  - `ComponentStatus{status string, latency_ms int64}`
  - `DBPoolStats{open, in_use, idle, wait_count int, wait_duration_ms int64}`
  - `Snapshot{instance_id, role, ready, worker_status, components map[string]ComponentStatus, ws_connections map[string]int, goroutines int, heap_alloc_bytes, sys_bytes uint64, uptime_seconds, started_at int64, db_pool *DBPoolStats}`
  - `NodeHealth{Snapshot 内嵌展开; state string(online/offline); heartbeat_at int64}`
  - `SystemHealthResp{overall string(healthy/degraded), updated_at int64, warning string,omitempty, nodes []NodeHealth}`
- Handler 增私有字段：`registry *nodereg.Registry`、`wsStats func() map[string]any`、`dbStats func() *sql.DBStats`、`startedAt time.Time`（NewHandler 内初始化）。**NewHandler 签名不变**，新依赖走 setter（同 SetPublisher 惯例）：`SetRegistry` / `SetWSStats` / `SetDBStats`
- 重构：`Ready()` 内组件检查抽为 `checkComponents(ctx) (map[string]ComponentStatus, reason string)`，`Ready()` 复用，对外响应格式不变
- 新方法 `BuildSnapshot(ctx) Snapshot`：3s 超时；组件 Ping 记录延迟 ms；`runtime.NumGoroutine()` + `runtime.ReadMemStats`（HeapAlloc/Sys）；uptime/startedAt；wsStats/dbStats 可空
- 新 handler `SystemHealth(c *gin.Context)`：
  - 本地 `BuildSnapshot` 实时数据 → registry 为 nil 或 List 出错时降级 `nodes=[本地]` + `warning`（**仍 200**，健康页是诊断 Redis 故障的入口，不能 500）
  - 成功聚合：各 Record unmarshal，`age > OfflineAfter` 标 offline；本地 instanceID 的条目用实时快照覆盖并标 online
  - overall：有 offline 或任一 online 节点组件 fail / !ready → `degraded`，否则 `healthy`
  - `response.Success(c, resp)`

### 3. 修改 `cmd/server/main.go`（+~45 行）

第 10 步 healthHandler 创建后：

- redisClient 非空：`nodereg.NewReporter(redisClient.Raw(), cfg.InstanceID)` + `nodereg.NewRegistry(...)`，`healthHandler.SetRegistry(...)`；无 Redis（本地裸跑）则不设，页面退化为单节点
- `SetWSStats(wsManager.GetStats)`；`SetDBStats`（闭包读 `repository.DB.DB().Stats()`，handler 引 repository 已有先例）
- 心跳 goroutine（项目惯例：生命周期在 main 管理，同 worker/subscriber）：
  - hbCtx/hbCancel + hbDone chan；先立即上报一次，再 `ticker 10s` 循环 `BuildSnapshot → json.Marshal → Report`（失败仅 log）
  - ctx.Done 后用独立 2s ctx 调 `Deregister`，close(hbDone)
- 关闭序列：`SetReady(false)` 后立即 `hbCancel()`；`redisClient.Close()` 之前 `<-hbDone`（保证 HDEL 先于 Redis 关闭）

### 4. 修改 `internal/api/router/router.go`（+1 行）

users 路由附近加：

```go
v1.GET("/system/health", middleware.RequireAdmin(), healthHandler.SystemHealth)
```

选 `/system/health` 而非 `/admin/health`：与根路径无鉴权 `/health`（live 别名）天然区分，且 API 路径无 /admin 前缀先例。

## 前端改动（5 文件）

### 5. 新建 `web/src/api/modules/health.ts`（~40 行）

TS 接口与后端 JSON snake_case 对齐（SystemHealthResp/NodeHealth/Snapshot/ComponentStatus/DBPoolStats）；`healthApi.getSystemHealth()` → `apiRequest<SystemHealthResp>({ url: '/v1/system/health', method: 'GET' })`。`web/src/api/index.ts` 加 re-export。

### 6. 新建 `web/src/views/admin/HealthView.vue`（~250 行）

风格对齐 admin/DashboardView（`.page-container > .page-header + .page-content`，scoped style `@use '@/styles/variables' as *`）：

- header：标题「节点健康」+ 手动刷新按钮 + 自动刷新 el-switch
- 总状态 `el-alert`（healthy 绿 / degraded 橙，附 warning 文案）+ 在线/离线节点数统计
- 节点卡片 `el-row/el-col/el-card`：instance_id、state tag、role / worker_status、uptime / goroutines / 堆内存（局部 `formatBytes`/`formatUptime`）、WS 连接合计、组件状态点（ok 绿 / fail 红 + latency_ms，缺省组件显示 —）、db_pool InUse/Open
- `onMounted` 加载 + `setInterval` 10s 自动刷新，`onUnmounted` 清理
- offline/overall 由后端算好，前端只渲染（判定依赖 Redis 时钟，前端无法复算）

### 7. 修改 `web/src/router/index.ts`（+7 行）

/admin children（users 之后）加：

```ts
{ path: 'health', name: 'AdminHealth',
  component: () => import('@/views/admin/HealthView.vue'),
  meta: { title: '节点健康', icon: 'Connection' } }
```

### 8. 修改 `web/src/config/menu.ts`（+1 行）

工作台 children「用户管理」后加：`{ path: '/admin/health', title: '节点健康', icon: 'Connection', roles: ['admin'] }`

## 边界处理

| 场景 | 行为 |
|---|---|
| Redis 不可用 | 心跳失败仅记日志；SystemHealth 降级为本地单节点 + warning，仍 200 |
| worker-only 节点 | main 无条件启动 HTTP + 心跳，自然被聚合看见 |
| 节点 crash | HDEL 不执行 → 30s 后 offline（保留故障痕迹），1h 后 List 清理，key 24h TTL 兜底 |
| 优雅关闭 | SetReady(false) → hbCancel（Deregister）→ worker/sub/http → <-hbDone → redis close |
| 无 Redis 配置 | registry 为 nil，页面单节点展示 |

## 验证（docker-compose dev，联调优先）

1. `docker restart drill-backend-dev`（go run 会重编译），日志确认心跳启动；`redis-cli HGETALL drill:health:nodes` 应见 node-1 的 JSON + hb 字段
2. 登录取 token 后 `curl http://localhost:8080/api/v1/system/health -H "Authorization: Bearer $TOKEN"`，校验 overall/nodes 结构；非 admin token 期望 403
3. 拨测 offline：`docker kill -s KILL drill-backend-dev` 后页面刷新，30s 后 node-1 转 offline；正常 restart 则条目消失（HDEL 生效）
4. 浏览器（前端 Vite 改动刷新即生效）：admin 侧边栏出现「节点健康」；页面指标随 10s 刷新变化、开关可暂停；executor 直接访问 URL → 403 跳转

## 改动量估算

后端 4 文件（1 新建 ~80 行、handler +160、main +45、router +1）；前端 5 文件（2 新建 ~290 行、api/index +1、router +7、menu +1）。合计约 9 文件 580 行。
