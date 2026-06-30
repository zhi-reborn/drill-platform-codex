package service

import "drill-platform/internal/domain/entity"

// terminalStepStatuses 终态步骤状态：已完成 / 已跳过 / 已超时 / 异常。
// 与前端 ScreenView.vue 的 progressPercent 计算保持一致，作为全局完成率的唯一口径。
var terminalStepStatuses = map[string]bool{
	"completed": true,
	"skipped":   true,
	"timeout":   true,
	"issue":     true,
}

// ComputeProgressPct 依据步骤实例实时计算演练完成率（0-100）。
// 口径：仅统计叶子步骤（无子步骤的步骤），终态包含 completed/skipped/timeout/issue。
// 与前端 ScreenView.vue 的 leafSteps + progressPercent 逻辑完全对应，
// 保证大屏进度环、监控页进度条、列表页进度条等同源同值。
func ComputeProgressPct(steps []entity.StepInstance) int {
	if len(steps) == 0 {
		return 0
	}

	// 收集所有作为父节点的步骤 ID
	parentIDs := make(map[uint64]bool)
	for i := range steps {
		if steps[i].ParentStepID != nil && *steps[i].ParentStepID != 0 {
			parentIDs[*steps[i].ParentStepID] = true
		}
	}

	total, completed := 0, 0
	for i := range steps {
		// 跳过父节点，只统计叶子步骤
		if parentIDs[steps[i].ID] {
			continue
		}
		total++
		if terminalStepStatuses[steps[i].Status] {
			completed++
		}
	}

	// 无层级结构（无叶子）时回退为全量步骤统计
	if total == 0 {
		total = len(steps)
		for i := range steps {
			if terminalStepStatuses[steps[i].Status] {
				completed++
			}
		}
	}

	if total == 0 {
		return 0
	}
	// 与前端 Math.round((completed/total)*100) 对齐，避免整数除法导致的 1% 偏差
	return (completed*100 + total/2) / total
}
