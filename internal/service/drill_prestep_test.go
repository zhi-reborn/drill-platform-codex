package service

import (
	"encoding/json"
	"testing"

	"drill-platform/internal/domain/entity"
	"drill-platform/internal/repository"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("无法打开内存数据库: %v", err)
	}
	if err := db.AutoMigrate(&entity.StepInstance{}); err != nil {
		t.Fatalf("数据库迁移失败: %v", err)
	}
	return db
}

func ptrUint64(v uint64) *uint64 { return &v }

func buildTestSteps() []entity.StepInstance {
	return []entity.StepInstance{
		{ID: 196, DrillInstanceID: 74, StepTemplateID: 252, Name: "A", Seq: 1, StepType: "serial", AssigneeIDs: "[]"},
		{ID: 197, DrillInstanceID: 74, StepTemplateID: 253, Name: "B", Seq: 2, StepType: "parallel", AssigneeIDs: "[]"},
		{ID: 198, DrillInstanceID: 74, ParentStepID: ptrUint64(197), StepTemplateID: 254, Name: "B-1", Seq: 3, StepType: "serial", AssigneeIDs: "[]"},
		{ID: 199, DrillInstanceID: 74, ParentStepID: ptrUint64(197), StepTemplateID: 255, Name: "B-2", Seq: 4, StepType: "serial", AssigneeIDs: "[]"},
		{ID: 200, DrillInstanceID: 74, StepTemplateID: 256, Name: "C", Seq: 5, StepType: "parallel", AssigneeIDs: "[]"},
		{ID: 201, DrillInstanceID: 74, ParentStepID: ptrUint64(200), StepTemplateID: 257, Name: "C-1", Seq: 6, StepType: "serial", AssigneeIDs: "[]"},
		{ID: 202, DrillInstanceID: 74, StepTemplateID: 258, Name: "D", Seq: 7, StepType: "serial", AssigneeIDs: "[]"},
		{ID: 203, DrillInstanceID: 74, StepTemplateID: 259, Name: "E", Seq: 8, StepType: "serial", AssigneeIDs: "[]"},
		{ID: 204, DrillInstanceID: 74, StepTemplateID: 260, Name: "H", Seq: 9, StepType: "serial", AssigneeIDs: "[]"},
		{ID: 205, DrillInstanceID: 74, StepTemplateID: 261, Name: "I", Seq: 10, StepType: "serial", AssigneeIDs: "[]"},
		{ID: 206, DrillInstanceID: 74, StepTemplateID: 262, Name: "J", Seq: 11, StepType: "serial", AssigneeIDs: "[]"},
		{ID: 207, DrillInstanceID: 74, ParentStepID: ptrUint64(206), StepTemplateID: 263, Name: "J-1", Seq: 12, StepType: "serial", AssigneeIDs: "[]"},
		{ID: 208, DrillInstanceID: 74, StepTemplateID: 264, Name: "K", Seq: 13, StepType: "parallel", AssigneeIDs: "[]"},
		{ID: 209, DrillInstanceID: 74, StepTemplateID: 265, Name: "L", Seq: 14, StepType: "parallel", AssigneeIDs: "[]"},
		{ID: 210, DrillInstanceID: 74, StepTemplateID: 266, Name: "M", Seq: 15, StepType: "parallel", AssigneeIDs: "[]"},
		{ID: 211, DrillInstanceID: 74, StepTemplateID: 267, Name: "N", Seq: 16, StepType: "serial", AssigneeIDs: "[]"},
	}
}

func expectedPreStepIDs() map[uint64][]uint64 {
	return map[uint64][]uint64{
		196: {},
		197: {196},
		198: {196},
		199: {198},
		200: {196},
		201: {196},
		202: {197, 200},
		203: {202},
		204: {203},
		205: {204},
		206: {205},
		207: {205},
		208: {206},
		209: {206},
		210: {206},
		211: {208, 209, 210},
	}
}

func preIDsToJSON(ids []uint64) string {
	if len(ids) == 0 {
		return "[]"
	}
	b, _ := json.Marshal(ids)
	return string(b)
}

func TestComputeInstancePreStepIDs(t *testing.T) {
	db := setupTestDB(t)
	origDB := repository.DB
	repository.DB = db
	defer func() { repository.DB = origDB }()

	steps := buildTestSteps()
	for i := range steps {
		steps[i].PreStepIDs = ""
		if err := db.Create(&steps[i]).Error; err != nil {
			t.Fatalf("插入步骤 %s (id=%d) 失败: %v", steps[i].Name, steps[i].ID, err)
		}
	}

	svc := &DrillService{}
	svc.computeInstancePreStepIDs(steps, nil)

	var results []entity.StepInstance
	if err := db.Order("seq asc").Find(&results).Error; err != nil {
		t.Fatalf("查询结果失败: %v", err)
	}

	expected := expectedPreStepIDs()
	for _, result := range results {
		exp := expected[result.ID]
		expJSON := preIDsToJSON(exp)
		if result.PreStepIDs != expJSON {
			t.Errorf("%s (id=%d, seq=%d, type=%s):\n  期望: %s\n  实际: %s",
				result.Name, result.ID, result.Seq, result.StepType, expJSON, result.PreStepIDs)
		}
	}
}

func TestComputeInstancePreStepIDs_Instance91(t *testing.T) {
	db := setupTestDB(t)
	origDB := repository.DB
	repository.DB = db
	defer func() { repository.DB = origDB }()

	steps := []entity.StepInstance{
		{ID: 468, DrillInstanceID: 91, StepTemplateID: 284, Name: "A", Seq: 1, StepType: "serial", AssigneeIDs: "[]"},
		{ID: 469, DrillInstanceID: 91, StepTemplateID: 285, Name: "B", Seq: 2, StepType: "parallel", AssigneeIDs: "[]"},
		{ID: 470, DrillInstanceID: 91, ParentStepID: ptrUint64(469), StepTemplateID: 286, Name: "B-1", Seq: 3, StepType: "serial", AssigneeIDs: "[]"},
		{ID: 471, DrillInstanceID: 91, ParentStepID: ptrUint64(469), StepTemplateID: 287, Name: "B-2", Seq: 4, StepType: "serial", AssigneeIDs: "[]"},
		{ID: 472, DrillInstanceID: 91, StepTemplateID: 288, Name: "C", Seq: 5, StepType: "parallel", AssigneeIDs: "[]"},
		{ID: 473, DrillInstanceID: 91, ParentStepID: ptrUint64(472), StepTemplateID: 289, Name: "C-1", Seq: 6, StepType: "serial", AssigneeIDs: "[]"},
		{ID: 474, DrillInstanceID: 91, StepTemplateID: 290, Name: "D", Seq: 7, StepType: "serial", AssigneeIDs: "[]"},
		{ID: 475, DrillInstanceID: 91, StepTemplateID: 291, Name: "E", Seq: 8, StepType: "serial", AssigneeIDs: "[]"},
		{ID: 476, DrillInstanceID: 91, StepTemplateID: 292, Name: "H", Seq: 9, StepType: "serial", AssigneeIDs: "[]"},
		{ID: 477, DrillInstanceID: 91, StepTemplateID: 293, Name: "I", Seq: 10, StepType: "serial", AssigneeIDs: "[]"},
		{ID: 478, DrillInstanceID: 91, StepTemplateID: 294, Name: "J", Seq: 11, StepType: "serial", AssigneeIDs: "[]"},
		{ID: 479, DrillInstanceID: 91, ParentStepID: ptrUint64(478), StepTemplateID: 295, Name: "J-1", Seq: 12, StepType: "serial", AssigneeIDs: "[]"},
		{ID: 480, DrillInstanceID: 91, StepTemplateID: 296, Name: "K", Seq: 13, StepType: "serial", AssigneeIDs: "[]"},
		{ID: 481, DrillInstanceID: 91, StepTemplateID: 297, Name: "L", Seq: 14, StepType: "parallel", AssigneeIDs: "[]"},
		{ID: 482, DrillInstanceID: 91, StepTemplateID: 298, Name: "M", Seq: 15, StepType: "parallel", AssigneeIDs: "[]"},
		{ID: 483, DrillInstanceID: 91, StepTemplateID: 299, Name: "N", Seq: 16, StepType: "serial", AssigneeIDs: "[]"},
	}

	for i := range steps {
		steps[i].PreStepIDs = ""
		if err := db.Create(&steps[i]).Error; err != nil {
			t.Fatalf("插入步骤 %s 失败: %v", steps[i].Name, err)
		}
	}

	svc := &DrillService{}
	svc.computeInstancePreStepIDs(steps, nil)

	var results []entity.StepInstance
	db.Order("seq asc").Find(&results)

	// B 和 C 是并行组 → C 依赖 A（与 B 一致），D 等待 [B, C]
	// L 和 M 是并行组 → L 和 M 都依赖 K，N 等待 [L, M]
	expected := map[uint64][]uint64{
		468: {},
		469: {468},
		470: {468},
		471: {470},
		472: {468},
		473: {468},
		474: {469, 472},
		475: {474},
		476: {475},
		477: {476},
		478: {477},
		479: {477},
		480: {478},
		481: {480},
		482: {480},
		483: {481, 482},
	}

	for _, result := range results {
		exp := expected[result.ID]
		expJSON := preIDsToJSON(exp)
		if result.PreStepIDs != expJSON {
			t.Errorf("%s (id=%d, seq=%d, type=%s):\n  期望: %s\n  实际: %s",
				result.Name, result.ID, result.Seq, result.StepType, expJSON, result.PreStepIDs)
		}
	}
}

func TestComputeInstancePreStepIDs_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() []entity.StepInstance
		expected map[uint64][]uint64
	}{
		{
			name:     "空步骤列表",
			setup:    func() []entity.StepInstance { return []entity.StepInstance{} },
			expected: map[uint64][]uint64{},
		},
		{
			name: "单个步骤",
			setup: func() []entity.StepInstance {
				return []entity.StepInstance{
					{ID: 1, DrillInstanceID: 1, StepTemplateID: 1, Name: "单步", Seq: 1, StepType: "serial", AssigneeIDs: "[]"},
				}
			},
			expected: map[uint64][]uint64{1: {}},
		},
		{
			name: "两个串行步骤",
			setup: func() []entity.StepInstance {
				return []entity.StepInstance{
					{ID: 1, DrillInstanceID: 2, StepTemplateID: 1, Name: "第一步", Seq: 1, StepType: "serial", AssigneeIDs: "[]"},
					{ID: 2, DrillInstanceID: 2, StepTemplateID: 2, Name: "第二步", Seq: 2, StepType: "serial", AssigneeIDs: "[]"},
				}
			},
			expected: map[uint64][]uint64{1: {}, 2: {1}},
		},
		{
			name: "连续并行组无子步骤",
			setup: func() []entity.StepInstance {
				return []entity.StepInstance{
					{ID: 1, DrillInstanceID: 3, StepTemplateID: 1, Name: "开始", Seq: 1, StepType: "serial", AssigneeIDs: "[]"},
					{ID: 2, DrillInstanceID: 3, StepTemplateID: 2, Name: "并行A", Seq: 2, StepType: "parallel", AssigneeIDs: "[]"},
					{ID: 3, DrillInstanceID: 3, StepTemplateID: 3, Name: "并行B", Seq: 3, StepType: "parallel", AssigneeIDs: "[]"},
					{ID: 4, DrillInstanceID: 3, StepTemplateID: 4, Name: "结束", Seq: 4, StepType: "serial", AssigneeIDs: "[]"},
				}
			},
			expected: map[uint64][]uint64{
				1: {}, 2: {1}, 3: {1}, 4: {2, 3},
			},
		},
		{
			name: "非连续并行组合子步骤",
			setup: func() []entity.StepInstance {
				return []entity.StepInstance{
					{ID: 1, DrillInstanceID: 4, StepTemplateID: 1, Name: "开始", Seq: 1, StepType: "serial", AssigneeIDs: "[]"},
					{ID: 2, DrillInstanceID: 4, StepTemplateID: 2, Name: "并行A", Seq: 2, StepType: "parallel", AssigneeIDs: "[]"},
					{ID: 3, DrillInstanceID: 4, ParentStepID: ptrUint64(2), StepTemplateID: 3, Name: "A-1", Seq: 3, StepType: "serial", AssigneeIDs: "[]"},
					{ID: 4, DrillInstanceID: 4, ParentStepID: ptrUint64(2), StepTemplateID: 4, Name: "A-2", Seq: 4, StepType: "serial", AssigneeIDs: "[]"},
					{ID: 5, DrillInstanceID: 4, StepTemplateID: 5, Name: "并行B", Seq: 5, StepType: "parallel", AssigneeIDs: "[]"},
					{ID: 6, DrillInstanceID: 4, ParentStepID: ptrUint64(5), StepTemplateID: 6, Name: "B-1", Seq: 6, StepType: "serial", AssigneeIDs: "[]"},
					{ID: 7, DrillInstanceID: 4, StepTemplateID: 7, Name: "汇合", Seq: 7, StepType: "serial", AssigneeIDs: "[]"},
				}
			},
			expected: map[uint64][]uint64{
				1: {}, 2: {1}, 3: {1}, 4: {3}, 5: {1}, 6: {1}, 7: {2, 5},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			origDB := repository.DB
			repository.DB = db
			defer func() { repository.DB = origDB }()

			steps := tt.setup()
			for i := range steps {
				steps[i].PreStepIDs = ""
				if err := db.Create(&steps[i]).Error; err != nil {
					t.Fatalf("插入步骤失败: %v", err)
				}
			}

			svc := &DrillService{}
			svc.computeInstancePreStepIDs(steps, nil)

			var results []entity.StepInstance
			db.Order("seq asc").Find(&results)

			for _, result := range results {
				exp := tt.expected[result.ID]
				expJSON := preIDsToJSON(exp)
				if result.PreStepIDs != expJSON {
					t.Errorf("%s (id=%d): 期望=%s, 实际=%s", result.Name, result.ID, expJSON, result.PreStepIDs)
				}
			}
		})
	}
}
