package service

import (
	"encoding/json"
	"log"
	"sort"

	"drill-platform/internal/domain/entity"
	"drill-platform/internal/repository"
)

type stepInfo struct {
	id       uint64
	seq      int
	stepType string
	parentID uint64
}

func (s *DrillService) computeInstancePreStepIDs(instanceSteps []entity.StepInstance, _ []entity.StepTemplate) {
	if len(instanceSteps) == 0 {
		return
	}

	all := make(map[uint64]*stepInfo)
	childrenOf := make(map[uint64][]uint64)
	roots := make([]uint64, 0)

	for i := range instanceSteps {
		si := &instanceSteps[i]
		p := &stepInfo{
			id: si.ID, seq: si.Seq, stepType: si.StepType,
		}
		if si.ParentStepID != nil && *si.ParentStepID > 0 {
			p.parentID = *si.ParentStepID
			childrenOf[p.parentID] = append(childrenOf[p.parentID], p.id)
		} else {
			roots = append(roots, p.id)
		}
		all[p.id] = p
	}

	sort.Slice(roots, func(i, j int) bool { return all[roots[i]].seq < all[roots[j]].seq })

	writePre := func(id uint64, ids []uint64) {
		if ids == nil {
			ids = []uint64{}
		}
		b, _ := json.Marshal(ids)
		repository.DB.Model(&entity.StepInstance{}).Where("id = ?", id).Update("pre_step_ids", string(b))
	}

	type group struct {
		rootIDs     []uint64
		consecutive bool
	}
	groupOf := make(map[uint64]*group)
	processed := make(map[uint64]bool)

	for i := 0; i < len(roots); i++ {
		ri := all[roots[i]]
		if ri.stepType != "parallel" {
			continue
		}
		j := i + 1
		for j < len(roots) && all[roots[j]].stepType == "parallel" {
			j++
		}
		consecutive := j > i+1
		for k := i; k < j-1; k++ {
			if all[roots[k+1]].seq-all[roots[k]].seq != 1 {
				consecutive = false
				break
			}
		}
		g := &group{consecutive: consecutive}
		for k := i; k < j; k++ {
			g.rootIDs = append(g.rootIDs, roots[k])
			groupOf[roots[k]] = g
		}
		i = j - 1
	}

	for _, rid := range roots {
		if _, ok := groupOf[rid]; !ok {
			groupOf[rid] = &group{rootIDs: []uint64{rid}}
		}
	}

	type wave struct {
		groups  []*group
		grouped bool
	}
	waves := []*wave{}
	var cur *wave

	for _, rid := range roots {
		if processed[rid] {
			continue
		}
		g := groupOf[rid]
		for _, mid := range g.rootIDs {
			processed[mid] = true
		}
		ri := all[rid]

		isNewWave := false
		if cur != nil && ri.stepType != "parallel" {
			isNewWave = true
		}

		if isNewWave || cur == nil {
			if cur != nil {
				waves = append(waves, cur)
			}
			cur = &wave{groups: []*group{g}, grouped: g.consecutive}
			if ri.stepType != "parallel" {
				waves = append(waves, cur)
				cur = nil
			}
		} else {
			cur.groups = append(cur.groups, g)
		}
	}
	if cur != nil {
		waves = append(waves, cur)
	}

	computed := make(map[uint64][]uint64)

	var resolve func(id uint64, waveIdx int) []uint64
	resolve = func(id uint64, waveIdx int) []uint64 {
		if v, ok := computed[id]; ok {
			return v
		}
		p := all[id]

		var result []uint64
		if p.parentID > 0 {
			siblings := childrenOf[p.parentID]
			sort.Slice(siblings, func(i, j int) bool { return all[siblings[i]].seq < all[siblings[j]].seq })
			if siblings[0] == p.id {
				result = resolve(p.parentID, waveIdx)
			} else {
				for idx := range siblings {
					if siblings[idx] == p.id && idx > 0 {
						result = append(result, siblings[idx-1])
						break
					}
				}
			}
		} else {
			if waveIdx == 0 {
				result = nil
			} else {
				prevWave := waves[waveIdx-1]
				for _, g := range prevWave.groups {
					result = append(result, g.rootIDs...)
				}
			}
		}

		computed[id] = result
		return result
	}

	waveOf := make(map[uint64]int)
	for wi, w := range waves {
		for _, g := range w.groups {
			for _, rid := range g.rootIDs {
				waveOf[rid] = wi
			}
		}
	}

	for _, rid := range roots {
		wi := waveOf[rid]
		writePre(rid, resolve(rid, wi))
		for _, ch := range childrenOf[rid] {
			var walk func(cid uint64)
			walk = func(cid uint64) {
				writePre(cid, resolve(cid, wi))
				for _, gch := range childrenOf[cid] {
					walk(gch)
				}
			}
			walk(ch)
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
