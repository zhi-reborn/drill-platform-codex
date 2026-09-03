package repository

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"drill-platform/internal/domain/entity"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupDashboardRepo(t *testing.T) *DashboardRepo {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		IgnoreRelationshipsWhenMigrating:         true,
		Logger:                                   logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	previousDB := DB
	DB = db
	t.Cleanup(func() {
		DB = previousDB
		_ = sqlDB.Close()
	})
	if err := db.AutoMigrate(&entity.User{}, &entity.DrillTemplate{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	// SQLite index names are database-wide; MySQL permits this name on both tables.
	if err := db.Migrator().DropIndex(&entity.DrillTemplate{}, "idx_status"); err != nil {
		t.Fatalf("drop conflicting template index: %v", err)
	}
	if err := db.AutoMigrate(&entity.DrillInstance{}, &entity.StepInstance{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return NewDashboardRepo()
}

func createDashboardRows(t *testing.T, value any) {
	t.Helper()
	if err := DB.Create(value).Error; err != nil {
		t.Fatalf("create fixtures: %v", err)
	}
}

func TestDashboardCountTeamUsesAllEnabledExistingUsers(t *testing.T) {
	repo := setupDashboardRepo(t)
	users := make([]entity.User, 63)
	for i := range users {
		users[i] = entity.User{ID: uint64(i + 1), Username: fmt.Sprintf("user-%d", i), Status: 1}
	}
	createDashboardRows(t, &users)
	if err := DB.Model(&entity.User{}).Where("id = ?", 62).Update("status", 0).Error; err != nil {
		t.Fatal(err)
	}
	if err := DB.Delete(&entity.User{}, 63).Error; err != nil {
		t.Fatal(err)
	}
	total, online, err := repo.CountTeam(context.Background(), []uint64{1, 1, 61, 62, 63, 999})
	if err != nil || total != 61 || online != 2 {
		t.Fatalf("CountTeam = (%d, %d, %v), want (61, 2, nil)", total, online, err)
	}
	total, online, err = repo.CountTeam(context.Background(), nil)
	if err != nil || total != 61 || online != 0 {
		t.Fatalf("CountTeam without presence = (%d, %d, %v), want (61, 0, nil)", total, online, err)
	}
}

func TestDashboardCountMyTemplatesFiltersCreatorAndDeletedRows(t *testing.T) {
	repo := setupDashboardRepo(t)
	deletedAt := time.Now()
	templates := make([]entity.DrillTemplate, 60)
	for i := range templates {
		templates[i] = entity.DrillTemplate{Name: fmt.Sprintf("template-%d", i), CreatedBy: 7, Status: 1}
	}
	templates = append(templates,
		entity.DrillTemplate{Name: "other creator", CreatedBy: 8},
		entity.DrillTemplate{Name: "deleted", CreatedBy: 7, DeletedAt: &deletedAt},
	)
	createDashboardRows(t, &templates)
	if err := DB.Model(&entity.DrillTemplate{}).Where("id = ?", templates[0].ID).Update("status", 0).Error; err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		userID uint64
		want   int64
	}{{7, 60}, {8, 1}, {9, 0}} {
		got, err := repo.CountMyTemplates(context.Background(), tc.userID)
		if err != nil || got != tc.want {
			t.Fatalf("CountMyTemplates(%d) = (%d, %v), want (%d, nil)", tc.userID, got, err, tc.want)
		}
	}
}

func dashboardStep(id, drillID uint64, duration time.Duration) entity.StepInstance {
	start := time.Date(2026, 9, 3, 10, 0, 0, 0, time.UTC)
	end := start.Add(duration)
	return entity.StepInstance{
		ID: id, DrillInstanceID: drillID, Name: fmt.Sprintf("step-%d", id),
		Status: "completed", AssigneeIDs: "[]", StartTime: &start, EndTime: &end,
	}
}

func TestDashboardAverageStepDurationFiltersToCompletedLeafSteps(t *testing.T) {
	repo := setupDashboardRepo(t)
	createDashboardRows(t, &[]entity.DrillInstance{{ID: 1, Name: "running", Status: "running"}, {ID: 2, Name: "completed", Status: "completed"}, {ID: 3, Name: "deleted"}})
	steps := make([]entity.StepInstance, 60)
	for i := range steps {
		steps[i] = dashboardStep(uint64(i+1), uint64(i%2+1), 10*time.Second)
	}
	// A reference from a different drill must not turn step 1 into a parent.
	crossDrill := dashboardStep(61, 2, 100*time.Second)
	crossDrill.ParentStepID = &steps[0].ID
	steps = append(steps, crossDrill)
	parent := dashboardStep(100, 1, time.Hour)
	child := dashboardStep(101, 1, 500*time.Second)
	child.ParentStepID = &parent.ID
	grandchild := dashboardStep(102, 1, 200*time.Second)
	grandchild.ParentStepID = &child.ID
	steps = append(steps, parent, child, grandchild)
	stage := dashboardStep(103, 1, time.Hour)
	stage.Phase, stage.PhaseStep = "stage", "stage"
	steps = append(steps, stage)
	ordinary := dashboardStep(104, 1, 300*time.Second)
	ordinary.Phase, ordinary.PhaseStep = "stage", "ordinary"
	steps = append(steps, ordinary)
	for i, status := range []string{"skipped", "timeout", "issue", "pending", "running"} {
		step := dashboardStep(uint64(110+i), 1, time.Hour)
		step.Status = status
		steps = append(steps, step)
	}
	noStart := dashboardStep(120, 1, time.Hour)
	noStart.StartTime = nil
	noEnd := dashboardStep(121, 1, time.Hour)
	noEnd.EndTime = nil
	steps = append(steps, noStart, noEnd, dashboardStep(122, 1, -time.Hour),
		dashboardStep(123, 999, time.Hour), dashboardStep(124, 3, time.Hour))
	createDashboardRows(t, &steps)
	if err := DB.Delete(&entity.DrillInstance{}, 3).Error; err != nil {
		t.Fatal(err)
	}
	got, err := repo.AverageStepDuration(context.Background())
	if err != nil || got == nil || *got != 19 {
		t.Fatalf("AverageStepDuration = (%v, %v), want (19, nil)", got, err)
	}
}

func TestDashboardAverageStepDurationRoundsFractionalMean(t *testing.T) {
	for _, tc := range []struct {
		name      string
		durations []time.Duration
		want      int64
	}{
		{"integer samples", []time.Duration{time.Second, 2 * time.Second}, 2},
		{"fractional samples", []time.Duration{1200 * time.Millisecond, 1800 * time.Millisecond}, 2},
		{"round down", []time.Duration{time.Second, 1800 * time.Millisecond}, 1},
		{"zero", []time.Duration{0}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := setupDashboardRepo(t)
			createDashboardRows(t, &entity.DrillInstance{ID: 1, Name: "drill"})
			for i, duration := range tc.durations {
				step := dashboardStep(uint64(i+1), 1, duration)
				createDashboardRows(t, &step)
			}
			got, err := repo.AverageStepDuration(context.Background())
			if err != nil || got == nil || *got != tc.want {
				t.Fatalf("AverageStepDuration = (%v, %v), want (%d, nil)", got, err, tc.want)
			}
		})
	}
}

func TestDashboardEmptyAggregates(t *testing.T) {
	repo := setupDashboardRepo(t)
	total, online, err := repo.CountTeam(context.Background(), []uint64{1})
	if err != nil || total != 0 || online != 0 {
		t.Fatalf("CountTeam = (%d, %d, %v), want (0, 0, nil)", total, online, err)
	}
	got, err := repo.AverageStepDuration(context.Background())
	if err != nil || got != nil {
		t.Fatalf("AverageStepDuration = (%v, %v), want (nil, nil)", got, err)
	}
	createDashboardRows(t, &entity.DrillInstance{ID: 1, Name: "drill"})
	step := dashboardStep(1, 1, time.Second)
	step.Status = "skipped"
	createDashboardRows(t, &step)
	got, err = repo.AverageStepDuration(context.Background())
	if err != nil || got != nil {
		t.Fatalf("AverageStepDuration with invalid samples = (%v, %v), want (nil, nil)", got, err)
	}
}

func TestDashboardPropagatesDatabaseErrors(t *testing.T) {
	repo := setupDashboardRepo(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := repo.CountTeam(ctx, nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("CountTeam error = %v, want context.Canceled", err)
	}
	if _, err := repo.CountMyTemplates(ctx, 1); !errors.Is(err, context.Canceled) {
		t.Fatalf("CountMyTemplates error = %v, want context.Canceled", err)
	}
	if _, err := repo.AverageStepDuration(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("AverageStepDuration error = %v, want context.Canceled", err)
	}
	sqlDB, err := DB.DB()
	if err != nil {
		t.Fatal(err)
	}
	if err := sqlDB.Close(); err != nil {
		t.Fatal(err)
	}
	if _, _, err := repo.CountTeam(context.Background(), []uint64{1}); err == nil {
		t.Fatal("CountTeam swallowed database error")
	}
	if _, err := repo.CountMyTemplates(context.Background(), 1); err == nil {
		t.Fatal("CountMyTemplates swallowed database error")
	}
	if _, err := repo.AverageStepDuration(context.Background()); err == nil {
		t.Fatal("AverageStepDuration swallowed database error")
	}
}
