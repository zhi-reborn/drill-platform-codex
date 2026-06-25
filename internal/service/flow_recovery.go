package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"drill-platform/internal/domain/entity"
	"drill-platform/internal/repository"
	"drill-platform/internal/worker"
)

// FlowRecovery rebuilds in-memory flow state for running and paused drills
// after a Worker wins leadership. It is the authoritative recovery entry point
// and implements worker.Recoverer.
//
// Recovery is a side-effect-free orchestration layer: it delegates engine
// reconstruction to DrillService.Recover and then registers future timeouts
// with the scheduler and enqueues deterministic internal commands for expired
// timeouts. It does not create logs, notifications, or WebSocket events on
// its own. (Note: DrillService.Recover itself may write to the DB for step
// template backfill and pre-step ID recomputation; see the concern comment
// in RecoverAll.)
type FlowRecovery struct {
	drillService    *DrillService
	drillRepo       *repository.DrillRepo
	stepRepo        *repository.StepRepo
	flowCommandRepo *repository.FlowCommandRepo
}

// NewFlowRecovery constructs a FlowRecovery with the given dependencies.
func NewFlowRecovery(
	drillService *DrillService,
	drillRepo *repository.DrillRepo,
	stepRepo *repository.StepRepo,
	flowCommandRepo *repository.FlowCommandRepo,
) *FlowRecovery {
	return &FlowRecovery{
		drillService:    drillService,
		drillRepo:       drillRepo,
		stepRepo:        stepRepo,
		flowCommandRepo: flowCommandRepo,
	}
}

// Recover implements worker.Recoverer by delegating to RecoverAll.
func (r *FlowRecovery) Recover(ctx context.Context) error {
	return r.RecoverAll(ctx)
}

// RecoverAll loads all running and paused drills from the database, rebuilds
// the engine state for each via DrillService.Recover, then:
//
//   - Registers future (unexpired) timeouts with the engine's TimeoutScheduler.
//   - Enqueues a deterministic internal step_timeout command for expired
//     timeouts so the elected Worker can process them through the normal
//     command queue.
//
// The timeout scheduler remains in the elected Worker only. Its callback
// submits a deterministic internal command with idempotency key:
//
//	timeout:<drill-id>:<step-id>:<timeout-unix>
//
// Recovery itself does not execute timeout effects directly.
func (r *FlowRecovery) RecoverAll(ctx context.Context) error {
	var drills []entity.DrillInstance
	if err := repository.DB.
		Where("status IN ?", []string{"running", "paused"}).
		Find(&drills).Error; err != nil {
		return fmt.Errorf("load running/paused drills: %w", err)
	}

	for i := range drills {
		drillID := drills[i].ID
		if err := r.drillService.Recover(drillID); err != nil {
			log.Printf("[FlowRecovery] Recover drill %d failed: %v (continuing)", drillID, err)
			continue
		}
		if err := r.registerTimeouts(drillID); err != nil {
			log.Printf("[FlowRecovery] register timeouts for drill %d failed: %v (continuing)", drillID, err)
		}
	}
	return nil
}

// registerTimeouts inspects the steps of a recovered drill and:
//
//   - Registers steps with a future TimeoutAt and running status to the
//     engine's TimeoutScheduler.
//   - Submits an internal step_timeout command for steps with an expired
//     TimeoutAt and running status, so the Worker can process the timeout
//     through the durable command queue.
func (r *FlowRecovery) registerTimeouts(drillID uint64) error {
	engine := r.drillService.Engine()
	if engine == nil {
		return fmt.Errorf("engine not initialized for drill %d", drillID)
	}
	scheduler := engine.TimeoutScheduler()
	if scheduler == nil {
		return fmt.Errorf("timeout scheduler not available for drill %d", drillID)
	}

	steps, err := r.stepRepo.FindStepsByDrillID(drillID)
	if err != nil {
		return fmt.Errorf("load steps for drill %d: %w", drillID, err)
	}

	now := time.Now()
	for _, step := range steps {
		if step.Status != "running" || step.TimeoutAt == nil {
			continue
		}
		timeoutAt := *step.TimeoutAt
		flowInstID := int64(drillID)
		stepDefID := int64(step.StepTemplateID)
		stepInstID := int64(step.ID)

		if timeoutAt.After(now) {
			// Future timeout: register with the scheduler.
			scheduler.Register(flowInstID, stepDefID, stepInstID, timeoutAt)
			continue
		}

		// Expired timeout: enqueue a deterministic internal command.
		if err := r.submitTimeoutCommand(drillID, step.ID, timeoutAt); err != nil {
			log.Printf("[FlowRecovery] submit timeout command for drill %d step %d failed: %v",
				drillID, step.ID, err)
		}
	}
	return nil
}

// submitTimeoutCommand enqueues an internal step_timeout command with a
// deterministic idempotency key so duplicate recovery runs do not produce
// duplicate commands.
func (r *FlowRecovery) submitTimeoutCommand(drillID, stepID uint64, timeoutAt time.Time) error {
	stepIDCopy := stepID
	cmd := &entity.FlowCommand{
		CommandType:     "step_timeout",
		DrillInstanceID: drillID,
		StepInstanceID:  &stepIDCopy,
		OperatorID:      0,
		IdempotencyKey:  fmt.Sprintf("timeout:%d:%d:%d", drillID, stepID, timeoutAt.Unix()),
		Payload:         "{}",
		Status:          entity.FlowCommandPending,
	}
	if _, _, err := r.flowCommandRepo.CreateOrGet(cmd); err != nil {
		return fmt.Errorf("create timeout command: %w", err)
	}
	return nil
}

// Compile-time assertion that FlowRecovery satisfies worker.Recoverer.
var _ worker.Recoverer = (*FlowRecovery)(nil)
