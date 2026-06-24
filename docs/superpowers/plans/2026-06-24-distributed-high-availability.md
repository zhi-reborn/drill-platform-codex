# Distributed High Availability Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Convert the current single-process flow runtime into Active-Active API/WebSocket nodes backed by a durable MySQL command queue and a Redis-elected singleton flow Worker.

**Architecture:** MySQL remains the source of truth for flow state, commands, logs, and notifications. Redis provides a fenced Worker lease plus Pub/Sub fan-out; API nodes submit and briefly await commands, while only the elected Worker rebuilds flow state and executes mutations.

**Tech Stack:** Go 1.23, Gin, GORM, MySQL 8, Redis 7/go-redis v9, Vue 3, Axios, Nginx, Docker Compose

---

## File map

New focused backend units:

- `internal/domain/entity/flow_command.go`: durable command model and constants.
- `internal/repository/flow_command.go`: command creation, lookup, claim, lease recovery, and completion.
- `internal/infrastructure/redis/lease.go`: compare-token Worker lease.
- `internal/infrastructure/events/redis_bus.go`: Redis Pub/Sub transport.
- `internal/service/flow_command.go`: API-facing submit/wait/query service.
- `internal/service/flow_command_executor.go`: maps durable commands to existing flow operations.
- `internal/worker/worker.go`: election, recovery, claim loop, and shutdown.
- `internal/api/handler/flowcommand/handler.go`: command status endpoint and shared mutation response.
- `internal/api/handler/health/handler.go`: liveness/readiness endpoints.
- `internal/pkg/appconfig/config.go`: YAML plus environment override loading.

Existing files remain responsible for their current domains. Handler changes are restricted to replacing direct mutation calls with command submission. Existing flow engine algorithms stay in `internal/pkg/flowengine`.

### Task 1: Add the durable command schema

**Files:**

- Create: `internal/domain/entity/flow_command.go`
- Create: `scripts/migration/2026-06-24-add-flow-command.sql`
- Modify: `scripts/init-db.sql`
- Modify: `internal/repository/database.go`
- Test: `internal/domain/entity/flow_command_test.go`

- [ ] **Step 1: Write the failing entity test**

```go
func TestFlowCommandTableAndTerminalStatus(t *testing.T) {
	cmd := entity.FlowCommand{Status: entity.FlowCommandSucceeded}
	if got := cmd.TableName(); got != "drill_flow_command" {
		t.Fatalf("TableName() = %q", got)
	}
	if !cmd.IsTerminal() {
		t.Fatal("succeeded command must be terminal")
	}
}
```

- [ ] **Step 2: Run the test and verify RED**

Run:

```bash
go test ./internal/domain/entity -run TestFlowCommandTableAndTerminalStatus -count=1
```

Expected: FAIL because `FlowCommand` does not exist.

- [ ] **Step 3: Add the command entity**

Implement these exact public fields and constants:

```go
type FlowCommandStatus string

const (
	FlowCommandPending    FlowCommandStatus = "pending"
	FlowCommandProcessing FlowCommandStatus = "processing"
	FlowCommandSucceeded  FlowCommandStatus = "succeeded"
	FlowCommandFailed     FlowCommandStatus = "failed"
)

type FlowCommand struct {
	ID              uint64            `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	CommandType     string            `gorm:"type:varchar(64);not null;column:command_type" json:"command_type"`
	DrillInstanceID uint64            `gorm:"type:bigint unsigned;not null;column:drill_instance_id" json:"drill_instance_id"`
	StepInstanceID  *uint64           `gorm:"type:bigint unsigned;column:step_instance_id" json:"step_instance_id,omitempty"`
	OperatorID      uint64            `gorm:"type:bigint unsigned;not null;column:operator_id" json:"operator_id"`
	IdempotencyKey  string            `gorm:"type:varchar(128);not null;uniqueIndex:uk_flow_command_idempotency;column:idempotency_key" json:"idempotency_key"`
	Payload         string            `gorm:"type:json;not null;column:payload" json:"payload"`
	Status          FlowCommandStatus `gorm:"type:varchar(20);not null;index:idx_flow_command_pending,priority:1;column:status" json:"status"`
	WorkerID        *string           `gorm:"type:varchar(128);column:worker_id" json:"worker_id,omitempty"`
	LeaseUntil      *time.Time        `gorm:"column:lease_until;index:idx_flow_command_lease,priority:2" json:"lease_until,omitempty"`
	Attempts        int               `gorm:"not null;default:0;column:attempts" json:"attempts"`
	Result          *string           `gorm:"type:json;column:result" json:"result,omitempty"`
	ErrorCode       *string           `gorm:"type:varchar(64);column:error_code" json:"error_code,omitempty"`
	ErrorMessage    *string           `gorm:"type:varchar(500);column:error_message" json:"error_message,omitempty"`
	CreatedAt       time.Time         `gorm:"column:created_at;autoCreateTime;index:idx_flow_command_pending,priority:2" json:"created_at"`
	StartedAt       *time.Time        `gorm:"column:started_at" json:"started_at,omitempty"`
	FinishedAt      *time.Time        `gorm:"column:finished_at" json:"finished_at,omitempty"`
	UpdatedAt       time.Time         `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}
```

Add nullable `CommandID *uint64` fields to `DrillInstanceLog` and `Notification`, with unique indexes `(command_id, action, task_instance_id)` and `(command_id, user_id, type, step_id)`.

- [ ] **Step 4: Add an idempotent MySQL migration**

Create the table with the indexes from the approved design, then add `command_id` to both side-effect tables. Use `INFORMATION_SCHEMA.COLUMNS` checks and prepared statements so rerunning the migration succeeds.

Update `scripts/init-db.sql` to create the same final schema on a fresh database. Add `FlowCommand` and `Notification` to `AutoMigrate`.

- [ ] **Step 5: Run tests and schema lint**

```bash
go test ./internal/domain/entity -count=1
git diff --check
```

Expected: PASS and no whitespace errors.

- [ ] **Step 6: Commit**

```bash
git add internal/domain/entity/flow_command.go internal/domain/entity/flow_command_test.go internal/domain/entity/log.go internal/domain/entity/notification.go internal/repository/database.go scripts/init-db.sql scripts/migration/2026-06-24-add-flow-command.sql
git commit -m "feat: add durable flow command schema"
```

### Task 2: Implement idempotent command persistence

**Files:**

- Create: `internal/repository/flow_command.go`
- Create: `internal/repository/flow_command_test.go`
- Modify: `internal/repository/database.go`

- [ ] **Step 1: Write failing repository tests**

Use an in-memory SQLite database for creation and lookup:

```go
func TestFlowCommandRepoCreateOrGetUsesIdempotencyKey(t *testing.T) {
	repo, cleanup := setupFlowCommandRepoTest(t)
	defer cleanup()

	first, created, err := repo.CreateOrGet(&entity.FlowCommand{
		CommandType: "start_drill", DrillInstanceID: 7, OperatorID: 3,
		IdempotencyKey: "same-key", Payload: `{}`, Status: entity.FlowCommandPending,
	})
	if err != nil || !created {
		t.Fatalf("first create: created=%v err=%v", created, err)
	}
	second, created, err := repo.CreateOrGet(&entity.FlowCommand{
		CommandType: "start_drill", DrillInstanceID: 7, OperatorID: 3,
		IdempotencyKey: "same-key", Payload: `{}`, Status: entity.FlowCommandPending,
	})
	if err != nil || created || second.ID != first.ID {
		t.Fatalf("duplicate create returned a second command")
	}
}
```

Also test `MarkSucceeded`, `MarkFailed`, and lookup restricted by operator ID.

- [ ] **Step 2: Verify RED**

```bash
go test ./internal/repository -run FlowCommand -count=1
```

Expected: FAIL because `FlowCommandRepo` is missing.

- [ ] **Step 3: Implement the repository**

Expose:

```go
type FlowCommandRepo struct{ db *gorm.DB }

func NewFlowCommandRepo(db ...*gorm.DB) *FlowCommandRepo
func (r *FlowCommandRepo) CreateOrGet(cmd *entity.FlowCommand) (*entity.FlowCommand, bool, error)
func (r *FlowCommandRepo) FindByID(id uint64) (*entity.FlowCommand, error)
func (r *FlowCommandRepo) FindByIDForOperator(id, operatorID uint64) (*entity.FlowCommand, error)
func (r *FlowCommandRepo) MarkSucceeded(id uint64, result any) error
func (r *FlowCommandRepo) MarkFailed(id uint64, code, message string) error
func (r *FlowCommandRepo) RequeueExpired(now time.Time) (int64, error)
```

`CreateOrGet` must treat only duplicate `idempotency_key` as the existing-command path; other insert failures must be returned.

- [ ] **Step 4: Add MySQL-only claim operations**

Implement:

```go
func (r *FlowCommandRepo) ClaimNext(ctx context.Context, workerID string, lease time.Duration) (*entity.FlowCommand, error)
func (r *FlowCommandRepo) ExtendLease(ctx context.Context, id uint64, workerID string, until time.Time) (bool, error)
```

`ClaimNext` must use a transaction and:

```sql
SELECT id
FROM drill_flow_command
WHERE status = 'pending'
ORDER BY created_at, id
LIMIT 1
FOR UPDATE SKIP LOCKED
```

Then update the selected row to `processing`, set `worker_id`, increment `attempts`, and assign `lease_until`.

- [ ] **Step 5: Run tests**

```bash
go test ./internal/repository -run FlowCommand -count=1
go test ./internal/repository ./internal/domain/entity -count=1
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/repository/flow_command.go internal/repository/flow_command_test.go internal/repository/database.go
git commit -m "feat: persist idempotent flow commands"
```

### Task 3: Add a fenced Redis Worker lease

**Files:**

- Create: `internal/infrastructure/redis/lease.go`
- Create: `internal/infrastructure/redis/lease_test.go`
- Modify: `internal/infrastructure/redis/client.go`

- [ ] **Step 1: Write failing lease tests**

Test against a small fake implementing `SetNX` and `Eval`:

```go
func TestLeaseCannotRenewOrReleaseAnotherToken(t *testing.T) {
	store := newFakeLeaseStore()
	first := redisinfra.NewLease(store, "leader", "worker-a", time.Second)
	second := redisinfra.NewLease(store, "leader", "worker-b", time.Second)

	if ok, _ := first.Acquire(context.Background()); !ok {
		t.Fatal("first lease not acquired")
	}
	if ok, _ := second.Renew(context.Background()); ok {
		t.Fatal("second token renewed first lease")
	}
	if ok, _ := second.Release(context.Background()); ok {
		t.Fatal("second token released first lease")
	}
}
```

- [ ] **Step 2: Verify RED**

```bash
go test ./internal/infrastructure/redis -run Lease -count=1
```

Expected: FAIL because `Lease` is missing.

- [ ] **Step 3: Implement compare-token lease operations**

Use a random UUID token in the stored value:

```go
type Lease struct {
	store LeaseStore
	key   string
	value string
	ttl   time.Duration
}
```

Acquire with `SET NX PX`. Renew and release with Lua scripts:

```lua
if redis.call("GET", KEYS[1]) == ARGV[1] then
  return redis.call("PEXPIRE", KEYS[1], ARGV[2])
end
return 0
```

```lua
if redis.call("GET", KEYS[1]) == ARGV[1] then
  return redis.call("DEL", KEYS[1])
end
return 0
```

Expose `Ping(ctx)`, `SetNX`, `Eval`, and `Raw()` on the Redis client instead of relying on the package-global context.

- [ ] **Step 4: Run tests**

```bash
go test ./internal/infrastructure/redis -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/infrastructure/redis/client.go internal/infrastructure/redis/lease.go internal/infrastructure/redis/lease_test.go
git commit -m "feat: add fenced redis worker lease"
```

### Task 4: Route WebSocket events through Redis Pub/Sub

**Files:**

- Create: `internal/infrastructure/events/event.go`
- Create: `internal/infrastructure/events/redis_bus.go`
- Create: `internal/infrastructure/events/redis_bus_test.go`
- Modify: `internal/infrastructure/websocket/manager.go`
- Modify: `internal/infrastructure/websocket/broadcast.go`
- Modify: `internal/infrastructure/redis/client.go`

- [ ] **Step 1: Write a failing cross-manager event test**

Create two WebSocket managers with recording clients, publish one drill event through a fake bus, and assert both subscribers receive the same event while only matching drill clients are notified.

- [ ] **Step 2: Verify RED**

```bash
go test ./internal/infrastructure/events ./internal/infrastructure/websocket -run Event -count=1
```

Expected: FAIL because the event transport is missing.

- [ ] **Step 3: Implement the transport-neutral event**

```go
type Event struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	DrillID   uint64          `json:"drill_id,omitempty"`
	UserID    uint64          `json:"user_id,omitempty"`
	Payload   json.RawMessage `json:"payload"`
	CreatedAt time.Time       `json:"created_at"`
}

type Publisher interface {
	Publish(context.Context, Event) error
}

type Subscriber interface {
	Subscribe(context.Context, func(Event)) error
	Healthy() bool
}
```

- [ ] **Step 4: Implement Redis Pub/Sub**

Publish JSON to `drill:events`. Subscribe in a reconnecting loop with bounded backoff. Mark the subscriber unhealthy between connection loss and successful resubscription.

Add `Manager.DeliverEvent(event)` that maps drill events to `BroadcastToDrillRaw` and user events to `BroadcastToUserRaw`. Existing `SendStepChange`, `SendDrillStatus`, and `SendTaskUpdate` become event constructors used by the Worker publisher rather than direct cross-node delivery.

- [ ] **Step 5: Run tests**

```bash
go test ./internal/infrastructure/events ./internal/infrastructure/websocket -count=1
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/infrastructure/events internal/infrastructure/websocket internal/infrastructure/redis/client.go
git commit -m "feat: distribute websocket events through redis"
```

### Task 5: Build the command submission and status API

**Files:**

- Create: `internal/service/flow_command.go`
- Create: `internal/service/flow_command_test.go`
- Create: `internal/api/handler/flowcommand/handler.go`
- Create: `internal/api/handler/flowcommand/handler_test.go`
- Modify: `internal/pkg/response/response.go`
- Modify: `internal/api/router/router.go`
- Modify: `internal/service/service.go`

- [ ] **Step 1: Write failing service tests**

Cover:

- supplied idempotency key is preserved;
- missing key generates one;
- duplicate submission returns the existing command;
- terminal commands return immediately;
- pending commands return after the configured wait timeout.

Use a 10 ms timeout in tests.

- [ ] **Step 2: Verify RED**

```bash
go test ./internal/service -run FlowCommand -count=1
```

Expected: FAIL because `FlowCommandService` is missing.

- [ ] **Step 3: Implement submit, wait, and query**

```go
type SubmitCommandRequest struct {
	CommandType     string
	DrillInstanceID uint64
	StepInstanceID  *uint64
	OperatorID      uint64
	IdempotencyKey  string
	Payload         any
}

type SubmitCommandResult struct {
	Command *entity.FlowCommand
	Pending bool
}
```

Poll MySQL every 50 ms until terminal, timeout, or request cancellation. This wait is an HTTP convenience only; command durability must not depend on the request remaining connected.

- [ ] **Step 4: Add HTTP response semantics**

Add:

```go
func Accepted(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusAccepted, Response{Code: CodeSuccess, Message: message, Data: data})
}
```

Add `GET /api/v1/flow-commands/:id`. Restrict normal users to their own commands; allow director/admin access after verifying the command's drill is visible under existing authorization rules.

- [ ] **Step 5: Test handler status codes**

Assert completed commands produce `200`, pending commands produce `202`, and querying another user's command produces `404` to avoid leaking IDs.

```bash
go test ./internal/service ./internal/api/handler/flowcommand -count=1
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/service/flow_command.go internal/service/flow_command_test.go internal/api/handler/flowcommand internal/pkg/response/response.go internal/api/router/router.go internal/service/service.go
git commit -m "feat: expose durable flow command api"
```

### Task 6: Move every flow mutation behind command submission

**Files:**

- Modify: `internal/api/handler/drill/handler.go`
- Modify: `internal/api/handler/task/handler.go`
- Modify: `internal/api/handler/drill/handler_test.go`
- Create: `internal/api/handler/task/handler_test.go`
- Modify: `internal/api/router/router.go`

- [ ] **Step 1: Add failing handler tests**

For each mutation family, inject a fake `FlowCommandService` and assert handlers submit the expected type and payload without calling `DrillService.Engine()`:

```text
start_drill pause_drill resume_drill terminate_drill
start_step complete_step report_issue skip_step
force_complete_step resume_task assign_step update_step_info
```

Test both director routes and executor task routes.

- [ ] **Step 2: Verify RED**

```bash
go test ./internal/api/handler/drill ./internal/api/handler/task -run Command -count=1
```

Expected: FAIL because handlers still execute mutations directly.

- [ ] **Step 3: Replace direct mutation calls**

Handlers continue performing authentication, permission checks, route ID parsing, and JSON validation. They then call one shared helper:

```go
result, err := h.commandService.SubmitAndWait(c.Request.Context(), service.SubmitCommandRequest{
	CommandType:     commandType,
	DrillInstanceID: drillID,
	StepInstanceID:  stepID,
	OperatorID:      middleware.GetUserID(c),
	IdempotencyKey:  c.GetHeader("Idempotency-Key"),
	Payload:         payload,
})
```

Return `200` for succeeded, the saved business error for failed, and `202` for pending. Include `Idempotency-Key` in every mutation response header.

Do not queue template CRUD, drill creation, drill deletion, notification read state, or report export; they are not flow-engine mutations.

- [ ] **Step 4: Run handler regressions**

```bash
go test ./internal/api/handler/drill ./internal/api/handler/task -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/api/handler/drill internal/api/handler/task internal/api/router/router.go
git commit -m "refactor: submit flow mutations as commands"
```

### Task 7: Implement the singleton Worker runtime

**Files:**

- Create: `internal/worker/status.go`
- Create: `internal/worker/worker.go`
- Create: `internal/worker/worker_test.go`
- Modify: `internal/service/service.go`

- [ ] **Step 1: Write failing election-loop tests**

With fake lease and fake command repo, verify:

- standby does not claim commands;
- acquiring leadership invokes recovery before claim;
- renewal failure stops new claims;
- shutdown releases only the current token.

- [ ] **Step 2: Verify RED**

```bash
go test ./internal/worker -count=1
```

Expected: FAIL because the Worker package is missing.

- [ ] **Step 3: Implement Worker states**

```go
type Status string

const (
	StatusStandby     Status = "standby"
	StatusRecovering  Status = "recovering"
	StatusLeaderReady Status = "leader-ready"
	StatusStopping    Status = "stopping"
)
```

`Run(ctx)` performs:

1. lease acquisition loop;
2. `Recover(ctx)` after acquisition;
3. expired command requeue;
4. claim/execute loop;
5. renewal ticker;
6. immediate demotion on renewal failure.

Use 15 s lease TTL, 5 s renewal, 60 s command lease, and 500 ms idle polling as defaults supplied by config rather than constants inside the loop.

- [ ] **Step 4: Run tests**

```bash
go test ./internal/worker -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/worker internal/service/service.go
git commit -m "feat: add elected flow worker runtime"
```

### Task 8: Execute commands transactionally and idempotently

**Files:**

- Create: `internal/service/flow_command_executor.go`
- Create: `internal/service/flow_command_executor_test.go`
- Modify: `internal/service/drill.go`
- Modify: `internal/service/task.go`
- Modify: `internal/service/flow_adapter.go`
- Modify: `internal/repository/drill.go`
- Modify: `internal/repository/step.go`
- Modify: `internal/repository/notification.go`

- [ ] **Step 1: Write failing executor tests**

Using SQLite for transaction behavior, cover:

- completing a `running` step changes it once;
- replaying the same command returns success without a second log or notification;
- completing a `pending` step fails with `invalid_status`;
- different command types decode their typed payloads;
- event publication occurs only after transaction commit.

- [ ] **Step 2: Verify RED**

```bash
go test ./internal/service -run FlowCommandExecutor -count=1
```

Expected: FAIL because the executor is missing.

- [ ] **Step 3: Add a database-backed executor boundary**

```go
type FlowCommandExecutor struct {
	db        *gorm.DB
	commands  *repository.FlowCommandRepo
	drills    *DrillService
	tasks     *TaskService
	publisher events.Publisher
	leader    LeaderGuard
}

func (e *FlowCommandExecutor) Execute(ctx context.Context, cmd *entity.FlowCommand) error
```

Before each transaction, verify the Worker still owns the Redis lease. Obtain an演练-level MySQL named lock using a dedicated `*sql.Conn`:

```sql
SELECT GET_LOCK(CONCAT('drill-flow:', ?), 5)
```

Begin the GORM transaction on that same connection, execute exactly one command, mark the command terminal, commit, release the named lock, and then publish collected events.

- [ ] **Step 4: Convert state writes to conditional transitions**

Add repository methods such as:

```go
func (r *StepRepo) TransitionStatus(tx *gorm.DB, id uint64, from []string, to string, updates map[string]any) (bool, error)
func (r *DrillRepo) TransitionStatus(tx *gorm.DB, id uint64, from []string, to string, updates map[string]any) (bool, error)
```

Every command must:

- treat target state as idempotent success;
- reject unrelated states with a stable error code;
- create logs/notifications with `command_id`;
- update progress in the same transaction;
- collect, but not publish, WebSocket events until commit.

Move existing mutation bodies into transaction-aware service methods. Remove handler-only database writes for skip, assign, and update-step-info.

- [ ] **Step 5: Run service and flow engine tests**

```bash
go test ./internal/service ./internal/repository ./internal/pkg/flowengine -count=1
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/service/flow_command_executor.go internal/service/flow_command_executor_test.go internal/service/drill.go internal/service/task.go internal/service/flow_adapter.go internal/repository/drill.go internal/repository/step.go internal/repository/notification.go
git commit -m "feat: execute flow commands transactionally"
```

### Task 9: Recover running flows and durable timeouts

**Files:**

- Create: `internal/service/flow_recovery.go`
- Create: `internal/service/flow_recovery_test.go`
- Modify: `internal/pkg/flowengine/timeout.go`
- Modify: `internal/service/drill.go`
- Modify: `internal/worker/worker.go`

- [ ] **Step 1: Write failing recovery tests**

Persist a running drill with:

- one future `timeout_at`;
- one expired `timeout_at`;
- one completed predecessor and pending successor.

Assert recovery rebuilds the engine, registers only the future timeout, and enqueues an internal timeout command for the expired step.

- [ ] **Step 2: Verify RED**

```bash
go test ./internal/service -run FlowRecovery -count=1
```

Expected: FAIL because recovery orchestration is missing.

- [ ] **Step 3: Implement authoritative recovery**

`FlowRecovery.RecoverAll(ctx)` loads `running` and `paused` drills from MySQL and calls a side-effect-free reconstruction method. Reconstruction must not create logs, notifications, or WebSocket events.

The timeout scheduler remains in the elected Worker only. Its callback submits a deterministic internal command with idempotency key:

```text
timeout:<drill-id>:<step-id>:<timeout-unix>
```

Do not execute timeout effects directly from the ticker.

- [ ] **Step 4: Run tests**

```bash
go test ./internal/service ./internal/pkg/flowengine ./internal/worker -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/service/flow_recovery.go internal/service/flow_recovery_test.go internal/pkg/flowengine/timeout.go internal/service/drill.go internal/worker/worker.go
git commit -m "feat: recover flows and durable timeouts"
```

### Task 10: Add configuration, readiness, and graceful shutdown

**Files:**

- Create: `internal/pkg/appconfig/config.go`
- Create: `internal/pkg/appconfig/config_test.go`
- Create: `internal/api/handler/health/handler.go`
- Create: `internal/api/handler/health/handler_test.go`
- Modify: `cmd/server/main.go`
- Modify: `internal/api/router/router.go`
- Modify: `internal/repository/database.go`
- Modify: `internal/pkg/loginlog/logger.go`
- Modify: `configs/config.yaml`

- [ ] **Step 1: Write failing config and readiness tests**

Verify environment variables override YAML and `/ready` returns:

- `200` when MySQL, Redis, and subscriber are healthy;
- `503` when either dependency is unavailable;
- `200` for a healthy standby Worker.

- [ ] **Step 2: Verify RED**

```bash
go test ./internal/pkg/appconfig ./internal/api/handler/health -count=1
```

Expected: FAIL because the packages are missing.

- [ ] **Step 3: Centralize configuration**

Support the approved variables:

```text
APP_ROLE INSTANCE_ID DATABASE_HOST DATABASE_PORT DATABASE_USER
DATABASE_PASSWORD DATABASE_NAME REDIS_ADDR REDIS_PASSWORD JWT_SECRET
PUBLIC_BASE_URL CAS_PUBLIC_URL CAS_SERVICE_URL WORKER_LEASE_TTL
WORKER_RENEW_INTERVAL COMMAND_WAIT_TIMEOUT LOGIN_LOG_FILE
```

Validate `APP_ROLE` as `api`, `worker`, or `all`; reject an empty production JWT secret and an empty instance ID for Worker roles.

- [ ] **Step 4: Add health endpoints**

Keep `/health` as a compatibility alias for `/live`.

Return:

```json
{
  "status": "ready",
  "role": "all",
  "instance_id": "node-a",
  "worker_status": "standby"
}
```

Readiness checks must use context timeouts and must not mutate dependency state.

- [ ] **Step 5: Replace `r.Run` with managed `http.Server`**

Use `signal.NotifyContext`, mark readiness false before shutdown, stop the Worker, shut down HTTP with a timeout, close subscriptions/WebSockets, then close Redis and MySQL.

Change login logging so an empty `LOGIN_LOG_FILE` writes stdout only.

- [ ] **Step 6: Run tests**

```bash
go test ./internal/pkg/appconfig ./internal/api/handler/health ./cmd/server -count=1
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/pkg/appconfig internal/api/handler/health cmd/server/main.go internal/api/router/router.go internal/repository/database.go internal/pkg/loginlog/logger.go configs/config.yaml
git commit -m "feat: add readiness and graceful lifecycle"
```

### Task 11: Make the frontend command-aware

**Files:**

- Create: `web/src/types/flowCommand.ts`
- Create: `web/src/api/modules/flowCommand.ts`
- Create: `web/src/api/idempotency.ts`
- Create: `web/src/api/idempotency.test.ts`
- Modify: `web/src/api/request.ts`
- Modify: `web/src/api/modules/drill.ts`
- Modify: `web/src/api/modules/task.ts`
- Modify: `web/vite.config.ts`
- Modify: `web/package.json`
- Modify: `web/package-lock.json`

- [ ] **Step 1: Write a failing idempotency test**

Using the existing frontend test runner, verify one generated key is reused when a mutation retry receives a network failure, while a new user action receives a new key.

- [ ] **Step 2: Verify RED**

Install Vitest and add `"test": "vitest run"` to `web/package.json`:

```bash
npm --prefix web install --save-dev vitest
```

Add this to the object returned from `defineConfig` in `web/vite.config.ts`:

```ts
test: {
  environment: 'node',
  include: ['src/**/*.test.ts'],
},
```

Then run:

```bash
npm --prefix web run test -- idempotency.test.ts
```

Expected: FAIL because the helper is missing.

- [ ] **Step 3: Implement command-aware requests**

Define:

```ts
export interface FlowCommand {
  id: number
  status: 'pending' | 'processing' | 'succeeded' | 'failed'
  result?: unknown
  error_code?: string
  error_message?: string
}
```

Mutation helpers generate `crypto.randomUUID()`, send it in `Idempotency-Key`, and return either the completed result or a pending command. Add `flowCommandApi.get(id)`.

Do not globally attach idempotency keys to GET requests or unrelated POST endpoints such as login.

- [ ] **Step 4: Preserve current UI behavior**

Existing views may continue displaying their current success message. For a `202`, show “操作已受理” and rely on WebSocket plus existing detail refresh. If the command later reports `failed`, surface `error_message`.

- [ ] **Step 5: Run frontend verification**

```bash
npm --prefix web run test
npm --prefix web run build
```

Expected: tests and TypeScript/Vite build PASS.

- [ ] **Step 6: Commit**

```bash
git add web/src/types/flowCommand.ts web/src/api/modules/flowCommand.ts web/src/api/idempotency.ts web/src/api/idempotency.test.ts web/src/api/request.ts web/src/api/modules/drill.ts web/src/api/modules/task.ts web/vite.config.ts web/package.json web/package-lock.json
git commit -m "feat: handle asynchronous flow commands in web client"
```

### Task 12: Add multi-node deployment configuration

**Files:**

- Modify: `nginx/nginx.conf`
- Modify: `Dockerfile-dev`
- Create: `Dockerfile`
- Create: `docker-compose.ha.yml`
- Create: `.env.ha.example`
- Create: `docs/deployment/high-availability.md`

- [ ] **Step 1: Add a config assertion script/test**

Add a small Go test or shell-safe validation that checks:

- Nginx has at least two backend servers;
- `/api/` and `/ws/` target the same upstream;
- WebSocket targets port `8080`;
- `proxy_pass` preserves `/api/v1`.

- [ ] **Step 2: Verify RED**

Run the new validation and confirm it fails against the current `backend:8080`/`backend:8081` split.

- [ ] **Step 3: Update Nginx**

Use:

```nginx
upstream drill_backend {
    least_conn;
    server backend-a:8080 max_fails=3 fail_timeout=10s;
    server backend-b:8080 max_fails=3 fail_timeout=10s;
    keepalive 32;
}
```

Both locations use `proxy_pass http://drill_backend;` without a trailing URI, preserving the original path. Set WebSocket read timeout to 300 seconds.

- [ ] **Step 4: Add a production image and HA compose topology**

The production Dockerfile builds the Go binary and frontend assets in separate stages, then runs a minimal image. `docker-compose.ha.yml` starts:

- shared MySQL and Redis for local HA simulation;
- `backend-a` and `backend-b` with unique `INSTANCE_ID`;
- Nginx in front of both.

This compose file demonstrates application failover only; the deployment guide must state that production MySQL and Redis require their own HA products.

- [ ] **Step 5: Validate configuration**

```bash
docker compose -f docker-compose.ha.yml config
docker run --rm -v "$PWD/nginx/nginx.conf:/etc/nginx/nginx.conf:ro" nginx:alpine nginx -t
```

Expected: both commands PASS.

- [ ] **Step 6: Commit**

```bash
git add nginx/nginx.conf Dockerfile Dockerfile-dev docker-compose.ha.yml .env.ha.example docs/deployment/high-availability.md
git commit -m "feat: add multi-node deployment topology"
```

### Task 13: Add MySQL/Redis integration and failover tests

**Files:**

- Create: `internal/integration/flow_command_mysql_test.go`
- Create: `internal/integration/worker_failover_test.go`
- Create: `internal/integration/websocket_pubsub_test.go`
- Create: `scripts/test-ha.sh`
- Modify: `docker-compose.ha.yml`

- [ ] **Step 1: Write tagged MySQL integration tests**

Use `//go:build integration`. Verify:

- concurrent `CreateOrGet` returns one command ID;
- two concurrent claimers cannot claim the same command;
- expired processing commands are reclaimed;
- MySQL named lock serializes commands for one drill;
- commands for two drills can proceed independently.

- [ ] **Step 2: Run and verify failures before final wiring**

```bash
go test -tags=integration ./internal/integration -run 'Command|Lock' -count=1
```

Expected: tests fail until the compose dependencies and environment are supplied.

- [ ] **Step 3: Add two-node failover coverage**

Start two Workers against the same MySQL/Redis. Assert exactly one reaches `leader-ready`. Cancel it, wait for lease expiry, and assert the standby recovers a running drill and completes a pending command.

- [ ] **Step 4: Add cross-node WebSocket coverage**

Connect a WebSocket client to backend B, submit a command through backend A, and assert the event arrives at B. Disconnect/reconnect the client and verify the REST detail endpoint returns the committed state.

- [ ] **Step 5: Create the reproducible HA test script**

`scripts/test-ha.sh` must:

1. build and start `docker-compose.ha.yml`;
2. wait for both `/ready` endpoints;
3. run integration tests;
4. identify and stop the leader container;
5. verify command completion through the surviving node;
6. always collect logs and tear down using a shell trap.

- [ ] **Step 6: Run the complete verification**

```bash
go test ./... -count=1
npm --prefix web run test
npm --prefix web run build
./scripts/test-ha.sh
git diff --check
```

Expected: all unit tests, frontend checks, multi-node WebSocket test, and Worker failover test PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/integration scripts/test-ha.sh docker-compose.ha.yml
git commit -m "test: verify multi-node flow failover"
```

### Task 14: Final review and production handoff

**Files:**

- Modify only if verification reveals a defect directly related to this feature.

- [ ] **Step 1: Confirm no API node executes flow mutations directly**

```bash
rg -n 'Engine\\(\\)|ManualStartStep|DirectorCompleteStep|DirectorSkipStep|DirectorForceComplete|ActionResumeTask' internal/api/handler
```

Expected: no mutation call sites in handlers.

- [ ] **Step 2: Confirm all side effects are command-linked**

```bash
rg -n 'DrillInstanceLog\\{|Notification\\{' internal/service
```

Review every flow mutation path and confirm it assigns `CommandID`.

- [ ] **Step 3: Run race and static checks**

```bash
go test -race ./internal/worker ./internal/infrastructure/events ./internal/infrastructure/websocket ./internal/service -count=1
go vet ./...
```

Expected: PASS.

- [ ] **Step 4: Record operational evidence**

In `docs/deployment/high-availability.md`, record:

- the exact tested image/commit;
- Worker lease and takeover timings;
- `/ready` examples for leader and standby;
- the expected behavior while Redis is unavailable;
- rollback steps that disable command submission and return to one backend node.

- [ ] **Step 5: Request code review**

Use the `superpowers:requesting-code-review` skill, address only verified findings, then rerun the complete verification from Task 13.
