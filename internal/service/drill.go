package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"drill-platform/internal/domain/dto"
	"drill-platform/internal/domain/entity"
	"drill-platform/internal/infrastructure/websocket"
	"drill-platform/internal/pkg/flowengine"
	"drill-platform/internal/repository"

	"gorm.io/gorm"
)

type DrillService struct {
	drillRepo           *repository.DrillRepo
	templateRepo        *repository.TemplateRepo
	stepRepo            *repository.StepRepo
	userRepo            *repository.UserRepo
	engine              *flowengine.Engine
	adapter             *DrillFlowAdapter
	wsManager           *websocket.Manager
	notificationService *NotificationService
	redis               RedisClient
}

func NewDrillService(drillRepo *repository.DrillRepo, templateRepo *repository.TemplateRepo, stepRepo *repository.StepRepo, userRepo *repository.UserRepo) *DrillService {
	return &DrillService{
		drillRepo:    drillRepo,
		templateRepo: templateRepo,
		stepRepo:     stepRepo,
		userRepo:     userRepo,
	}
}

func (s *DrillService) SetRedis(redis RedisClient) {
	s.redis = redis
}

func (s *DrillService) SetEngine(engine *flowengine.Engine, adapter *DrillFlowAdapter) {
	s.engine = engine
	s.adapter = adapter
}

func (s *DrillService) Engine() *flowengine.Engine {
	return s.engine
}

func (s *DrillService) Recover(id uint64) error {
	drill, err := s.drillRepo.FindByID(id)
	if err != nil {
		return err
	}

	template, err := s.templateRepo.FindByID(drill.TemplateID)
	if err != nil || template == nil {
		return errors.New("关联模板不存在")
	}

	flowDef := s.adapter.BuildFlowDef(template)
	flowDef.ID = int64(drill.ID)
	assignees := s.adapter.BuildAssignees(drill.ID)

	inst, err := s.engine.CreateInstance(flowDef, assignees, int64(drill.CreatedBy))
	if err != nil {
		return err
	}
	inst.Status = flowengine.FlowStatus(drill.Status)

	steps, err := s.stepRepo.FindStepsByDrillID(id)
	if err != nil {
		return err
	}

	if err := s.backfillMissingStepTemplateIDs(drill.ID, template.Steps, steps); err != nil {
		return err
	}

	// 同步内存中步骤实例的 ID 到数据库 ID
	s.adapter.SyncStepInstanceIDs(int64(drill.ID))

	// 重新计算前序步骤 ID（确保使用最新算法）
	s.computeInstancePreStepIDsTx(steps, template.Steps, repository.DB)

	// 同步前序步骤 ID（将实例 ID 转换为模板步骤 ID）
	s.syncPreStepIDsToEngine(int64(drill.ID))

	for _, step := range steps {
		si, exists := inst.Steps[int64(step.StepTemplateID)]
		if exists {
			si.Status = flowengine.StepStatus(step.Status)
			si.StartTime = step.StartTime
			si.EndTime = step.EndTime
			si.TimeoutAt = step.TimeoutAt
			si.Remark = step.Remark
			si.IssueDesc = step.IssueDesc
			if step.ActualOperator != nil {
				op := int64(*step.ActualOperator)
				si.ActualOperator = &op
			}

			if step.Status == "running" && step.TimeoutAt != nil {
				s.engine.TimeoutScheduler().Register(int64(drill.ID), int64(step.StepTemplateID), int64(step.ID), *step.TimeoutAt)
			}
		}
	}

	// 协调状态：自动完成所有子步骤已终态但自身未终态的父步骤
	s.reconcileParentSteps(int64(drill.ID), steps)

	// 协调状态：激活所有前序步骤已完成但自身仍为 pending 的步骤
	// 对每个已终态的步骤调用 AdvanceFlow，触发 handleStepCompletion 推进流程
	if drill.Status == "running" {
		for _, step := range steps {
			if step.Status == "completed" || step.Status == "skipped" ||
				step.Status == "timeout" || step.Status == "issue" {
				s.engine.AdvanceFlow(int64(drill.ID), int64(step.StepTemplateID))
			}
		}
	}

	return nil
}

// reconcileParentSteps 检查并自动完成所有子步骤已终态但自身未终态的父步骤
// 从最深层开始向上逐层处理，确保多层嵌套时祖先节点也能正确完成
func (s *DrillService) reconcileParentSteps(flowInstID int64, steps []entity.StepInstance) {
	inst, ok := s.engine.GetInstanceForMutate(flowInstID)
	if !ok {
		return
	}

	terminalStatuses := map[string]bool{
		string(flowengine.StepStatusCompleted): true,
		string(flowengine.StepStatusSkipped):   true,
		string(flowengine.StepStatusTimeout):   true,
		string(flowengine.StepStatusIssue):     true,
	}

	// 计算每个步骤的深度（从根到叶）
	stepMap := make(map[int64]*flowengine.StepInst)
	for id, si := range inst.Steps {
		stepMap[id] = si
	}

	depth := make(map[int64]int)
	var calcDepth func(int64) int
	calcDepth = func(stepDefID int64) int {
		if d, ok := depth[stepDefID]; ok {
			return d
		}
		si, exists := stepMap[stepDefID]
		if !exists || si.ParentStepID == 0 {
			depth[stepDefID] = 0
			return 0
		}
		d := calcDepth(si.ParentStepID) + 1
		depth[stepDefID] = d
		return d
	}
	maxDepth := 0
	for id := range stepMap {
		d := calcDepth(id)
		if d > maxDepth {
			maxDepth = d
		}
	}

	// 从最深层开始，逐层向上处理
	for d := maxDepth; d >= 0; d-- {
		for stepDefID, si := range stepMap {
			if depth[stepDefID] != d {
				continue
			}
			// 跳过已终态的步骤
			if si.Status == flowengine.StepStatusCompleted ||
				si.Status == flowengine.StepStatusSkipped ||
				si.Status == flowengine.StepStatusTimeout ||
				si.Status == flowengine.StepStatusIssue {
				continue
			}

			// 检查是否有子步骤
			var childIDs []int64
			for cid, csi := range stepMap {
				if csi.ParentStepID == stepDefID {
					childIDs = append(childIDs, cid)
				}
			}
			if len(childIDs) == 0 {
				continue // 叶子步骤，跳过
			}

			// 检查所有子步骤是否已终态
			allChildrenTerminal := true
			for _, cid := range childIDs {
				csi, exists := stepMap[cid]
				if !exists || !terminalStatuses[string(csi.Status)] {
					allChildrenTerminal = false
					break
				}
			}

			if !allChildrenTerminal {
				continue
			}

			// 所有子步骤已终态，自动完成该父步骤
			now := time.Now()
			si.Status = flowengine.StepStatusCompleted
			si.EndTime = &now

			// 更新数据库
			for _, step := range steps {
				if step.StepTemplateID == uint64(stepDefID) {
					repository.DB.Model(&entity.StepInstance{}).Where("id = ?", step.ID).Updates(map[string]interface{}{
						"status":   string(flowengine.StepStatusCompleted),
						"end_time": &now,
					})
					break
				}
			}

			// 写日志
			var stepName string
			for _, step := range steps {
				if step.StepTemplateID == uint64(stepDefID) {
					stepName = step.Name
					break
				}
			}
			repository.DB.Create(&entity.DrillInstanceLog{
				DrillInstanceID: uint64(flowInstID),
				Action:          "auto_complete",
				OperatorName:    "流程引擎",
				Content:         fmt.Sprintf("[%s] 所有子步骤已终态，自动完成（恢复协调）", stepName),
				CreatedAt:       now,
			})

			// 恢复协调不自动推进流程，只修正状态
		}
	}
}

func (s *DrillService) backfillMissingStepTemplateIDs(drillID uint64, templateSteps []entity.StepTemplate, steps []entity.StepInstance) error {
	bySeqName := make(map[string]uint64, len(templateSteps))
	bySeq := make(map[int]uint64, len(templateSteps))
	templateIDs := make(map[uint64]struct{}, len(templateSteps))
	for _, step := range templateSteps {
		bySeqName[stepTemplateKey(step.Seq, step.Name)] = step.ID
		if _, exists := bySeq[step.Seq]; !exists {
			bySeq[step.Seq] = step.ID
		}
		templateIDs[step.ID] = struct{}{}
	}

	for i := range steps {
		// 已存在且指向有效模板步骤则跳过；否则按 seq+name 重新匹配
		if _, ok := templateIDs[steps[i].StepTemplateID]; ok {
			continue
		}
		stepTemplateID, exists := bySeqName[stepTemplateKey(steps[i].Seq, steps[i].Name)]
		if !exists {
			stepTemplateID, exists = bySeq[steps[i].Seq]
		}
		if !exists {
			return errors.New("步骤模板映射不存在")
		}
		steps[i].StepTemplateID = stepTemplateID
		if err := repository.DB.Model(&entity.StepInstance{}).
			Where("drill_instance_id = ? AND id = ?", drillID, steps[i].ID).
			Update("template_step_id", stepTemplateID).Error; err != nil {
			return err
		}
	}

	return nil
}

func stepTemplateKey(seq int, name string) string {
	return strconv.Itoa(seq) + "\x00" + name
}

func (s *DrillService) SetWebSocketManager(wsManager *websocket.Manager) {
	s.wsManager = wsManager
}

func (s *DrillService) SetNotificationService(ns *NotificationService) {
	s.notificationService = ns
}

func (s *DrillService) GetList(page, pageSize int, status string) ([]entity.DrillInstance, int64, error) {
	return s.drillRepo.List(page, pageSize, status)
}

func (s *DrillService) GetUserByID(id uint64) (*entity.User, error) {
	return s.userRepo.FindByID(id)
}

func (s *DrillService) GetUsersByIDs(ids []uint64) (map[uint64]*entity.User, error) {
	return s.userRepo.FindByIDs(ids)
}

func (s *DrillService) GetDetail(id uint64) (*entity.DrillInstance, error) {
	return s.drillRepo.FindByID(id)
}

func (s *DrillService) GetSteps(id uint64) ([]entity.StepInstance, error) {
	recovered, err := s.recoverRunningDrillIfNeeded(id)
	if err != nil {
		log.Printf("[GetSteps] recover drill %d failed (continue with db steps): %v", id, err)
	}
	if recovered {
		InvalidateStepCache(s.redis, id)
	}

	running := s.isRunningDrill(id)
	if !running {
		if steps, ok := GetCachedSteps(s.redis, id); ok {
			s.reconcilePreStepIDs(id, steps)
			return steps, nil
		}
	}

	steps, err := s.stepRepo.FindStepsByDrillID(id)
	if err != nil {
		return nil, err
	}
	if running && s.advanceRunningDrillFromTerminalSteps(id, steps) {
		InvalidateStepCache(s.redis, id)
		steps, err = s.stepRepo.FindStepsByDrillID(id)
		if err != nil {
			return nil, err
		}
	}

	if !running {
		SetCachedSteps(s.redis, id, steps)
	}

	// 协调父步骤状态：如果所有子步骤已终态但父步骤未终态，自动完成父步骤
	s.reconcileParentStepsFromDB(id, steps)

	// 协调 pre_step_ids：如果 parallel 步骤的 pre_step_ids 包含同级步骤，重新计算
	s.reconcilePreStepIDs(id, steps)

	return steps, nil
}

func (s *DrillService) isRunningDrill(id uint64) bool {
	if s.engine != nil {
		if inst, ok := s.engine.GetInstance(int64(id)); ok {
			return inst.Status == flowengine.FlowStatusRunning
		}
	}

	drill, err := s.drillRepo.FindByID(id)
	if err != nil {
		return false
	}
	return drill.Status == "running"
}

func (s *DrillService) recoverRunningDrillIfNeeded(id uint64) (bool, error) {
	if s.engine == nil {
		return false, nil
	}
	if _, ok := s.engine.GetInstance(int64(id)); ok {
		return false, nil
	}

	drill, err := s.drillRepo.FindByID(id)
	if err != nil {
		return false, err
	}
	if drill.Status != "running" {
		return false, nil
	}

	return true, s.Recover(id)
}

func (s *DrillService) advanceRunningDrillFromTerminalSteps(id uint64, steps []entity.StepInstance) bool {
	if s.engine == nil {
		return false
	}
	inst, ok := s.engine.GetInstanceForMutate(int64(id))
	if !ok || inst.Status != flowengine.FlowStatusRunning {
		return false
	}

	for _, step := range steps {
		if si, exists := inst.Steps[int64(step.StepTemplateID)]; exists {
			si.ID = int64(step.ID)
			si.Status = flowengine.StepStatus(step.Status)
			si.StartTime = step.StartTime
			si.EndTime = step.EndTime
			si.TimeoutAt = step.TimeoutAt
		}
	}

	before := make(map[int64]flowengine.StepStatus, len(inst.Steps))
	for stepDefID, si := range inst.Steps {
		before[stepDefID] = si.Status
	}

	for _, step := range steps {
		if isStepTerminalStatus(flowengine.StepStatus(step.Status)) {
			_ = s.engine.AdvanceFlow(int64(id), int64(step.StepTemplateID))
		}
	}

	for stepDefID, oldStatus := range before {
		if oldStatus == flowengine.StepStatusPending && inst.Steps[stepDefID].Status == flowengine.StepStatusRunning {
			return true
		}
	}
	return false
}

// reconcileParentStepsFromDB 基于 DB 数据协调父步骤状态，不依赖引擎内存
// 从最深层开始向上逐层处理，确保多层嵌套时祖先节点也能正确完成
func (s *DrillService) reconcileParentStepsFromDB(drillID uint64, steps []entity.StepInstance) {
	terminalStatuses := map[string]bool{
		"completed": true,
		"skipped":   true,
		"timeout":   true,
		"issue":     true,
	}

	// 构建步骤映射和父子关系
	stepMap := make(map[uint64]*entity.StepInstance, len(steps))
	childrenMap := make(map[uint64][]uint64) // parentID -> childIDs
	for i := range steps {
		stepMap[steps[i].ID] = &steps[i]
		if steps[i].ParentStepID != nil && *steps[i].ParentStepID > 0 {
			childrenMap[*steps[i].ParentStepID] = append(childrenMap[*steps[i].ParentStepID], steps[i].ID)
		}
	}

	// 计算每个步骤的深度
	depth := make(map[uint64]int)
	var calcDepth func(uint64) int
	calcDepth = func(stepID uint64) int {
		if d, ok := depth[stepID]; ok {
			return d
		}
		step := stepMap[stepID]
		if step == nil || step.ParentStepID == nil || *step.ParentStepID == 0 {
			depth[stepID] = 0
			return 0
		}
		d := calcDepth(*step.ParentStepID) + 1
		depth[stepID] = d
		return d
	}
	maxDepth := 0
	for _, step := range steps {
		d := calcDepth(step.ID)
		if d > maxDepth {
			maxDepth = d
		}
	}

	// 从最深层开始，逐层向上处理
	changed := false
	for d := maxDepth; d >= 0; d-- {
		for i := range steps {
			if depth[steps[i].ID] != d {
				continue
			}
			if terminalStatuses[steps[i].Status] {
				continue
			}
			childIDs := childrenMap[steps[i].ID]
			if len(childIDs) == 0 {
				continue // 叶子步骤
			}

			allChildrenTerminal := true
			for _, cid := range childIDs {
				child := stepMap[cid]
				if child == nil || !terminalStatuses[child.Status] {
					allChildrenTerminal = false
					break
				}
			}
			if !allChildrenTerminal {
				continue
			}

			// 自动完成父步骤
			now := time.Now()
			steps[i].Status = "completed"
			steps[i].EndTime = &now
			changed = true

			repository.DB.Model(&entity.StepInstance{}).Where("id = ?", steps[i].ID).Updates(map[string]interface{}{
				"status":   "completed",
				"end_time": &now,
			})

			repository.DB.Create(&entity.DrillInstanceLog{
				DrillInstanceID: drillID,
				StepInstanceID:  &steps[i].ID,
				Action:          "auto_complete",
				OperatorName:    "流程引擎",
				Content:         fmt.Sprintf("[%s] 所有子步骤已终态，自动完成", steps[i].Name),
				CreatedAt:       now,
			})
		}
	}

	if changed {
		InvalidateStepCache(s.redis, drillID)
	}
}

// reconcilePreStepIDs 检测子步骤的 pre_step_ids 是否符合当前层级执行规则
// 如果发现异常，重新计算所有 pre_step_ids
func (s *DrillService) reconcilePreStepIDs(drillID uint64, steps []entity.StepInstance) {
	// 构建兄弟关系：parentID -> set of childIDs
	siblingsOf := make(map[uint64]map[uint64]bool)
	for _, step := range steps {
		var parentID uint64
		if step.ParentStepID != nil && *step.ParentStepID > 0 {
			parentID = *step.ParentStepID
		}
		if siblingsOf[parentID] == nil {
			siblingsOf[parentID] = make(map[uint64]bool)
		}
		siblingsOf[parentID][step.ID] = true
	}

	stepMap := make(map[uint64]entity.StepInstance, len(steps))
	for _, step := range steps {
		stepMap[step.ID] = step
	}

	// 检查子步骤的 pre_step_ids 是否符合父级执行模式
	needsRecompute := false
	for _, step := range steps {
		var parentID uint64
		if step.ParentStepID != nil && *step.ParentStepID > 0 {
			parentID = *step.ParentStepID
		}
		siblings := siblingsOf[parentID]
		parent := stepMap[parentID]
		var preIDs []uint64
		if json.Unmarshal([]byte(step.PreStepIDs), &preIDs) == nil {
			for _, preID := range preIDs {
				if parentID > 0 && parent.StepType == "parallel" {
					// 并行父步骤下，parallel 子步骤不应等待同级兄弟步骤。
					if step.StepType == "parallel" && siblings[preID] {
						needsRecompute = true
						break
					}
				} else if !siblings[preID] {
					// 子步骤可以继承父节点自身的前序；其他跨层前序需要重算。
					if parentID == 0 || !preStepIDsContain(parent.PreStepIDs, preID) {
						needsRecompute = true
						break
					}
				}
			}
		}
		if needsRecompute {
			break
		}
	}

	if !needsRecompute {
		return
	}

	// 重新计算 pre_step_ids
	s.computeInstancePreStepIDsTx(steps, nil, repository.DB)

	// 重新读取更新后的数据
	updatedSteps, err := s.stepRepo.FindStepsByDrillID(drillID)
	if err == nil && len(updatedSteps) == len(steps) {
		for i := range steps {
			steps[i] = updatedSteps[i]
		}
	}

	InvalidateStepCache(s.redis, drillID)
}

func preStepIDsContain(raw string, id uint64) bool {
	var ids []uint64
	if json.Unmarshal([]byte(raw), &ids) != nil {
		return false
	}
	for _, currentID := range ids {
		if currentID == id {
			return true
		}
	}
	return false
}

func (s *DrillService) InvalidateStepCache(drillID uint64) {
	InvalidateStepCache(s.redis, drillID)
}

func (s *DrillService) EnrichStepsWithAssigneeNames(steps []entity.StepInstance) []entity.StepInstance {
	allIDs := make(map[uint64]bool)
	for i := range steps {
		if steps[i].AssigneeIDs == "" || steps[i].AssigneeIDs == "[]" {
			continue
		}
		var ids []uint64
		if json.Unmarshal([]byte(steps[i].AssigneeIDs), &ids) == nil {
			for _, id := range ids {
				allIDs[id] = true
			}
		}
	}
	if len(allIDs) == 0 {
		return steps
	}

	ids := make([]uint64, 0, len(allIDs))
	for id := range allIDs {
		ids = append(ids, id)
	}

	nameMap := GetCachedUserNames(s.redis, ids)
	if nameMap == nil {
		var users []entity.User
		repository.DB.Where("id IN ?", ids).Find(&users)
		nameMap = make(map[uint64]string, len(users))
		for _, u := range users {
			nameMap[u.ID] = u.RealName
		}
		SetCachedUserNames(s.redis, users)
	}

	for i := range steps {
		if steps[i].AssigneeIDs == "" || steps[i].AssigneeIDs == "[]" {
			continue
		}
		var ids []uint64
		if json.Unmarshal([]byte(steps[i].AssigneeIDs), &ids) == nil && len(ids) > 0 {
			var names []string
			for _, id := range ids {
				if n, ok := nameMap[id]; ok {
					names = append(names, n)
				}
			}
			steps[i].AssigneeNames = namesJoin(names)
		}
	}
	return steps
}

func namesJoin(names []string) string {
	result := ""
	for i, n := range names {
		if i > 0 {
			result += ", "
		}
		result += n
	}
	return result
}

func (s *DrillService) Create(req *dto.CreateDrillRequest, createdBy uint64) (*entity.DrillInstance, error) {
	template, err := s.templateRepo.FindByID(req.TemplateID)
	if err != nil {
		return nil, errors.New("模板不存在")
	}

	if template.Status == 0 {
		return nil, errors.New("模板已禁用，无法创建演练")
	}

	departments := make(map[string]bool)
	for _, stepTpl := range template.Steps {
		if _, ok := req.Assignees[stepTpl.ID]; ok {
			continue
		}
		if stepTpl.ExecutorTeam != "" {
			departments[stepTpl.ExecutorTeam] = true
		}
	}

	var deptUsers map[string][]entity.User
	if len(departments) > 0 {
		deptList := make([]string, 0, len(departments))
		for d := range departments {
			deptList = append(deptList, d)
		}
		deptUsers, _ = s.userRepo.FindByDepartments(deptList)
	}

	drill := &entity.DrillInstance{
		TemplateID:  req.TemplateID,
		Name:        req.Name,
		Description: req.Description,
		Status:      "pending",
		CreatedBy:   createdBy,
	}

	if req.PlannedStart != "" {
		t, err := time.Parse(time.RFC3339, req.PlannedStart)
		if err == nil {
			drill.PlannedStart = &t
		}
	}

	err = repository.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(drill).Error; err != nil {
			return err
		}

		for _, stepTpl := range template.Steps {
			assigneeIDs := "[]"
			if userIDs, ok := req.Assignees[stepTpl.ID]; ok && len(userIDs) > 0 {
				if bytes, _ := json.Marshal(userIDs); bytes != nil {
					assigneeIDs = string(bytes)
				}
			} else if stepTpl.ExecutorTeam != "" {
				if users, ok := deptUsers[stepTpl.ExecutorTeam]; ok && len(users) > 0 {
					userIDs := make([]uint64, len(users))
					for i, u := range users {
						userIDs[i] = u.ID
					}
					if bytes, _ := json.Marshal(userIDs); bytes != nil {
						assigneeIDs = string(bytes)
					}
				}
			}

			step := entity.StepInstance{
				DrillInstanceID:          drill.ID,
				StepTemplateID:           stepTpl.ID,
				Name:                     stepTpl.Name,
				Seq:                      stepTpl.Seq,
				Status:                   "pending",
				AssigneeIDs:              assigneeIDs,
				StepType:                 stepTpl.StepType,
				TimeoutMinutes:           stepTpl.TimeoutMinutes,
				ParentStepID:             nil,
				PreStepIDs:               "[]",
				Phase:                    stepTpl.Phase,
				PhaseStep:                stepTpl.PhaseStep,
				DefaultAssigneeRole:      stepTpl.DefaultAssigneeRole,
				ExecutorTeam:             stepTpl.ExecutorTeam,
				EstimatedDurationMinutes: stepTpl.EstimatedDurationMinutes,
				EstimatedStartOffset:     stepTpl.EstimatedStartOffset,
				JSONAttributes:           stepTpl.JSONAttributes,
			}
			if err := tx.Create(&step).Error; err != nil {
				log.Printf("[ERROR] Failed to create step instance for template step %d: %v", stepTpl.ID, err)
				return err
			}
		}

		var instanceSteps []entity.StepInstance
		if err := tx.Where("drill_instance_id = ?", drill.ID).Order("seq ASC").Find(&instanceSteps).Error; err != nil {
			return err
		}

		tplIDtoInstID := make(map[uint64]uint64)
		for _, si := range instanceSteps {
			tplIDtoInstID[si.StepTemplateID] = si.ID
		}

		tplStepMap := make(map[uint64]entity.StepTemplate)
		for _, s := range template.Steps {
			tplStepMap[s.ID] = s
		}
		// 批量更新 parent_step_id，避免 N 次单条 UPDATE
		type parentUpdate struct {
			ID           uint64
			ParentStepID uint64
		}
		var parentUpdates []parentUpdate
		for i := range instanceSteps {
			tpl := tplStepMap[instanceSteps[i].StepTemplateID]
			if tpl.ParentStepID != nil && *tpl.ParentStepID > 0 {
				if instParentID, ok := tplIDtoInstID[*tpl.ParentStepID]; ok {
					instanceSteps[i].ParentStepID = &instParentID
					parentUpdates = append(parentUpdates, parentUpdate{
						ID:           instanceSteps[i].ID,
						ParentStepID: instParentID,
					})
				}
			}
		}
		if len(parentUpdates) > 0 {
			caseSQL := "UPDATE `drill_instance_step` SET `parent_step_id` = CASE `id` "
			var args []interface{}
			var allIDs []uint64
			for _, u := range parentUpdates {
				caseSQL += "WHEN ? THEN ? "
				args = append(args, u.ID, u.ParentStepID)
				allIDs = append(allIDs, u.ID)
			}
			caseSQL += "ELSE `parent_step_id` END WHERE `id` IN ?"
			args = append(args, allIDs)
			if err := tx.Exec(caseSQL, args...).Error; err != nil {
				return err
			}
		}

		s.computeInstancePreStepIDsTx(instanceSteps, template.Steps, tx)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return s.drillRepo.FindByID(drill.ID)
}

func (s *DrillService) Start(id uint64) error {
	drill, err := s.drillRepo.FindByID(id)
	if err != nil {
		return err
	}
	if drill.Status != "pending" {
		return errors.New("只有待启动状态的演练才能开始")
	}

	if s.engine == nil {
		return errors.New("流程引擎未初始化")
	}

	template, err := s.templateRepo.FindByID(drill.TemplateID)
	if err != nil || template == nil {
		return errors.New("关联模板不存在")
	}

	if len(template.Steps) == 0 {
		return errors.New("模板没有步骤，无法启动")
	}

	s.adapter.RegisterDrillContext(int64(drill.ID), drillContext{
		ID:         drill.ID,
		Name:       drill.Name,
		Status:     "running",
		TemplateID: drill.TemplateID,
	})

	flowDef := s.adapter.BuildFlowDef(template)
	flowDef.ID = int64(drill.ID)
	assignees := s.adapter.BuildAssignees(drill.ID)

	_, err = s.engine.CreateInstance(flowDef, assignees, int64(drill.CreatedBy))
	if err != nil {
		return err
	}

	// 同步内存中步骤实例的 ID 到数据库 ID
	s.adapter.SyncStepInstanceIDs(int64(drill.ID))

	s.syncPreStepIDsToEngine(int64(drill.ID))

	// 查询操作人姓名
	var user entity.User
	operatorName := ""
	repository.DB.Model(&entity.User{}).Where("id = ?", drill.CreatedBy).First(&user)
	if user.ID > 0 {
		operatorName = user.RealName
	}

	// 创建演练日志
	s.drillRepo.CreateLog(&entity.DrillInstanceLog{
		DrillInstanceID: drill.ID,
		Action:          "start",
		OperatorID:      drill.CreatedBy,
		OperatorName:    operatorName,
		Content:         "演练已启动",
	})

	return s.engine.Start(int64(drill.ID))
}

func (s *DrillService) Pause(id uint64) error {
	drill, err := s.drillRepo.FindByID(id)
	if err != nil {
		return err
	}
	if drill.Status != "running" {
		return errors.New("只有运行中的演练才能暂停")
	}
	prevStatus := drill.Status

	if err := s.drillRepo.UpdateStatus(id, "paused"); err != nil {
		return err
	}

	// 查询操作人姓名
	var user entity.User
	operatorName := ""
	repository.DB.Model(&entity.User{}).Where("id = ?", drill.CreatedBy).First(&user)
	if user.ID > 0 {
		operatorName = user.RealName
	}

	// 创建演练日志
	s.drillRepo.CreateLog(&entity.DrillInstanceLog{
		DrillInstanceID: id,
		Action:          "pause",
		OperatorID:      drill.CreatedBy,
		OperatorName:    operatorName,
		Content:         "演练已暂停",
	})

	// WebSocket 广播
	if s.wsManager != nil {
		payload := websocket.DrillStatusPayload{
			DrillID:        uint(drill.ID),
			DrillName:      drill.Name,
			PreviousStatus: prevStatus,
			NewStatus:      "paused",
			Operator:       operatorName,
		}
		s.wsManager.SendDrillStatus(uint(drill.ID), payload)
	}

	// 创建通知（自己操作不通知自己）
	if s.notificationService != nil {
		s.notificationService.CreateNotification(&entity.Notification{
			UserID:  drill.CreatedBy,
			Type:    "drill_paused",
			Title:   "演练已暂停",
			Content: drill.Name + " 已暂停",
			DrillID: &drill.ID,
			IsRead:  false,
		}, drill.CreatedBy)
	}

	return nil
}

func (s *DrillService) Resume(id uint64) error {
	drill, err := s.drillRepo.FindByID(id)
	if err != nil {
		return err
	}
	if drill.Status != "paused" {
		return errors.New("只有已暂停的演练才能继续")
	}
	prevStatus := drill.Status

	if err := s.drillRepo.UpdateStatus(id, "running"); err != nil {
		return err
	}

	// 查询操作人姓名
	var user entity.User
	operatorName := ""
	repository.DB.Model(&entity.User{}).Where("id = ?", drill.CreatedBy).First(&user)
	if user.ID > 0 {
		operatorName = user.RealName
	}

	// 创建演练日志
	s.drillRepo.CreateLog(&entity.DrillInstanceLog{
		DrillInstanceID: id,
		Action:          "resume",
		OperatorID:      drill.CreatedBy,
		OperatorName:    operatorName,
		Content:         "演练已恢复",
	})

	// WebSocket 广播
	if s.wsManager != nil {
		payload := websocket.DrillStatusPayload{
			DrillID:        uint(drill.ID),
			DrillName:      drill.Name,
			PreviousStatus: prevStatus,
			NewStatus:      "running",
			Operator:       operatorName,
		}
		s.wsManager.SendDrillStatus(uint(drill.ID), payload)
	}

	// 创建通知
	if s.notificationService != nil {
		s.notificationService.CreateNotification(&entity.Notification{
			UserID:  drill.CreatedBy,
			Type:    "drill_resumed",
			Title:   "演练已恢复",
			Content: drill.Name + " 已恢复执行",
			DrillID: &drill.ID,
			IsRead:  false,
		}, drill.CreatedBy)
	}

	return nil
}

func (s *DrillService) Terminate(id uint64) error {
	drill, err := s.drillRepo.FindByID(id)
	if err != nil {
		return err
	}
	if drill.Status == "completed" {
		return errors.New("已完成的演练不能终止")
	}
	if drill.Status == "terminated" {
		return errors.New("演练已终止，无法重复操作")
	}
	prevStatus := drill.Status

	now := time.Now()
	repository.DB.Model(&entity.DrillInstance{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":   "terminated",
		"end_time": &now,
	})

	// 查询操作人姓名
	var user entity.User
	operatorName := ""
	repository.DB.Model(&entity.User{}).Where("id = ?", drill.CreatedBy).First(&user)
	if user.ID > 0 {
		operatorName = user.RealName
	}

	// 创建演练日志
	s.drillRepo.CreateLog(&entity.DrillInstanceLog{
		DrillInstanceID: id,
		Action:          "terminate",
		OperatorID:      drill.CreatedBy,
		OperatorName:    operatorName,
		Content:         "演练已终止",
	})

	// WebSocket 广播
	if s.wsManager != nil {
		payload := websocket.DrillStatusPayload{
			DrillID:        uint(drill.ID),
			DrillName:      drill.Name,
			PreviousStatus: prevStatus,
			NewStatus:      "terminated",
			Operator:       operatorName,
		}
		s.wsManager.SendDrillStatus(uint(drill.ID), payload)
	}

	// 创建通知
	if s.notificationService != nil {
		s.notificationService.CreateNotification(&entity.Notification{
			UserID:  drill.CreatedBy,
			Type:    "drill_terminated",
			Title:   "演练已终止",
			Content: drill.Name + " 已终止",
			DrillID: &drill.ID,
			IsRead:  false,
		}, drill.CreatedBy)
	}

	return nil
}

func (s *DrillService) GetLogs(id uint64) ([]entity.DrillInstanceLog, error) {
	return s.drillRepo.GetLogs(id, 0) // 0 → 默认 200 条
}

func (s *DrillService) Delete(id uint64) error {
	return s.drillRepo.Delete(id)
}
