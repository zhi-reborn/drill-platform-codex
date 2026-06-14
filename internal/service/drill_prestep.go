package service

import (
	"encoding/json"
	"log"
	"sort"

	"drill-platform/internal/domain/entity"
	"drill-platform/internal/repository"

	"gorm.io/gorm"
)

type stepInfo struct {
	id       uint64
	seq      int
	stepType string
	parentID uint64
}

func (s *DrillService) computeInstancePreStepIDs(instanceSteps []entity.StepInstance, _ []entity.StepTemplate) {
	s.computeInstancePreStepIDsTx(instanceSteps, nil, repository.DB)
}

func (s *DrillService) computeInstancePreStepIDsTx(instanceSteps []entity.StepInstance, _ []entity.StepTemplate, db *gorm.DB) {
	if len(instanceSteps) == 0 {
		return
	}

	all := make(map[uint64]*stepInfo)
	childrenOf := make(map[uint64][]uint64)
	const rootParentID uint64 = 0

	for i := range instanceSteps {
		si := &instanceSteps[i]
		p := &stepInfo{
			id: si.ID, seq: si.Seq, stepType: si.StepType,
		}
		if si.ParentStepID != nil && *si.ParentStepID > 0 {
			p.parentID = *si.ParentStepID
			childrenOf[p.parentID] = append(childrenOf[p.parentID], p.id)
		} else {
			childrenOf[rootParentID] = append(childrenOf[rootParentID], p.id)
		}
		all[p.id] = p
	}

	for parentID := range childrenOf {
		ids := childrenOf[parentID]
		sort.Slice(ids, func(i, j int) bool { return all[ids[i]].seq < all[ids[j]].seq })
		childrenOf[parentID] = ids
	}

	writePre := func(id uint64, ids []uint64) {
		if ids == nil {
			ids = []uint64{}
		}
		b, _ := json.Marshal(ids)
		db.Model(&entity.StepInstance{}).Where("id = ?", id).Update("pre_step_ids", string(b))
	}

	computed := make(map[uint64][]uint64)

	copyIDs := func(ids []uint64) []uint64 {
		if len(ids) == 0 {
			return nil
		}
		out := make([]uint64, len(ids))
		copy(out, ids)
		return out
	}

	var computeLevel func(parentID uint64, inherited []uint64)
	computeLevel = func(parentID uint64, inherited []uint64) {
		siblings := childrenOf[parentID]

		// 核心规则：
		// - parallel 步骤的前序继承当前同级段的前序，同一段内可同时开始
		// - serial 步骤的前序包含前一个步骤/并行组
		// - 连续的 parallel 步骤组成并行组，组完成后成为后续 serial 步骤的前序
		// - 子步骤继承父节点自身的前序，避免绕过父节点启动条件

		// 第一遍：识别段（连续 parallel 为一段，单个 serial 为一段）
		type segment struct {
			startIdx   int
			endIdx     int // exclusive
			isParallel bool
		}
		var segments []segment
		for i := 0; i < len(siblings); {
			if all[siblings[i]].stepType == "parallel" {
				j := i + 1
				for j < len(siblings) && all[siblings[j]].stepType == "parallel" {
					j++
				}
				segments = append(segments, segment{startIdx: i, endIdx: j, isParallel: true})
				i = j
			} else {
				segments = append(segments, segment{startIdx: i, endIdx: i + 1, isParallel: false})
				i++
			}
		}

		// 第二遍：计算前序
		// parallel 段：所有步骤的 pre = currentInherited
		// serial 段：pre = 前一个段完成后的前序
		currentInherited := copyIDs(inherited)
		for _, seg := range segments {
			if seg.isParallel {
				groupIDs := make([]uint64, 0, seg.endIdx-seg.startIdx)
				for k := seg.startIdx; k < seg.endIdx; k++ {
					gid := siblings[k]
					if parentID != rootParentID && all[parentID] != nil && all[parentID].stepType == "parallel" {
						computed[gid] = copyIDs(inherited)
					} else {
						computed[gid] = copyIDs(currentInherited)
					}
					computeLevel(gid, computed[gid])
					groupIDs = append(groupIDs, gid)
				}
				currentInherited = groupIDs
			} else {
				id := siblings[seg.startIdx]
				computed[id] = copyIDs(currentInherited)
				computeLevel(id, computed[id])
				currentInherited = []uint64{id}
			}
		}
	}

	computeLevel(rootParentID, nil)

	for id, ids := range computed {
		writePre(id, ids)
	}

	for id := range all {
		if _, ok := computed[id]; !ok {
			writePre(id, nil)
		}
	}
}

func (s *DrillService) syncPreStepIDsToEngine(flowInstID int64) {
	inst, ok := s.engine.GetInstance(flowInstID)
	if !ok {
		return
	}

	var steps []entity.StepInstance
	repository.DB.Where("drill_instance_id = ?", flowInstID).Find(&steps)

	instIDToDefID := make(map[uint64]int64)
	instIDToPres := make(map[int64][]int64)
	for _, step := range steps {
		defID := int64(step.StepTemplateID)
		instIDToDefID[step.ID] = defID

		if step.PreStepIDs == "" || step.PreStepIDs == "[]" || step.PreStepIDs == "null" {
			// 数据库中 pre_step_ids 为空，也需要同步到引擎（覆盖模板中错误的 seq 值）
			instIDToPres[defID] = nil
			continue
		}
		var ids []uint64
		if json.Unmarshal([]byte(step.PreStepIDs), &ids) == nil {
			for _, iid := range ids {
				if did, ok := instIDToDefID[iid]; ok {
					instIDToPres[defID] = append(instIDToPres[defID], did)
				}
			}
		}
	}

	for defID, preDefIDs := range instIDToPres {
		if si, exists := inst.Steps[defID]; exists {
			si.PreStepIDs = preDefIDs
			log.Printf("[SYNC] PreStepIDs: step=%d(%s) -> %v", defID, si.Name, preDefIDs)
		}
	}
}
