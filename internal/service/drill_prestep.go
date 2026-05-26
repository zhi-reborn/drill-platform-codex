package service

import (
	"encoding/json"
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
		p := &stepInfo{id: si.ID, seq: si.Seq, stepType: si.StepType}
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
		if all[roots[i]].stepType != "parallel" {
			continue
		}
		j := i + 1
		for j < len(roots) && all[roots[j]].stepType == "parallel" {
			j++
		}
		consecutive := true
		for k := i; k < j-1; k++ {
			if all[roots[k+1]].seq-all[roots[k]].seq != 1 {
				consecutive = false
				break
			}
		}
		g := &group{consecutive: consecutive && j > i+1}
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

		if all[rid].stepType == "parallel" {
			if cur == nil {
				cur = &wave{groups: []*group{g}, grouped: g.consecutive}
			} else if cur.grouped && g.consecutive {
				cur.groups = append(cur.groups, g)
			} else {
				waves = append(waves, cur)
				cur = &wave{groups: []*group{g}, grouped: g.consecutive}
			}
		} else {
			if cur != nil {
				waves = append(waves, cur)
			}
			cur = &wave{groups: []*group{g}, grouped: false}
			waves = append(waves, cur)
			cur = nil
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
				if prevWave.grouped {
					for _, g := range prevWave.groups {
						result = append(result, g.rootIDs[len(g.rootIDs)-1])
					}
				} else {
					for _, g := range prevWave.groups {
						result = append(result, g.rootIDs...)
					}
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
