package service

import (
	"encoding/json"
	"testing"
	"time"

	"drill-platform/internal/domain/entity"
	"drill-platform/internal/repository"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type flowAdapterTestRedis struct {
	values map[string]string
}

func newFlowAdapterTestRedis() *flowAdapterTestRedis {
	return &flowAdapterTestRedis{values: map[string]string{}}
}

func (r *flowAdapterTestRedis) Get(key string) (string, error) {
	return r.values[key], nil
}

func (r *flowAdapterTestRedis) Set(key string, value interface{}, _ time.Duration) error {
	switch v := value.(type) {
	case []byte:
		r.values[key] = string(v)
	case string:
		r.values[key] = v
	default:
		buf, _ := json.Marshal(v)
		r.values[key] = string(buf)
	}
	return nil
}

func (r *flowAdapterTestRedis) Delete(keys ...string) error {
	for _, key := range keys {
		delete(r.values, key)
	}
	return nil
}

func setupFlowAdapterTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.Exec(`CREATE TABLE drill_instance_step (
		id INTEGER PRIMARY KEY,
		drill_instance_id INTEGER NOT NULL,
		template_step_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		seq INTEGER NOT NULL,
		status TEXT NOT NULL,
		assignee_ids TEXT NOT NULL,
		start_time DATETIME NULL,
		timeout_at DATETIME NULL
	)`).Error; err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestOnStepStartedInvalidatesStepCache(t *testing.T) {
	db := setupFlowAdapterTestDB(t)
	origDB := repository.DB
	repository.DB = db
	defer func() { repository.DB = origDB }()

	step := entity.StepInstance{
		ID:              12,
		DrillInstanceID: 33,
		StepTemplateID:  302130,
		Name:            "检查网络设备",
		Seq:             7,
		Status:          "pending",
		AssigneeIDs:     "[]",
	}
	if err := db.Table("drill_instance_step").Create(map[string]interface{}{
		"id":                step.ID,
		"drill_instance_id": step.DrillInstanceID,
		"template_step_id":  step.StepTemplateID,
		"name":              step.Name,
		"seq":               step.Seq,
		"status":            step.Status,
		"assignee_ids":      step.AssigneeIDs,
	}).Error; err != nil {
		t.Fatalf("create step: %v", err)
	}

	redis := newFlowAdapterTestRedis()
	SetCachedSteps(redis, 33, []entity.StepInstance{step})

	adapter := NewDrillFlowAdapter(nil, nil, nil, nil, nil, nil, nil)
	adapter.SetRedis(redis)
	adapter.OnStepStarted(12, time.Now().Add(time.Hour))

	if _, ok := GetCachedSteps(redis, 33); ok {
		t.Fatalf("expected step cache to be invalidated after step start")
	}
}
