package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
	baseURL      = "http://localhost:8080/api/v1"
	username     = "admin"
	password     = "admin123"
	defaultTime  = 120
)

type loginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type stepTemplateReq struct {
	Name           string `json:"name"`
	Seq            int    `json:"seq"`
	StepType       string `json:"step_type"`
	ParentStepID   *int   `json:"parent_step_id"`
	Phase          string `json:"phase"`
	PhaseStep      string `json:"phase_step"`
	TimeoutMinutes int    `json:"timeout_minutes"`
	Attributes     string `json:"attributes"`
}

func main() {
	token, err := login()
	if err != nil {
		fmt.Fprintf(os.Stderr, "登录失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("登录成功")

	// 删除旧模板（如果存在）
	deleteTemplateByName(token, "IT故障应急演练模板")

	// 创建模板
	templateID, err := createTemplate(token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "创建模板失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("模板已创建, ID=%d\n", templateID)

	// 生成步骤
	steps := generateSteps()
	fmt.Printf("生成 %d 个步骤\n", len(steps))

	// 打印前10步调试信息
	fmt.Println("\n--- 前10步调试 ---")
	for i, s := range steps {
		if i >= 10 {
			break
		}
		pid := "nil"
		if s.ParentStepID != nil {
			pid = fmt.Sprintf("%d", *s.ParentStepID)
		}
		fmt.Printf("  pos=%d seq=%d parent=%s name=%s [%s]\n", i+1, s.Seq, pid, s.Name, s.StepType)
	}

	// 保存步骤
	if err := updateSteps(token, templateID, steps); err != nil {
		fmt.Fprintf(os.Stderr, "更新步骤失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("\n步骤已保存")

	// 验证
	verifyTemplate(token, templateID)
}

func login() (string, error) {
	body, _ := json.Marshal(loginReq{Username: username, Password: password})
	resp, err := http.Post(baseURL+"/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var result struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Data.Token, nil
}

func deleteTemplateByName(token, name string) {
	resp, err := apiGet(token, "/templates")
	if err != nil {
		return
	}
	var result struct {
		Data []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
	}
	json.Unmarshal(resp, &result)
	for _, t := range result.Data {
		if t.Name == name {
			apiDelete(token, fmt.Sprintf("/templates/%d", t.ID))
			fmt.Printf("已删除旧模板 ID=%d\n", t.ID)
		}
	}
}

func createTemplate(token string) (int, error) {
	body, _ := json.Marshal(map[string]string{
		"name":        "IT故障应急演练模板",
		"category":    "灾备",
		"description": "完整的4阶段故障应急演练流程模板，涵盖故障发现、应急响应、故障恢复、总结改进四个阶段",
	})
	resp, err := apiPost(token, "/templates", body)
	if err != nil {
		return 0, err
	}
	var result struct {
		Data struct {
			ID int `json:"id"`
		} `json:"data"`
	}
	json.Unmarshal(resp, &result)
	if result.Data.ID == 0 {
		return 0, fmt.Errorf("创建模板返回 ID 为 0, 响应: %s", string(resp))
	}
	return result.Data.ID, nil
}

func updateSteps(token string, templateID int, steps []stepTemplateReq) error {
	body, _ := json.Marshal(map[string]interface{}{"steps": steps})
	url := fmt.Sprintf("/templates/%d/steps", templateID)
	resp, err := apiPut(token, url, body)
	if err != nil {
		return err
	}
	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	json.Unmarshal(resp, &result)
	if result.Code != 0 {
		return fmt.Errorf("API 错误: %s", result.Message)
	}
	return nil
}

func verifyTemplate(token string, templateID int) {
	resp, err := apiGet(token, fmt.Sprintf("/templates/%d", templateID))
	if err != nil {
		fmt.Fprintf(os.Stderr, "验证失败: %v\n", err)
		return
	}
	var result struct {
		Data struct {
			Name  string `json:"name"`
			Steps []struct {
				ID           int    `json:"id"`
				Name         string `json:"name"`
				Seq          int    `json:"seq"`
				StepType     string `json:"step_type"`
				ParentStepID *int   `json:"parent_step_id"`
			} `json:"steps"`
		} `json:"data"`
	}
	json.Unmarshal(resp, &result)

	roots := 0
	for _, s := range result.Data.Steps {
		if s.ParentStepID == nil {
			roots++
		}
	}
	fmt.Printf("\n--- 验证结果 ---\n")
	fmt.Printf("模板: %s\n", result.Data.Name)
	fmt.Printf("总步骤数: %d\n", len(result.Data.Steps))
	fmt.Printf("根节点数: %d\n", roots)

	if roots != 4 {
		fmt.Fprintf(os.Stderr, "错误: 期望4个阶段(根节点)，实际%d个！\n", roots)
		fmt.Fprintf(os.Stderr, "根节点列表:\n")
		for _, s := range result.Data.Steps {
			if s.ParentStepID == nil {
				fmt.Fprintf(os.Stderr, "  id=%d seq=%d name=%s\n", s.ID, s.Seq, s.Name)
			}
		}
		os.Exit(1)
	}

	// 打印树形结构
	children := map[int][]int{} // parentID -> child indices
	for i, s := range result.Data.Steps {
		if s.ParentStepID != nil {
			children[*s.ParentStepID] = append(children[*s.ParentStepID], i)
		}
	}
	for _, s := range result.Data.Steps {
		if s.ParentStepID == nil {
			printTree(result.Data.Steps, children, s.ID, 0)
		}
	}
	fmt.Printf("\n模板创建完成！ID=%d\n", templateID)
}

func printTree(steps []struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Seq          int    `json:"seq"`
	StepType     string `json:"step_type"`
	ParentStepID *int   `json:"parent_step_id"`
}, children map[int][]int, id int, indent int) {
	for _, s := range steps {
		if s.ID == id {
			fmt.Printf("%s- %s [%s]\n", repeat("  ", indent), s.Name, s.StepType)
			if idxs, ok := children[id]; ok {
				for _, idx := range idxs {
					printTree(steps, children, steps[idx].ID, indent+1)
				}
			}
			return
		}
	}
}

func repeat(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}

// HTTP helpers
func apiGet(token, path string) ([]byte, error) {
	req, _ := http.NewRequest("GET", baseURL+path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func apiPost(token, path string, body []byte) ([]byte, error) {
	req, _ := http.NewRequest("POST", baseURL+path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func apiPut(token, path string, body []byte) ([]byte, error) {
	req, _ := http.NewRequest("PUT", baseURL+path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func apiDelete(token, path string) ([]byte, error) {
	req, _ := http.NewRequest("DELETE", baseURL+path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// generateSteps 生成4阶段 × 3-4环节 × 3-4任务 × 3-4操作 的完整步骤树
func generateSteps() []stepTemplateReq {
	type taskDef struct {
		name     string
		stepType string
		ops      []string
	}
	type stageDef struct {
		name     string
		stepType string
		tasks    []taskDef
	}
	type phaseDef struct {
		name     string
		stepType string
		stages   []stageDef
	}

	phases := []phaseDef{
		{
			name: "故障发现与报告", stepType: "serial",
			stages: []stageDef{
				{name: "告警监测", stepType: "serial", tasks: []taskDef{
					{name: "告警确认", stepType: "serial", ops: []string{"查看告警详情", "确认告警级别", "通知相关人员"}},
					{name: "告警分析", stepType: "parallel", ops: []string{"分析指标趋势", "检查日志异常", "确认影响范围"}},
					{name: "告警升级", stepType: "serial", ops: []string{"编写告警报告", "提交上级审批", "通知相关团队"}},
				}},
				{name: "故障报告", stepType: "parallel", tasks: []taskDef{
					{name: "用户反馈接收", stepType: "serial", ops: []string{"接收用户反馈", "分类反馈信息", "确认故障现象"}},
					{name: "运维巡检", stepType: "serial", ops: []string{"执行巡检任务", "记录异常发现", "上报巡检结果"}},
					{name: "自动检测", stepType: "parallel", ops: []string{"触发自动检测脚本", "分析检测结果", "生成检测报告"}},
					{name: "初步评估", stepType: "serial", ops: []string{"确认受影响系统", "评估业务影响", "判定故障等级", "确认响应级别"}},
				}},
				{name: "信息通报", stepType: "serial", tasks: []taskDef{
					{name: "内部通报", stepType: "parallel", ops: []string{"通知技术团队", "通知运维团队", "通知安全团队"}},
					{name: "外部通报", stepType: "serial", ops: []string{"评估通报必要性", "编写通报内容", "发送外部通报"}},
					{name: "管理层汇报", stepType: "serial", ops: []string{"准备汇报材料", "向管理层汇报", "记录管理层指示"}},
				}},
			},
		},
		{
			name: "应急响应与评估", stepType: "parallel",
			stages: []stageDef{
				{name: "团队集结", stepType: "serial", tasks: []taskDef{
					{name: "应急团队组建", stepType: "serial", ops: []string{"确定应急指挥", "召集核心成员", "分配应急角色"}},
					{name: "通信渠道建立", stepType: "parallel", ops: []string{"建立应急通信群", "搭建临时会议室", "确认联络方式"}},
					{name: "应急资源盘点", stepType: "serial", ops: []string{"盘点技术资源", "盘点人力资源", "盘点外部支持资源"}},
				}},
				{name: "方案制定", stepType: "serial", tasks: []taskDef{
					{name: "故障定位", stepType: "parallel", ops: []string{"分析故障根因", "排查关联系统", "确认故障链路"}},
					{name: "方案设计", stepType: "serial", ops: []string{"制定恢复方案", "评估方案风险", "确定备选方案"}},
					{name: "方案评审", stepType: "serial", ops: []string{"组织方案评审", "收集评审意见", "确认最终方案"}},
					{name: "方案审批", stepType: "serial", ops: []string{"提交审批申请", "获取管理层批准", "下发执行指令"}},
				}},
				{name: "资源调配", stepType: "parallel", tasks: []taskDef{
					{name: "技术资源调配", stepType: "serial", ops: []string{"申请服务器资源", "调配网络设备", "准备恢复工具"}},
					{name: "人员调配", stepType: "serial", ops: []string{"通知相关人员到岗", "确认人员到位", "分配具体任务"}},
					{name: "外部资源协调", stepType: "serial", ops: []string{"联系供应商", "协调外部专家", "确认外部支持到位"}},
				}},
				{name: "应急处置", stepType: "serial", tasks: []taskDef{
					{name: "紧急止血", stepType: "parallel", ops: []string{"执行流量切换", "启动降级策略", "隔离故障节点"}},
					{name: "数据保护", stepType: "serial", ops: []string{"确认数据备份状态", "执行紧急数据备份", "验证数据完整性"}},
					{name: "影响控制", stepType: "serial", ops: []string{"评估当前影响范围", "执行影响控制措施", "监控影响扩散"}},
					{name: "状态同步", stepType: "serial", ops: []string{"更新处置进展", "同步各方状态", "调整处置策略"}},
				}},
			},
		},
		{
			name: "故障恢复与验证", stepType: "serial",
			stages: []stageDef{
				{name: "故障修复", stepType: "parallel", tasks: []taskDef{
					{name: "系统修复", stepType: "serial", ops: []string{"执行修复操作", "验证修复结果", "确认系统状态"}},
					{name: "数据恢复", stepType: "serial", ops: []string{"执行数据恢复", "校验数据一致性", "确认数据完整"}},
					{name: "配置恢复", stepType: "serial", ops: []string{"恢复系统配置", "验证配置正确性", "确认配置生效"}},
					{name: "服务重启", stepType: "serial", ops: []string{"按顺序重启服务", "验证服务状态", "确认服务可用"}},
				}},
				{name: "功能验证", stepType: "serial", tasks: []taskDef{
					{name: "基础功能验证", stepType: "parallel", ops: []string{"验证核心功能", "验证辅助功能", "验证安全功能"}},
					{name: "性能验证", stepType: "serial", ops: []string{"执行性能测试", "对比性能基线", "确认性能达标"}},
					{name: "业务验证", stepType: "serial", ops: []string{"执行业务测试", "确认业务流程", "验证数据准确"}},
				}},
				{name: "全量恢复", stepType: "serial", tasks: []taskDef{
					{name: "流量恢复", stepType: "serial", ops: []string{"逐步恢复流量", "监控系统负载", "确认流量正常"}},
					{name: "降级解除", stepType: "serial", ops: []string{"解除降级策略", "恢复完整功能", "确认降级已解除"}},
					{name: "监控确认", stepType: "parallel", ops: []string{"确认监控指标正常", "确认告警已消除", "确认日志无异常"}},
					{name: "恢复通报", stepType: "serial", ops: []string{"通知内部恢复状态", "通知外部恢复状态", "通知管理层恢复完成"}},
				}},
			},
		},
		{
			name: "总结与改进", stepType: "serial",
			stages: []stageDef{
				{name: "复盘分析", stepType: "serial", tasks: []taskDef{
					{name: "时间线梳理", stepType: "serial", ops: []string{"收集操作日志", "梳理故障时间线", "标注关键节点"}},
					{name: "根因分析", stepType: "parallel", ops: []string{"技术根因分析", "流程根因分析", "管理根因分析"}},
					{name: "处置评估", stepType: "serial", ops: []string{"评估响应时效", "评估处置效果", "评估协作效率"}},
				}},
				{name: "改进措施", stepType: "parallel", tasks: []taskDef{
					{name: "技术改进", stepType: "serial", ops: []string{"制定技术改进方案", "评估改进优先级", "排期改进任务"}},
					{name: "流程改进", stepType: "serial", ops: []string{"梳理流程缺陷", "制定流程优化方案", "更新应急预案"}},
					{name: "监控改进", stepType: "serial", ops: []string{"完善监控指标", "优化告警规则", "增加检测手段"}},
					{name: "培训改进", stepType: "serial", ops: []string{"总结培训需求", "制定培训计划", "安排演练频次"}},
				}},
				{name: "知识沉淀", stepType: "serial", tasks: []taskDef{
					{name: "文档归档", stepType: "parallel", ops: []string{"编写故障报告", "归档处置记录", "更新知识库"}},
					{name: "经验分享", stepType: "serial", ops: []string{"组织经验分享会", "编写案例材料", "发布经验通报"}},
					{name: "预案更新", stepType: "serial", ops: []string{"更新应急预案", "更新联系人清单", "确认资源清单"}},
				}},
			},
		},
	}

	var steps []stepTemplateReq
	seq := 1
	pos := 1

	for _, phase := range phases {
		phasePos := pos
		steps = append(steps, stepTemplateReq{
			Name: phase.name, Seq: seq, StepType: phase.stepType,
			ParentStepID: nil, Phase: phase.name, PhaseStep: phase.name,
			TimeoutMinutes: defaultTime, Attributes: "{}",
		})
		seq++
		pos++

		for _, stage := range phase.stages {
			stagePos := pos
			steps = append(steps, stepTemplateReq{
				Name: stage.name, Seq: seq, StepType: stage.stepType,
				ParentStepID: intPtr(phasePos), Phase: phase.name, PhaseStep: stage.name,
				TimeoutMinutes: defaultTime, Attributes: "{}",
			})
			seq++
			pos++

			for _, task := range stage.tasks {
				taskPos := pos
				steps = append(steps, stepTemplateReq{
					Name: task.name, Seq: seq, StepType: task.stepType,
					ParentStepID: intPtr(stagePos), Phase: phase.name, PhaseStep: stage.name,
					TimeoutMinutes: defaultTime, Attributes: "{}",
				})
				seq++
				pos++

				for oi, op := range task.ops {
					opType := "serial"
					if oi%3 == 2 {
						opType = "parallel"
					}
					steps = append(steps, stepTemplateReq{
						Name: op, Seq: seq, StepType: opType,
						ParentStepID: intPtr(taskPos), Phase: phase.name, PhaseStep: stage.name,
						TimeoutMinutes: defaultTime, Attributes: "{}",
					})
					seq++
					pos++
				}
			}
		}
	}

	return steps
}

func intPtr(v int) *int { return &v }
