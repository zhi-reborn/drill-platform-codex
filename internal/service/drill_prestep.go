package service

import (
	"encoding/json"
	"sort"

	"drill-platform/internal/domain/entity"
	"drill-platform/internal/repository"
)

type rawStep struct {
	instanceID uint64
	tplID      uint64
	seq        int
	stepType   string
	parentID   *uint64
}

func (s *DrillService) computeInstancePreStepIDs(instanceSteps []entity.StepInstance, tplSteps []entity.StepTemplate) {
	if len(instanceSteps) == 0 {
		return
	}

	instMap := make(map[uint64]*rawStep)
	tplIDToInst := make(map[uint64]uint64)
	tplMap := make(map[uint64]entity.StepTemplate)

	for i := range tplSteps {
		tplMap[tplSteps[i].ID] = tplSteps[i]
	}
	for i := range instanceSteps {
		tplIDToInst[instanceSteps[i].StepTemplateID] = instanceSteps[i].ID
	}

	internalPre := make(map[uint64][]uint64)
	children := make(map[uint64][]*rawStep)
	rootSteps := make([]*rawStep, 0)

	for i := range instanceSteps {
		si := &instanceSteps[i]
		tpl := tplMap[si.StepTemplateID]

		rs := &rawStep{
			instanceID: si.ID,
			tplID:      si.StepTemplateID,
			seq:        si.Seq,
			stepType:   si.StepType,
			parentID:   si.ParentStepID,
		}
		instMap[rs.instanceID] = rs

		if tpl.PreStepIDs != "" && tpl.PreStepIDs != "null" {
			var tplIDs []uint64
			if json.Unmarshal([]byte(tpl.PreStepIDs), &tplIDs) == nil {
				for _, tid := range tplIDs {
					if instID, ok := tplIDToInst[tid]; ok {
						internalPre[rs.instanceID] = append(internalPre[rs.instanceID], instID)
					}
				}
			}
		}

		if si.ParentStepID != nil && *si.ParentStepID > 0 {
			children[*si.ParentStepID] = append(children[*si.ParentStepID], rs)
		} else {
			rootSteps = append(rootSteps, rs)
		}
	}

	seqGroups := make(map[int][]*rawStep)
	for _, rs := range rootSteps {
		seqGroups[rs.seq] = append(seqGroups[rs.seq], rs)
	}
	keys := sortedKeys(seqGroups)

	groupIDs := make(map[int][]uint64)
	var collect func(rs *rawStep) []uint64
	collect = func(rs *rawStep) []uint64 {
		ids := []uint64{rs.instanceID}
		for _, ch := range children[rs.instanceID] {
			ids = append(ids, collect(ch)...)
		}
		return ids
	}
	for seq, roots := range seqGroups {
		for _, rs := range roots {
			groupIDs[seq] = append(groupIDs[seq], collect(rs)...)
		}
	}

	writePre := func(id uint64, ids []uint64) {
		b, _ := json.Marshal(ids)
		repository.DB.Model(&entity.StepInstance{}).Where("id = ?", id).Update("pre_step_ids", string(b))
	}

	for i := 1; i < len(keys); i++ {
		prevIDs := groupIDs[keys[i-1]]
		for _, rs := range seqGroups[keys[i]] {
			s.assignPreStepIDs(rs, prevIDs, internalPre, children, writePre)
		}
	}
}

// assignPreStepIDs 递归设置步骤及其后代的 pre_step_ids。
// 决策：若步骤无同组内部依赖（模板 PreStepIDs 全在 prevIDs 中或为空），
// 则仅依赖 prevIDs；若有同组内部前驱，则合并内部前驱与 prevIDs。
func (s *DrillService) assignPreStepIDs(
	rs *rawStep,
	prevIDs []uint64,
	internalPre map[uint64][]uint64,
	children map[uint64][]*rawStep,
	writePre func(uint64, []uint64),
) {
	mapped := internalPre[rs.instanceID]

	var result []uint64
	hasInternalDep := false
	for _, mid := range mapped {
		if !sliceContains(prevIDs, mid) {
			hasInternalDep = true
			break
		}
	}

	if hasInternalDep {
		result = append([]uint64{}, mapped...)
		for _, pid := range prevIDs {
			if !sliceContains(result, pid) {
				result = append(result, pid)
			}
		}
	} else {
		result = prevIDs
	}

	writePre(rs.instanceID, result)

	for _, ch := range children[rs.instanceID] {
		s.assignPreStepIDs(ch, prevIDs, internalPre, children, writePre)
	}
}

func sortedKeys[T any](m map[int]T) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}

func sliceContains(slice []uint64, val uint64) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}