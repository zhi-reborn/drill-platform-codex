package service

import (
	"context"
	"encoding/json"
	"time"

	"drill-platform/internal/domain/entity"
	"drill-platform/internal/repository"

	"github.com/google/uuid"
)

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

type FlowCommandService struct {
	repo         *repository.FlowCommandRepo
	waitTimeout  time.Duration
	pollInterval time.Duration
}

func NewFlowCommandService(repo *repository.FlowCommandRepo, waitTimeout time.Duration) *FlowCommandService {
	return &FlowCommandService{
		repo:         repo,
		waitTimeout:  waitTimeout,
		pollInterval: 50 * time.Millisecond,
	}
}

func (s *FlowCommandService) SubmitAndWait(ctx context.Context, req SubmitCommandRequest) (*SubmitCommandResult, error) {
	payloadBytes, err := json.Marshal(req.Payload)
	if err != nil {
		return nil, err
	}
	idempotencyKey := req.IdempotencyKey
	if idempotencyKey == "" {
		idempotencyKey = uuid.NewString()
	}

	cmd, _, err := s.repo.CreateOrGet(&entity.FlowCommand{
		CommandType:     req.CommandType,
		DrillInstanceID: req.DrillInstanceID,
		StepInstanceID:  req.StepInstanceID,
		OperatorID:      req.OperatorID,
		IdempotencyKey:  idempotencyKey,
		Payload:         string(payloadBytes),
		Status:          entity.FlowCommandPending,
	})
	if err != nil {
		return nil, err
	}
	if cmd.IsTerminal() || s.waitTimeout <= 0 {
		return &SubmitCommandResult{Command: cmd, Pending: !cmd.IsTerminal()}, nil
	}

	deadline := time.NewTimer(s.waitTimeout)
	defer deadline.Stop()
	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-deadline.C:
			latest, err := s.repo.FindByID(cmd.ID)
			if err != nil {
				return nil, err
			}
			return &SubmitCommandResult{Command: latest, Pending: !latest.IsTerminal()}, nil
		case <-ticker.C:
			latest, err := s.repo.FindByID(cmd.ID)
			if err != nil {
				return nil, err
			}
			if latest.IsTerminal() {
				return &SubmitCommandResult{Command: latest, Pending: false}, nil
			}
		}
	}
}

// Submit enqueues a command keyed by IdempotencyKey and returns immediately
// without waiting for a worker to execute it. API handlers use this so request
// threads never block on command execution; clients poll
// GET /flow-commands/:id for the authoritative status. Duplicate submissions
// of the same idempotency key return the original command.
func (s *FlowCommandService) Submit(req SubmitCommandRequest) (*SubmitCommandResult, error) {
	payloadBytes, err := json.Marshal(req.Payload)
	if err != nil {
		return nil, err
	}
	idempotencyKey := req.IdempotencyKey
	if idempotencyKey == "" {
		idempotencyKey = uuid.NewString()
	}
	cmd, _, err := s.repo.CreateOrGet(&entity.FlowCommand{
		CommandType:     req.CommandType,
		DrillInstanceID: req.DrillInstanceID,
		StepInstanceID:  req.StepInstanceID,
		OperatorID:      req.OperatorID,
		IdempotencyKey:  idempotencyKey,
		Payload:         string(payloadBytes),
		Status:          entity.FlowCommandPending,
	})
	if err != nil {
		return nil, err
	}
	return &SubmitCommandResult{Command: cmd, Pending: !cmd.IsTerminal()}, nil
}

func (s *FlowCommandService) GetForOperator(id, operatorID uint64) (*entity.FlowCommand, error) {
	return s.repo.FindByIDForOperator(id, operatorID)
}

func (s *FlowCommandService) Get(id uint64) (*entity.FlowCommand, error) {
	return s.repo.FindByID(id)
}
