package service

import (
	"context"

	"drill-platform/internal/infrastructure/presence"
)

type DashboardRepository interface {
	CountTeam(context.Context, []uint64) (int64, int64, error)
	CountMyTemplates(context.Context, uint64) (int64, error)
	AverageStepDuration(context.Context) (*int64, error)
}

type PresenceStore interface {
	Touch(context.Context, uint64) error
	OnlineIDs(context.Context) ([]uint64, error)
}

type TeamStats struct {
	Online *int64 `json:"team_online_count"`
	Total  int64  `json:"team_total_count"`
}

type DashboardService struct {
	repo     DashboardRepository
	presence PresenceStore
}

func NewDashboardService(repo DashboardRepository, store PresenceStore) *DashboardService {
	return &DashboardService{repo: repo, presence: store}
}

// SetPresence is called only during startup, before serving requests.
func (s *DashboardService) SetPresence(store PresenceStore) { s.presence = store }

func (s *DashboardService) Touch(ctx context.Context, userID uint64) error {
	if s.presence == nil {
		return presence.ErrUnavailable
	}
	return s.presence.Touch(ctx, userID)
}

func (s *DashboardService) TeamStats(ctx context.Context) (TeamStats, error) {
	var ids []uint64
	available := false
	if s.presence != nil {
		var err error
		ids, err = s.presence.OnlineIDs(ctx)
		available = err == nil
		if !available {
			ids = nil
		}
	}
	total, online, err := s.repo.CountTeam(ctx, ids)
	if err != nil {
		return TeamStats{}, err
	}
	result := TeamStats{Total: total}
	if available {
		result.Online = &online
	}
	return result, nil
}

func (s *DashboardService) MyTemplateCount(ctx context.Context, userID uint64) (int64, error) {
	return s.repo.CountMyTemplates(ctx, userID)
}

func (s *DashboardService) AverageStepDuration(ctx context.Context) (*int64, error) {
	return s.repo.AverageStepDuration(ctx)
}
