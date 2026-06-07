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
		for i := 0; i < len(siblings); {
			id := siblings[i]
			if all[id].stepType == "parallel" {
				j := i + 1
				for j < len(siblings) && all[siblings[j]].stepType == "parallel" {
					j++
				}
				groupIDs := make([]uint64, 0, j-i)
				for _, gid := range siblings[i:j] {
					computed[gid] = copyIDs(inherited)
					computeLevel(gid, computed[gid])
					groupIDs = append(groupIDs, gid)
				}
				inherited = groupIDs
				i = j
				continue
			}

			computed[id] = copyIDs(inherited)
			computeLevel(id, computed[id])
			inherited = []uint64{id}
			i++
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
