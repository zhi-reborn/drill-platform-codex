package service

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type dashboardRepoStub struct {
	err    error
	ids    []uint64
	userID uint64
}

func (r *dashboardRepoStub) CountTeam(_ context.Context, ids []uint64) (int64, int64, error) {
	r.ids = ids
	return 10, 2, r.err
}
func (r *dashboardRepoStub) CountMyTemplates(_ context.Context, userID uint64) (int64, error) {
	r.userID = userID
	return 0, r.err
}
func (r *dashboardRepoStub) AverageStepDuration(context.Context) (*int64, error) {
	return nil, r.err
}

type dashboardPresenceStub struct{ err error }

func (s *dashboardPresenceStub) Touch(context.Context, uint64) error { return s.err }
func (s *dashboardPresenceStub) OnlineIDs(context.Context) ([]uint64, error) {
	return []uint64{1, 2}, s.err
}

func TestDashboardServiceTeamAndUnavailablePresence(t *testing.T) {
	ctx := context.Background()
	r := &dashboardRepoStub{}
	p := &dashboardPresenceStub{}
	s := NewDashboardService(r, p)
	got, err := s.TeamStats(ctx)
	if err != nil || got.Total != 10 || got.Online == nil || *got.Online != 2 || !reflect.DeepEqual(r.ids, []uint64{1, 2}) {
		t.Fatalf("team = %+v, %v; ids=%v", got, err, r.ids)
	}
	p.err = errors.New("redis down")
	got, err = s.TeamStats(ctx)
	if err != nil || got.Total != 10 || got.Online != nil || len(r.ids) != 0 {
		t.Fatalf("degraded team = %+v, %v; ids=%v", got, err, r.ids)
	}
	if s.Touch(ctx, 1) == nil {
		t.Fatal("heartbeat must report recording failure")
	}
	s.SetPresence(nil)
	got, err = s.TeamStats(ctx)
	if err != nil || got.Online != nil || got.Total != 10 {
		t.Fatalf("no Redis = %+v, %v", got, err)
	}
	if s.Touch(ctx, 1) == nil {
		t.Fatal("unconfigured heartbeat must fail")
	}
}

func TestDashboardServicePreservesEmptyResultsAndErrors(t *testing.T) {
	r := &dashboardRepoStub{}
	s := NewDashboardService(r, nil)
	ctx := context.Background()
	if count, err := s.MyTemplateCount(ctx, 7); err != nil || count != 0 || r.userID != 7 {
		t.Fatalf("templates = %d, %v", count, err)
	}
	if avg, err := s.AverageStepDuration(ctx); err != nil || avg != nil {
		t.Fatalf("empty average = %v, %v", avg, err)
	}
	r.err = errors.New("database down")
	if _, err := s.TeamStats(ctx); !errors.Is(err, r.err) {
		t.Fatalf("team error = %v", err)
	}
	if _, err := s.MyTemplateCount(ctx, 7); !errors.Is(err, r.err) {
		t.Fatalf("template error = %v", err)
	}
	if _, err := s.AverageStepDuration(ctx); !errors.Is(err, r.err) {
		t.Fatalf("average error = %v", err)
	}
}
