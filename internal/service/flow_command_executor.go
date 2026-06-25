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

	"gorm.io/gorm"
)

// LeaderGuard fences executor leadership. redis.Lease satisfies this interface.
type LeaderGuard interface {
	Value() string
	Renew(ctx context.Context) (bool, error)
}

// CommandRepo abstracts the durable command persistence boundary used by the
// executor. FlowCommandRepo satisfies this interface.
type CommandRepo interface {
	MarkSucceeded(id uint64, result any) error
	MarkFailed(id uint64, code, message string) error
}

// FlowCommandExecutor maps a durable FlowCommand to its transactional side
// effects. It implements worker.Executor.
type FlowCommandExecutor struct {
	db        *gorm.DB
	commands  CommandRepo
	drills    *DrillService
	tasks     *TaskService
	stepRepo  *repository.StepRepo
	drillRepo *repository.DrillRepo
	publisher events.Publisher
	leader    LeaderGuard
}

// NewFlowCommandExecutor constructs an executor with the given dependencies.
// drills and tasks may be nil in tests that exercise the DB-fallback path.
func NewFlowCommandExecutor(
	db *gorm.DB,
	commands CommandRepo,
	drills *DrillService,
	tasks *TaskService,
	publisher events.Publisher,
	leader LeaderGuard,
) *FlowCommandExecutor {
	return &FlowCommandExecutor{
		db:        db,
		commands:  commands,
		drills:    drills,
		tasks:     tasks,
		stepRepo:  repository.NewStepRepo(),
		drillRepo: repository.NewDrillRepo(),
		publisher: publisher,
		leader:    leader,
	}
}

// Execute dispatches a single FlowCommand to its typed handler. It is
// idempotent: terminal commands return nil immediately. The Worker ignores
// the returned error, so Execute marks the command terminal internally.
func (e *FlowCommandExecutor) Execute(ctx context.Context, cmd *entity.FlowCommand) error {
	if cmd.IsTerminal() {
		return nil
	}

	if e.leader != nil {
		if e.leader.Value() == "" {
			return errors.New("executor is not the elected leader")
		}
		if ok, err := e.leader.Renew(ctx); err != nil || !ok {
			return errors.New("executor lost leadership before executing command")
		}
	}

	execErr := e.dispatch(ctx, cmd)
	if execErr != nil {
		code, message := classifyError(execErr)
		_ = e.commands.MarkFailed(cmd.ID, code, message)
		return execErr
	}

	_ = e.commands.MarkSucceeded(cmd.ID, nil)
	return nil
}

func (e *FlowCommandExecutor) dispatch(ctx context.Context, cmd *entity.FlowCommand) error {
	switch cmd.CommandType {
	case "start_drill":
		return e.executeStartDrill(ctx, cmd)
	case "pause_drill":
		return e.executePauseDrill(ctx, cmd)
	case "resume_drill":
		return e.executeResumeDrill(ctx, cmd)
	case "terminate_drill":
		return e.executeTerminateDrill(ctx, cmd)
	case "start_step":
		return e.executeStartStep(ctx, cmd)
	case "complete_step":
		return e.executeCompleteStep(ctx, cmd)
	case "report_issue":
		return e.executeReportIssue(ctx, cmd)
	case "skip_step":
		return e.executeSkipStep(ctx, cmd)
	case "force_complete_step":
		return e.executeForceCompleteStep(ctx, cmd)
	case "resume_task":
		return e.executeResumeTask(ctx, cmd)
	case "assign_step":
		return e.executeAssignStep(ctx, cmd)
	case "update_step_info":
		return e.executeUpdateStepInfo(ctx, cmd)
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

func (e *FlowCommandExecutor) executeStartDrill(_ context.Context, cmd *entity.FlowCommand) error {
	if e.drills != nil {
		return e.drills.Start(cmd.DrillInstanceID)
	}
	return e.transitionDrillStatus(cmd, []string{"pending"}, "running")
}

func (e *FlowCommandExecutor) executePauseDrill(_ context.Context, cmd *entity.FlowCommand) error {
	if e.drills != nil {
		return e.drills.Pause(cmd.DrillInstanceID)
	}
	return e.transitionDrillStatus(cmd, []string{"running"}, "paused")
}

func (e *FlowCommandExecutor) executeResumeDrill(_ context.Context, cmd *entity.FlowCommand) error {
	if e.drills != nil {
		return e.drills.Resume(cmd.DrillInstanceID)
	}
	return e.transitionDrillStatus(cmd, []string{"paused"}, "running")
}

func (e *FlowCommandExecutor) executeTerminateDrill(_ context.Context, cmd *entity.FlowCommand) error {
	if e.drills != nil {
		return e.drills.Terminate(cmd.DrillInstanceID)
	}
	return e.transitionDrillStatus(cmd, []string{"running", "paused", "issue"}, "terminated")
}

func (e *FlowCommandExecutor) transitionDrillStatus(cmd *entity.FlowCommand, from []string, to string) error {
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
				return markCommandSucceededInTx(tx, cmd.ID)
			}
			return &commandError{Code: "invalid_status", Message: fmt.Sprintf("drill status is %s, expected one of the allowed transitions", drill.Status)}
		}
		return markCommandSucceededInTx(tx, cmd.ID)
	})
}

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

func (e *FlowCommandExecutor) executeStartStep(ctx context.Context, cmd *entity.FlowCommand) error {
	if cmd.StepInstanceID == nil {
		return &commandError{Code: "missing_step", Message: "step_instance_id is required"}
	}
	if e.tasks != nil {
		return e.tasks.StartStep(*cmd.StepInstanceID, cmd.OperatorID)
	}
	return e.transitionStepInTx(cmd, []string{"pending"}, "running", map[string]any{
		"start_time": time.Now(),
	})
}

func (e *FlowCommandExecutor) executeCompleteStep(ctx context.Context, cmd *entity.FlowCommand) error {
	if cmd.StepInstanceID == nil {
		return &commandError{Code: "missing_step", Message: "step_instance_id is required"}
	}
	var payload CompleteStepPayload
	if err := decodePayload(cmd.Payload, &payload); err != nil {
		return err
	}

	if e.drills != nil && e.drills.Engine() != nil {
		return e.executeCompleteStepViaEngine(cmd, payload)
	}

	return e.transitionStepInTx(cmd, []string{"running"}, "completed", map[string]any{
		"actual_operator": cmd.OperatorID,
		"end_time":        time.Now(),
		"remark":          payload.Remark,
	})
}

func (e *FlowCommandExecutor) executeCompleteStepViaEngine(cmd *entity.FlowCommand, payload CompleteStepPayload) error {
	step, err := e.loadStep(*cmd.StepInstanceID)
	if err != nil {
		return err
	}
	engine := e.drills.Engine()
	err = engine.CompleteStep(
		int64(step.DrillInstanceID),
		int64(step.StepTemplateID),
		int64(cmd.OperatorID),
		payload.Remark,
	)
	if errors.Is(err, flowengine.ErrInstanceNotFound) {
		if recErr := e.drills.Recover(step.DrillInstanceID); recErr != nil {
			return &commandError{Code: "recover_failed", Message: recErr.Error()}
		}
		err = engine.CompleteStep(
			int64(step.DrillInstanceID),
			int64(step.StepTemplateID),
			int64(cmd.OperatorID),
			payload.Remark,
		)
	}
	if errors.Is(err, flowengine.ErrStepNotActive) || errors.Is(err, flowengine.ErrInvalidStatus) {
		return e.checkStepIdempotentOrError(*cmd.StepInstanceID, "completed")
	}
	return err
}

func (e *FlowCommandExecutor) executeReportIssue(ctx context.Context, cmd *entity.FlowCommand) error {
	if cmd.StepInstanceID == nil {
		return &commandError{Code: "missing_step", Message: "step_instance_id is required"}
	}
	var payload ReportIssuePayload
	if err := decodePayload(cmd.Payload, &payload); err != nil {
		return err
	}

	if e.drills != nil && e.drills.Engine() != nil {
		return e.executeReportIssueViaEngine(cmd, payload)
	}

	return e.transitionStepInTx(cmd, []string{"running"}, "issue", map[string]any{
		"issue_desc": payload.IssueDesc,
		"end_time":   time.Now(),
	})
}

func (e *FlowCommandExecutor) executeReportIssueViaEngine(cmd *entity.FlowCommand, payload ReportIssuePayload) error {
	step, err := e.loadStep(*cmd.StepInstanceID)
	if err != nil {
		return err
	}
	engine := e.drills.Engine()
	err = engine.ReportIssue(
		int64(step.DrillInstanceID),
		int64(step.StepTemplateID),
		int64(cmd.OperatorID),
		payload.IssueDesc,
	)
	if errors.Is(err, flowengine.ErrInstanceNotFound) {
		if recErr := e.drills.Recover(step.DrillInstanceID); recErr != nil {
			return &commandError{Code: "recover_failed", Message: recErr.Error()}
		}
		err = engine.ReportIssue(
			int64(step.DrillInstanceID),
			int64(step.StepTemplateID),
			int64(cmd.OperatorID),
			payload.IssueDesc,
		)
	}
	if errors.Is(err, flowengine.ErrStepNotActive) || errors.Is(err, flowengine.ErrInvalidStatus) {
		return e.checkStepIdempotentOrError(*cmd.StepInstanceID, "issue")
	}
	return err
}

func (e *FlowCommandExecutor) executeSkipStep(ctx context.Context, cmd *entity.FlowCommand) error {
	if cmd.StepInstanceID == nil {
		return &commandError{Code: "missing_step", Message: "step_instance_id is required"}
	}

	if e.drills != nil && e.drills.Engine() != nil {
		step, err := e.loadStep(*cmd.StepInstanceID)
		if err != nil {
			return err
		}
		engine := e.drills.Engine()
		err = engine.DirectorSkipStep(
			int64(step.DrillInstanceID),
			int64(step.StepTemplateID),
			int64(cmd.OperatorID),
		)
		if errors.Is(err, flowengine.ErrInstanceNotFound) {
			if recErr := e.drills.Recover(step.DrillInstanceID); recErr != nil {
				return &commandError{Code: "recover_failed", Message: recErr.Error()}
			}
			err = engine.DirectorSkipStep(
				int64(step.DrillInstanceID),
				int64(step.StepTemplateID),
				int64(cmd.OperatorID),
			)
		}
		if errors.Is(err, flowengine.ErrStepNotActive) || errors.Is(err, flowengine.ErrInvalidStatus) {
			return e.checkStepIdempotentOrError(*cmd.StepInstanceID, "skipped")
		}
		return err
	}

	return e.transitionStepInTx(cmd, []string{"running", "pending"}, "skipped", map[string]any{
		"end_time": time.Now(),
	})
}

func (e *FlowCommandExecutor) executeForceCompleteStep(ctx context.Context, cmd *entity.FlowCommand) error {
	if cmd.StepInstanceID == nil {
		return &commandError{Code: "missing_step", Message: "step_instance_id is required"}
	}

	if e.drills != nil && e.drills.Engine() != nil {
		step, err := e.loadStep(*cmd.StepInstanceID)
		if err != nil {
			return err
		}
		engine := e.drills.Engine()
		err = engine.DirectorForceComplete(
			int64(step.DrillInstanceID),
			int64(step.StepTemplateID),
			int64(cmd.OperatorID),
		)
		if errors.Is(err, flowengine.ErrInstanceNotFound) {
			if recErr := e.drills.Recover(step.DrillInstanceID); recErr != nil {
				return &commandError{Code: "recover_failed", Message: recErr.Error()}
			}
			err = engine.DirectorForceComplete(
				int64(step.DrillInstanceID),
				int64(step.StepTemplateID),
				int64(cmd.OperatorID),
			)
		}
		if errors.Is(err, flowengine.ErrStepNotActive) || errors.Is(err, flowengine.ErrInvalidStatus) {
			return e.checkStepIdempotentOrError(*cmd.StepInstanceID, "completed")
		}
		return err
	}

	return e.transitionStepInTx(cmd, []string{"running", "pending", "issue"}, "completed", map[string]any{
		"actual_operator": cmd.OperatorID,
		"end_time":        time.Now(),
	})
}

func (e *FlowCommandExecutor) executeResumeTask(ctx context.Context, cmd *entity.FlowCommand) error {
	if cmd.StepInstanceID == nil {
		return &commandError{Code: "missing_step", Message: "step_instance_id is required"}
	}

	if e.drills != nil && e.drills.Engine() != nil {
		step, err := e.loadStep(*cmd.StepInstanceID)
		if err != nil {
			return err
		}
		engine := e.drills.Engine()
		stepDefID := int64(step.StepTemplateID)
		err = engine.Intervene(
			int64(step.DrillInstanceID),
			flowengine.ActionResumeTask,
			&stepDefID,
			int64(cmd.OperatorID),
		)
		if errors.Is(err, flowengine.ErrInstanceNotFound) {
			if recErr := e.drills.Recover(step.DrillInstanceID); recErr != nil {
				return &commandError{Code: "recover_failed", Message: recErr.Error()}
			}
			err = engine.Intervene(
				int64(step.DrillInstanceID),
				flowengine.ActionResumeTask,
				&stepDefID,
				int64(cmd.OperatorID),
			)
		}
		if errors.Is(err, flowengine.ErrStepNotActive) || errors.Is(err, flowengine.ErrInvalidStatus) {
			return e.checkStepIdempotentOrError(*cmd.StepInstanceID, "running")
		}
		return err
	}

	return e.transitionStepInTx(cmd, []string{"completed", "skipped", "timeout", "issue"}, "running", map[string]any{
		"start_time": time.Now(),
	})
}

func (e *FlowCommandExecutor) executeAssignStep(ctx context.Context, cmd *entity.FlowCommand) error {
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
		map[string]any{"assignee_ids": assigneeJSON},
		"assign",
		step.Name+" 分配已更新",
		notif,
	)
}

func (e *FlowCommandExecutor) executeUpdateStepInfo(ctx context.Context, cmd *entity.FlowCommand) error {
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
		map[string]any{"remark": payload.Remark},
		"update_info",
		content,
		nil,
	)
}

// transitionStepInTx performs a conditional step status transition within a
// GORM transaction. On success it creates a log and notification carrying
// command_id, marks the command succeeded in the same transaction, commits,
// then publishes a WebSocket event. This is the DB-fallback path used when
// the flow engine is not available (e.g. tests).
func (e *FlowCommandExecutor) transitionStepInTx(
	cmd *entity.FlowCommand,
	from []string,
	to string,
	updates map[string]any,
) error {
	db := e.db
	if db == nil {
		db = repository.DB
	}

	var collectedEvents []events.Event

	err := db.Transaction(func(tx *gorm.DB) error {
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
				return markCommandSucceededInTx(tx, cmd.ID)
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

		if err := markCommandSucceededInTx(tx, cmd.ID); err != nil {
			return err
		}

		collectedEvents = append(collectedEvents, events.Event{
			Type:      fmt.Sprintf("step_%s", to),
			DrillID:    step.DrillInstanceID,
			Payload:    []byte(cmd.Payload),
			CreatedAt:  time.Now(),
		})

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
func markCommandSucceededInTx(tx *gorm.DB, cmdID uint64) error {
	return tx.Model(&entity.FlowCommand{}).Where("id = ?", cmdID).Updates(map[string]any{
		"status":      entity.FlowCommandSucceeded,
		"finished_at": time.Now(),
	}).Error
}

// transitionStepFieldsInTx updates step fields (without changing status) inside
// a GORM transaction. It creates a log and optional notification carrying
// command_id, marks the command succeeded, commits, then publishes an event.
// Used by assign_step and update_step_info.
func (e *FlowCommandExecutor) transitionStepFieldsInTx(
	cmd *entity.FlowCommand,
	updates map[string]any,
	action string,
	content string,
	notif *entity.Notification,
) error {
	db := e.db
	if db == nil {
		db = repository.DB
	}

	var collectedEvents []events.Event

	err := db.Transaction(func(tx *gorm.DB) error {
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

		if err := markCommandSucceededInTx(tx, cmd.ID); err != nil {
			return err
		}

		collectedEvents = append(collectedEvents, events.Event{
			Type:      fmt.Sprintf("step_%s", action),
			DrillID:   step.DrillInstanceID,
			Payload:   []byte(cmd.Payload),
			CreatedAt: time.Now(),
		})

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

func (e *FlowCommandExecutor) publishCollected(evts []events.Event) {
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

// checkStepIdempotentOrError returns nil when the step already reached
// targetStatus (idempotent replay), otherwise returns an invalid_status
// commandError. Used by engine paths to tolerate command replays.
func (e *FlowCommandExecutor) checkStepIdempotentOrError(stepID uint64, targetStatus string) error {
	db := e.db
	if db == nil {
		db = repository.DB
	}
	var step entity.StepInstance
	if err := db.Select("status").First(&step, stepID).Error; err != nil {
		return err
	}
	if step.Status == targetStatus {
		return nil
	}
	return &commandError{Code: "invalid_status", Message: fmt.Sprintf("step status is %s, expected %s", step.Status, targetStatus)}
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
