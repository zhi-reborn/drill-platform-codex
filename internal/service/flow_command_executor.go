package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"drill-platform/internal/domain/entity"
	"drill-platform/internal/infrastructure/events"
	"drill-platform/internal/pkg/flowengine"
	"drill-platform/internal/repository"
	"drill-platform/internal/worker"

	"gorm.io/gorm"
)

// CommandRepo abstracts the durable command persistence boundary used by the
// executor. FlowCommandRepo satisfies this interface via its fenced methods.
// All mutations are fenced by CommandOwnership so a stale worker cannot commit
// a command that has been re-claimed by a newer worker.
type CommandRepo interface {
	MarkSucceededFenced(ctx context.Context, id uint64, ownership repository.CommandOwnership, result any) error
	MarkFailedFenced(ctx context.Context, id uint64, ownership repository.CommandOwnership, code, message string) error
}

// acquireNamedLock acquires a MySQL named lock for the given drill ID within a
// transaction. On non-MySQL dialectors (e.g. SQLite in tests) it is a no-op.
// Returns a lock_timeout commandError when GET_LOCK returns 0 (timeout), so the
// caller can treat the failure as retryable. This is a package-level variable
// so tests can substitute a stub to verify call sites without a real MySQL.
var acquireNamedLock = func(tx *gorm.DB, drillID uint64, timeoutSeconds int) error {
	if tx.Dialector.Name() != "mysql" {
		return nil
	}
	name := fmt.Sprintf("drill:%d", drillID)
	var result int
	if err := tx.Raw("SELECT GET_LOCK(?, ?)", name, timeoutSeconds).Scan(&result).Error; err != nil {
		return err
	}
	if result != 1 {
		return &commandError{Code: "lock_timeout", Message: fmt.Sprintf("could not acquire named lock %s within %ds", name, timeoutSeconds)}
	}
	return nil
}

// releaseNamedLock releases a previously acquired MySQL named lock. On
// non-MySQL dialectors it is a no-op. Errors are ignored because the lock is
// also released when the session ends.
var releaseNamedLock = func(tx *gorm.DB, drillID uint64) error {
	if tx.Dialector.Name() != "mysql" {
		return nil
	}
	name := fmt.Sprintf("drill:%d", drillID)
	return tx.Exec("DO RELEASE_LOCK(?)", name).Error
}

// FlowCommandExecutor maps a durable FlowCommand to its transactional side
// effects. It implements worker.Executor. The ExecutionFence passed to Execute
// is the sole authority for committing results; there is no separate leader
// guard.
//
// Every mutation flows through a single DB transaction path: there are no
// production bypass branches that shortcut to DrillService/TaskService. This
// guarantees the domain state change, command result, command terminal status,
// and event records all commit atomically inside one transaction, with a
// per-drill MySQL named lock serializing concurrent mutations.
type FlowCommandExecutor struct {
	db        *gorm.DB
	commands  CommandRepo
	stepRepo  *repository.StepRepo
	drillRepo *repository.DrillRepo
	publisher events.Publisher
}

// NewFlowCommandExecutor constructs an executor with the given dependencies.
func NewFlowCommandExecutor(
	db *gorm.DB,
	commands CommandRepo,
	publisher events.Publisher,
) *FlowCommandExecutor {
	return &FlowCommandExecutor{
		db:        db,
		commands:  commands,
		stepRepo:  repository.NewStepRepo(),
		drillRepo: repository.NewDrillRepo(),
		publisher: publisher,
	}
}

// Execute dispatches a single FlowCommand to its typed handler. It is
// idempotent: terminal commands return nil immediately. The fence carries the
// worker's epoch and the command's lease_token; all mutations are fenced by
// these values so a stale worker cannot flip a command that has been
// re-claimed. The Worker ignores the returned error, so Execute marks the
// command terminal internally.
//
// Lock timeout is treated as retryable: the command is left in processing
// (not marked failed) so the lease expires and RequeueExpired can re-claim it.
func (e *FlowCommandExecutor) Execute(ctx context.Context, cmd *entity.FlowCommand, fence worker.ExecutionFence) error {
	if cmd.IsTerminal() {
		return nil
	}

	ownership := repository.CommandOwnership{
		WorkerID: fence.WorkerID,
		Epoch:    fence.WorkerEpoch,
		Token:    fence.LeaseToken,
	}

	execErr := e.dispatch(ctx, cmd, ownership)
	if execErr != nil {
		// lock_timeout is retryable: leave the command in processing so the
		// lease expires and RequeueExpired re-claims it. Marking it failed
		// would prevent retry.
		var ce *commandError
		if errors.As(execErr, &ce) && ce.Code == "lock_timeout" {
			return execErr
		}
		code, message := classifyError(execErr)
		_ = e.commands.MarkFailedFenced(ctx, cmd.ID, ownership, code, message)
		return execErr
	}

	// Dispatch paths that use a transaction (transitionDrillStatus,
	// transitionStepInTx, transitionStepFieldsInTx) already mark the command
	// succeeded in-transaction via markCommandSucceededInTx. If the command
	// is already terminal (transition path), the fenced WHERE clause matches
	// zero rows and returns ErrOwnershipLost, which we intentionally ignore.
	_ = e.commands.MarkSucceededFenced(ctx, cmd.ID, ownership, nil)
	return nil
}

func (e *FlowCommandExecutor) dispatch(ctx context.Context, cmd *entity.FlowCommand, ownership repository.CommandOwnership) error {
	switch cmd.CommandType {
	case "start_drill":
		return e.executeStartDrill(ctx, cmd, ownership)
	case "pause_drill":
		return e.executePauseDrill(ctx, cmd, ownership)
	case "resume_drill":
		return e.executeResumeDrill(ctx, cmd, ownership)
	case "terminate_drill":
		return e.executeTerminateDrill(ctx, cmd, ownership)
	case "start_step":
		return e.executeStartStep(ctx, cmd, ownership)
	case "complete_step":
		return e.executeCompleteStep(ctx, cmd, ownership)
	case "report_issue":
		return e.executeReportIssue(ctx, cmd, ownership)
	case "skip_step":
		return e.executeSkipStep(ctx, cmd, ownership)
	case "force_complete_step":
		return e.executeForceCompleteStep(ctx, cmd, ownership)
	case "resume_task":
		return e.executeResumeTask(ctx, cmd, ownership)
	case "assign_step":
		return e.executeAssignStep(ctx, cmd, ownership)
	case "update_step_info":
		return e.executeUpdateStepInfo(ctx, cmd, ownership)
	default:
		return &commandError{Code: "unknown_command", Message: fmt.Sprintf("unknown command type: %s", cmd.CommandType)}
	}
}

// commandError carries a stable error code for MarkFailed.
type commandError struct {
	Code    string
	Message string
}

func (e *commandError) Error() string { return e.Message }

func classifyError(err error) (string, string) {
	var ce *commandError
	if errors.As(err, &ce) {
		return ce.Code, ce.Message
	}
	if errors.Is(err, flowengine.ErrInvalidStatus) || errors.Is(err, flowengine.ErrStepNotActive) {
		return "invalid_status", err.Error()
	}
	if errors.Is(err, flowengine.ErrInstanceNotFound) {
		return "instance_not_found", err.Error()
	}
	if errors.Is(err, flowengine.ErrStepNotFound) {
		return "step_not_found", err.Error()
	}
	if errors.Is(err, flowengine.ErrInstanceNotRunning) {
		return "instance_not_running", err.Error()
	}
	if errors.Is(err, flowengine.ErrPreStepsNotDone) {
		return "pre_steps_not_done", err.Error()
	}
	return "internal_error", err.Error()
}

// --- Drill-level commands ---

func (e *FlowCommandExecutor) executeStartDrill(_ context.Context, cmd *entity.FlowCommand, ownership repository.CommandOwnership) error {
	return e.transitionDrillStatus(cmd, ownership, []string{"pending"}, "running")
}

func (e *FlowCommandExecutor) executePauseDrill(_ context.Context, cmd *entity.FlowCommand, ownership repository.CommandOwnership) error {
	return e.transitionDrillStatus(cmd, ownership, []string{"running"}, "paused")
}

func (e *FlowCommandExecutor) executeResumeDrill(_ context.Context, cmd *entity.FlowCommand, ownership repository.CommandOwnership) error {
	return e.transitionDrillStatus(cmd, ownership, []string{"paused"}, "running")
}

func (e *FlowCommandExecutor) executeTerminateDrill(_ context.Context, cmd *entity.FlowCommand, ownership repository.CommandOwnership) error {
	return e.transitionDrillStatus(cmd, ownership, []string{"running", "paused", "issue"}, "terminated")
}

func (e *FlowCommandExecutor) transitionDrillStatus(cmd *entity.FlowCommand, ownership repository.CommandOwnership, from []string, to string) error {
	now := time.Now()
	updates := map[string]any{}
	if to == "running" {
		updates["start_time"] = &now
	}
	if to == "terminated" || to == "completed" {
		updates["end_time"] = &now
	}

	db := e.db
	if db == nil {
		db = repository.DB
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := acquireNamedLock(tx, cmd.DrillInstanceID, namedLockTimeoutSeconds); err != nil {
			return err
		}
		defer func() { _ = releaseNamedLock(tx, cmd.DrillInstanceID) }()

		changed, err := e.drillRepo.TransitionStatus(tx, cmd.DrillInstanceID, from, to, updates)
		if err != nil {
			return err
		}
		if !changed {
			var drill entity.DrillInstance
			if err := tx.Select("status").First(&drill, cmd.DrillInstanceID).Error; err != nil {
				return err
			}
			if drill.Status == to {
				return markCommandSucceededInTx(tx, cmd.ID, ownership)
			}
			return &commandError{Code: "invalid_status", Message: fmt.Sprintf("drill status is %s, expected one of the allowed transitions", drill.Status)}
		}
		return markCommandSucceededInTx(tx, cmd.ID, ownership)
	})
}

// namedLockTimeoutSeconds is the MySQL GET_LOCK timeout. A short timeout keeps
// latency bounded while still tolerating brief contention; timeouts are
// classified as retryable so the command is re-claimed after lease expiry.
const namedLockTimeoutSeconds = 5

// --- Step-level commands ---

// CompleteStepPayload is the typed payload for complete_step commands.
type CompleteStepPayload struct {
	Remark string `json:"remark"`
}

// ReportIssuePayload is the typed payload for report_issue commands.
type ReportIssuePayload struct {
	IssueDesc string `json:"issue_desc"`
}

// SkipStepPayload is the typed payload for skip_step commands.
type SkipStepPayload struct {
	Reason string `json:"reason"`
}

// AssignStepPayload is the typed payload for assign_step commands.
type AssignStepPayload struct {
	AssigneeIDs []uint64 `json:"assignee_ids"`
}

// UpdateStepInfoPayload is the typed payload for update_step_info commands.
type UpdateStepInfoPayload struct {
	Remark string `json:"remark"`
}

func (e *FlowCommandExecutor) executeStartStep(ctx context.Context, cmd *entity.FlowCommand, ownership repository.CommandOwnership) error {
	if cmd.StepInstanceID == nil {
		return &commandError{Code: "missing_step", Message: "step_instance_id is required"}
	}
	return e.transitionStepInTx(cmd, ownership, []string{"pending"}, "running", map[string]any{
		"start_time": time.Now(),
	})
}

func (e *FlowCommandExecutor) executeCompleteStep(ctx context.Context, cmd *entity.FlowCommand, ownership repository.CommandOwnership) error {
	if cmd.StepInstanceID == nil {
		return &commandError{Code: "missing_step", Message: "step_instance_id is required"}
	}
	var payload CompleteStepPayload
	if err := decodePayload(cmd.Payload, &payload); err != nil {
		return err
	}

	return e.transitionStepInTx(cmd, ownership, []string{"running"}, "completed", map[string]any{
		"actual_operator": cmd.OperatorID,
		"end_time":        time.Now(),
		"remark":          payload.Remark,
	})
}

func (e *FlowCommandExecutor) executeReportIssue(ctx context.Context, cmd *entity.FlowCommand, ownership repository.CommandOwnership) error {
	if cmd.StepInstanceID == nil {
		return &commandError{Code: "missing_step", Message: "step_instance_id is required"}
	}
	var payload ReportIssuePayload
	if err := decodePayload(cmd.Payload, &payload); err != nil {
		return err
	}

	return e.transitionStepInTx(cmd, ownership, []string{"running"}, "issue", map[string]any{
		"issue_desc": payload.IssueDesc,
		"end_time":   time.Now(),
	})
}

func (e *FlowCommandExecutor) executeSkipStep(ctx context.Context, cmd *entity.FlowCommand, ownership repository.CommandOwnership) error {
	if cmd.StepInstanceID == nil {
		return &commandError{Code: "missing_step", Message: "step_instance_id is required"}
	}

	return e.transitionStepInTx(cmd, ownership, []string{"running", "pending"}, "skipped", map[string]any{
		"end_time": time.Now(),
	})
}

func (e *FlowCommandExecutor) executeForceCompleteStep(ctx context.Context, cmd *entity.FlowCommand, ownership repository.CommandOwnership) error {
	if cmd.StepInstanceID == nil {
		return &commandError{Code: "missing_step", Message: "step_instance_id is required"}
	}

	return e.transitionStepInTx(cmd, ownership, []string{"running", "pending", "issue"}, "completed", map[string]any{
		"actual_operator": cmd.OperatorID,
		"end_time":        time.Now(),
	})
}

func (e *FlowCommandExecutor) executeResumeTask(ctx context.Context, cmd *entity.FlowCommand, ownership repository.CommandOwnership) error {
	if cmd.StepInstanceID == nil {
		return &commandError{Code: "missing_step", Message: "step_instance_id is required"}
	}

	return e.transitionStepInTx(cmd, ownership, []string{"completed", "skipped", "timeout", "issue"}, "running", map[string]any{
		"start_time": time.Now(),
	})
}

func (e *FlowCommandExecutor) executeAssignStep(ctx context.Context, cmd *entity.FlowCommand, ownership repository.CommandOwnership) error {
	if cmd.StepInstanceID == nil {
		return &commandError{Code: "missing_step", Message: "step_instance_id is required"}
	}
	var payload AssignStepPayload
	if err := decodePayload(cmd.Payload, &payload); err != nil {
		return err
	}

	assigneeJSON := "[]"
	if len(payload.AssigneeIDs) > 0 {
		bytes, _ := json.Marshal(payload.AssigneeIDs)
		assigneeJSON = string(bytes)
	}

	step, err := e.loadStep(*cmd.StepInstanceID)
	if err != nil {
		return err
	}
	stepID := step.ID
	drillID := step.DrillInstanceID
	notif := &entity.Notification{
		UserID:  cmd.OperatorID,
		Type:    entity.NotificationTypeTaskAssigned,
		Title:   "步骤分配通知",
		Content: step.Name + " 分配已更新",
		DrillID: &drillID,
		StepID:  &stepID,
		IsRead:  false,
	}
	return e.transitionStepFieldsInTx(
		cmd,
		ownership,
		map[string]any{"assignee_ids": assigneeJSON},
		"assign",
		step.Name+" 分配已更新",
		notif,
	)
}

func (e *FlowCommandExecutor) executeUpdateStepInfo(ctx context.Context, cmd *entity.FlowCommand, ownership repository.CommandOwnership) error {
	if cmd.StepInstanceID == nil {
		return &commandError{Code: "missing_step", Message: "step_instance_id is required"}
	}
	var payload UpdateStepInfoPayload
	if err := decodePayload(cmd.Payload, &payload); err != nil {
		return err
	}

	step, err := e.loadStep(*cmd.StepInstanceID)
	if err != nil {
		return err
	}
	content := payload.Remark
	if content == "" {
		content = step.Name + " 信息已更新"
	}
	return e.transitionStepFieldsInTx(
		cmd,
		ownership,
		map[string]any{"remark": payload.Remark},
		"update_info",
		content,
		nil,
	)
}

// transitionStepInTx performs a conditional step status transition within a
// GORM transaction. On success it creates a log and notification carrying
// command_id, marks the command succeeded in the same transaction (fenced by
// ownership), commits, then publishes a WebSocket event. The transaction
// acquires a per-drill MySQL named lock at the start to serialize concurrent
// mutations to the same drill.
func (e *FlowCommandExecutor) transitionStepInTx(
	cmd *entity.FlowCommand,
	ownership repository.CommandOwnership,
	from []string,
	to string,
	updates map[string]any,
) error {
	db := e.db
	if db == nil {
		db = repository.DB
	}

	var collectedEvents []events.WSMessage

	err := db.Transaction(func(tx *gorm.DB) error {
		if err := acquireNamedLock(tx, cmd.DrillInstanceID, namedLockTimeoutSeconds); err != nil {
			return err
		}
		defer func() { _ = releaseNamedLock(tx, cmd.DrillInstanceID) }()

		changed, err := e.stepRepo.TransitionStatus(tx, *cmd.StepInstanceID, from, to, updates)
		if err != nil {
			return err
		}
		if !changed {
			var current entity.StepInstance
			if err := tx.Select("status").First(&current, *cmd.StepInstanceID).Error; err != nil {
				return err
			}
			if current.Status == to {
				return markCommandSucceededInTx(tx, cmd.ID, ownership)
			}
			return &commandError{Code: "invalid_status", Message: fmt.Sprintf("step status is %s, expected one of %v", current.Status, from)}
		}

		var step entity.StepInstance
		if err := tx.First(&step, *cmd.StepInstanceID).Error; err != nil {
			return err
		}

		cmdID := cmd.ID
		logEntry := e.buildStepLog(tx, cmd, &step, to)
		logEntry.CommandID = &cmdID
		if err := tx.Create(logEntry).Error; err != nil {
			return err
		}

		notif := e.buildStepNotification(cmd, &step, to)
		if notif != nil {
			notif.CommandID = &cmdID
			if err := tx.Create(notif).Error; err != nil {
				return err
			}
		}

		if err := markCommandSucceededInTx(tx, cmd.ID, ownership); err != nil {
			return err
		}

		collectedEvents = append(collectedEvents, events.NewWSMessage(
			fmt.Sprintf("step_%s", to),
			step.DrillInstanceID,
			0,
			json.RawMessage(cmd.Payload),
		))

		return nil
	})
	if err != nil {
		return err
	}

	e.publishCollected(collectedEvents)
	return nil
}

// markCommandSucceededInTx flips a FlowCommand to succeeded within the given
// transaction so the command status commits atomically with the business state.
// The fenced WHERE clause ensures a stale worker (one whose epoch/token/lease
// no longer match) cannot flip the command; zero rows affected returns
// ErrOwnershipLost, which rolls back the enclosing transaction.
func markCommandSucceededInTx(tx *gorm.DB, cmdID uint64, ownership repository.CommandOwnership) error {
	now := time.Now()
	res := tx.Model(&entity.FlowCommand{}).
		Where("id = ? AND status = ? AND worker_id = ? AND worker_epoch = ? AND lease_token = ? AND lease_until > ?",
			cmdID, entity.FlowCommandProcessing, ownership.WorkerID, ownership.Epoch, ownership.Token, now).
		Updates(map[string]any{
			"status":      entity.FlowCommandSucceeded,
			"finished_at": now,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return repository.ErrOwnershipLost
	}
	return nil
}

// transitionStepFieldsInTx updates step fields (without changing status) inside
// a GORM transaction. It creates a log and optional notification carrying
// command_id, marks the command succeeded (fenced by ownership), commits, then
// publishes an event. Used by assign_step and update_step_info. The transaction
// acquires a per-drill MySQL named lock at the start to serialize concurrent
// mutations to the same drill.
func (e *FlowCommandExecutor) transitionStepFieldsInTx(
	cmd *entity.FlowCommand,
	ownership repository.CommandOwnership,
	updates map[string]any,
	action string,
	content string,
	notif *entity.Notification,
) error {
	db := e.db
	if db == nil {
		db = repository.DB
	}

	var collectedEvents []events.WSMessage

	err := db.Transaction(func(tx *gorm.DB) error {
		if err := acquireNamedLock(tx, cmd.DrillInstanceID, namedLockTimeoutSeconds); err != nil {
			return err
		}
		defer func() { _ = releaseNamedLock(tx, cmd.DrillInstanceID) }()

		res := tx.Model(&entity.StepInstance{}).
			Where("id = ?", *cmd.StepInstanceID).
			Updates(updates)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return &commandError{Code: "step_not_found", Message: fmt.Sprintf("step %d not found", *cmd.StepInstanceID)}
		}

		var step entity.StepInstance
		if err := tx.First(&step, *cmd.StepInstanceID).Error; err != nil {
			return err
		}

		cmdID := cmd.ID
		stepID := step.ID
		logEntry := &entity.DrillInstanceLog{
			DrillInstanceID: step.DrillInstanceID,
			StepInstanceID:  &stepID,
			Action:          action,
			OperatorID:      cmd.OperatorID,
			OperatorName:    resolveOperatorName(tx, cmd.OperatorID),
			Content:         content,
			CommandID:       &cmdID,
		}
		if err := tx.Create(logEntry).Error; err != nil {
			return err
		}

		if notif != nil {
			notif.CommandID = &cmdID
			if err := tx.Create(notif).Error; err != nil {
				return err
			}
		}

		if err := markCommandSucceededInTx(tx, cmd.ID, ownership); err != nil {
			return err
		}

		collectedEvents = append(collectedEvents, events.NewWSMessage(
			fmt.Sprintf("step_%s", action),
			step.DrillInstanceID,
			0,
			json.RawMessage(cmd.Payload),
		))

		return nil
	})
	if err != nil {
		return err
	}

	e.publishCollected(collectedEvents)
	return nil
}

func (e *FlowCommandExecutor) buildStepLog(db *gorm.DB, cmd *entity.FlowCommand, step *entity.StepInstance, to string) *entity.DrillInstanceLog {
	action := to
	content := step.Name + " -> " + to
	switch to {
	case "completed":
		action = "complete"
		var payload CompleteStepPayload
		if json.Unmarshal([]byte(cmd.Payload), &payload) == nil && payload.Remark != "" {
			content = payload.Remark
		} else {
			content = step.Name + " 已完成"
		}
	case "issue":
		action = "issue"
		var payload ReportIssuePayload
		if json.Unmarshal([]byte(cmd.Payload), &payload) == nil && payload.IssueDesc != "" {
			content = payload.IssueDesc
		}
	case "skipped":
		action = "skip"
		content = "指挥员跳过步骤"
	}
	stepID := step.ID
	return &entity.DrillInstanceLog{
		DrillInstanceID: step.DrillInstanceID,
		StepInstanceID:  &stepID,
		Action:          action,
		OperatorID:      cmd.OperatorID,
		OperatorName:    resolveOperatorName(db, cmd.OperatorID),
		Content:         content,
	}
}

// resolveOperatorName looks up the operator's display name. Falls back to
// "executor" when the user cannot be resolved (e.g. userRepo unset, user
// missing, or the users table is unavailable in the current DB). The db
// argument should be the transaction handle when called inside a transaction
// to avoid connection starvation under limited pool sizes.
func resolveOperatorName(db *gorm.DB, operatorID uint64) string {
	if operatorID == 0 || db == nil {
		return "executor"
	}
	var user entity.User
	if err := db.Select("real_name").First(&user, operatorID).Error; err != nil {
		return "executor"
	}
	if user.RealName == "" {
		return "executor"
	}
	return user.RealName
}

func (e *FlowCommandExecutor) buildStepNotification(cmd *entity.FlowCommand, step *entity.StepInstance, to string) *entity.Notification {
	var notifType entity.NotificationType
	title := ""
	content := ""
	switch to {
	case "completed":
		notifType = entity.NotificationTypeStepComplete
		title = "步骤已完成"
		content = step.Name + " 已完成"
	case "issue":
		notifType = entity.NotificationTypeStepIssue
		title = "步骤异常上报"
		content = step.Name + " 上报异常"
	default:
		return nil
	}
	stepID := step.ID
	drillID := step.DrillInstanceID
	return &entity.Notification{
		UserID:  cmd.OperatorID,
		Type:    notifType,
		Title:   title,
		Content: content,
		DrillID: &drillID,
		StepID:  &stepID,
		IsRead:  false,
	}
}

func (e *FlowCommandExecutor) publishCollected(evts []events.WSMessage) {
	if e.publisher == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for _, evt := range evts {
		_ = e.publisher.Publish(ctx, evt)
	}
}

func (e *FlowCommandExecutor) loadStep(stepID uint64) (*entity.StepInstance, error) {
	db := e.db
	if db == nil {
		db = repository.DB
	}
	var step entity.StepInstance
	if err := db.First(&step, stepID).Error; err != nil {
		return nil, err
	}
	return &step, nil
}

func decodePayload(raw string, target any) error {
	if raw == "" {
		return nil
	}
	if err := json.Unmarshal([]byte(raw), target); err != nil {
		return &commandError{Code: "invalid_payload", Message: fmt.Sprintf("decode payload: %v", err)}
	}
	return nil
}
