package main

// 分布式系统故障应急演练模板种子脚本
//
// 用法：
//   1. 确保后端服务已启动（默认监听 http://localhost:8080）
//   2. 运行： go run cmd/seed-template-distributed/main.go
//   3. 指向其他后端： BASE_URL=http://host:port/api/v1 go run cmd/seed-template-distributed/main.go
//
// 模板结构（符合规格）：
//   - 4 个阶段，每阶段 2-16 个环节（实际 3-5 个）
//   - 环节名 4-12 个字
//   - 每个环节下 1-6 个任务（实际 2 个）
//   - 每个任务下 0-6 个子任务（实际 2-3 个）
//   - 阶段、环节：串行（serial）
//   - 任务、子任务：串行/并行混合（serial / parallel）
//
// 复用方式：修改下方 generateSteps() 中的 phases 数据即可生成不同内容的模板。
// 脚本会先按名称删除同名旧模板，再重新创建，可安全重复执行。

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
	username     = "admin"
	password     = "admin123"
	defaultTime  = 120
	templateName = "分布式系统故障应急演练模板"
)

var baseURL = func() string {
	if v := os.Getenv("BASE_URL"); v != "" {
		return v
	}
	return "http://localhost:8080/api/v1"
}()

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

	// 删除同名旧模板，保证可重复执行
	deleteTemplateByName(token, templateName)

	templateID, err := createTemplate(token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "创建模板失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("模板已创建, ID=%d\n", templateID)

	steps := generateSteps()
	fmt.Printf("生成 %d 个步骤\n", len(steps))

	if err := updateSteps(token, templateID, steps); err != nil {
		fmt.Fprintf(os.Stderr, "更新步骤失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("步骤已保存")

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
		"name":        templateName,
		"category":    "灾备",
		"description": "面向分布式系统的4阶段故障应急演练模板，覆盖故障感知、应急响应、处置恢复、复盘改进，阶段与环节串行，任务/子任务串并行混合",
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
	fmt.Printf("请求体大小: %d bytes, 步骤数: %d\n", len(body), len(steps))
	resp, err := apiPut(token, url, body)
	if err != nil {
		return fmt.Errorf("HTTP 请求失败: %w", err)
	}
	fmt.Printf("API 响应: %s\n", string(resp))
	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	json.Unmarshal(resp, &result)
	if result.Code != 0 {
		return fmt.Errorf("API 错误(code=%d): %s", result.Code, result.Message)
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
	fmt.Printf("根节点数(阶段): %d\n", roots)

	if roots != 4 {
		fmt.Fprintf(os.Stderr, "错误: 期望 4 个阶段(根节点)，实际 %d 个！\n", roots)
		fmt.Fprintf(os.Stderr, "根节点列表:\n")
		for _, s := range result.Data.Steps {
			if s.ParentStepID == nil {
				fmt.Fprintf(os.Stderr, "  id=%d seq=%d name=%s\n", s.ID, s.Seq, s.Name)
			}
		}
		os.Exit(1)
	}

	// 打印树形结构
	children := map[int][]int{}
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
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return respBody, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}
	return respBody, nil
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

// generateSteps 生成 4阶段 × (3-5)环节 × 2任务 × (2-3)子任务 的完整步骤树
//
// 层级：phase(阶段) → stage(环节) → task(任务) → sub(子任务, 叶子)
// 约定：
//   - parent_step_id 传 1-based 位置索引（非 DB ID），repository 层做位置→ID 映射
//   - 步骤必须按 Seq 升序、深度优先排列（父在子前）
//   - step_type 同时表达"步骤类型"和"串/并行"：serial=串行, parallel=并行
func generateSteps() []stepTemplateReq {
	type subTaskDef struct {
		name     string
		stepType string
	}
	type taskDef struct {
		name     string
		stepType string
		subs     []subTaskDef
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
		// 阶段1：故障感知与预警（4 环节）
		{
			name: "故障感知与预警", stepType: "serial",
			stages: []stageDef{
				{name: "监控告警触发", stepType: "serial", tasks: []taskDef{
					{name: "告警信息确认", stepType: "serial", subs: []subTaskDef{
						{"查看告警详情", "serial"},
						{"确认告警级别", "serial"},
						{"关联告警分析", "parallel"},
					}},
					{name: "值班响应启动", stepType: "serial", subs: []subTaskDef{
						{"值班人员到位", "serial"},
						{"响应级别判定", "serial"},
					}},
				}},
				{name: "故障现象研判", stepType: "serial", tasks: []taskDef{
					{name: "现象信息收集", stepType: "parallel", subs: []subTaskDef{
						{"用户反馈汇总", "serial"},
						{"监控指标提取", "serial"},
						{"日志异常筛查", "parallel"},
					}},
					{name: "故障特征分析", stepType: "serial", subs: []subTaskDef{
						{"影响系统确认", "serial"},
						{"故障范围界定", "serial"},
					}},
				}},
				{name: "影响范围评估", stepType: "serial", tasks: []taskDef{
					{name: "业务影响评估", stepType: "serial", subs: []subTaskDef{
						{"受影响业务确认", "serial"},
						{"用户规模评估", "serial"},
					}},
					{name: "故障等级判定", stepType: "serial", subs: []subTaskDef{
						{"等级标准对照", "serial"},
						{"应急级别确定", "serial"},
					}},
				}},
				{name: "信息通报分发", stepType: "serial", tasks: []taskDef{
					{name: "内部通报", stepType: "parallel", subs: []subTaskDef{
						{"技术团队通知", "serial"},
						{"运维团队通知", "serial"},
						{"管理层汇报", "serial"},
					}},
					{name: "外部通报", stepType: "serial", subs: []subTaskDef{
						{"通报必要性评估", "serial"},
						{"通报内容编写", "serial"},
					}},
				}},
			},
		},
		// 阶段2：应急响应启动（3 环节）
		{
			name: "应急响应启动", stepType: "serial",
			stages: []stageDef{
				{name: "应急小组组建", stepType: "serial", tasks: []taskDef{
					{name: "指挥体系建立", stepType: "serial", subs: []subTaskDef{
						{"应急指挥确定", "serial"},
						{"核心成员召集", "parallel"},
					}},
					{name: "角色职责分配", stepType: "serial", subs: []subTaskDef{
						{"应急角色定义", "serial"},
						{"具体职责下发", "serial"},
					}},
				}},
				{name: "处置方案制定", stepType: "serial", tasks: []taskDef{
					{name: "故障根因研判", stepType: "parallel", subs: []subTaskDef{
						{"技术根因排查", "serial"},
						{"关联系统排查", "serial"},
						{"故障链路确认", "serial"},
					}},
					{name: "恢复方案设计", stepType: "serial", subs: []subTaskDef{
						{"方案选项生成", "parallel"},
						{"方案风险评估", "serial"},
						{"最终方案确定", "serial"},
					}},
				}},
				{name: "资源紧急调度", stepType: "serial", tasks: []taskDef{
					{name: "技术资源调配", stepType: "serial", subs: []subTaskDef{
						{"服务器资源申请", "serial"},
						{"网络设备调配", "parallel"},
					}},
					{name: "人员资源到位", stepType: "serial", subs: []subTaskDef{
						{"相关人员通知", "serial"},
						{"到岗情况确认", "serial"},
					}},
				}},
			},
		},
		// 阶段3：故障处置恢复（5 环节）
		{
			name: "故障处置恢复", stepType: "serial",
			stages: []stageDef{
				{name: "故障隔离阻断", stepType: "serial", tasks: []taskDef{
					{name: "故障节点隔离", stepType: "parallel", subs: []subTaskDef{
						{"异常节点摘除", "serial"},
						{"流量自动切走", "serial"},
					}},
					{name: "故障扩散阻断", stepType: "serial", subs: []subTaskDef{
						{"级联影响评估", "serial"},
						{"阻断措施执行", "serial"},
					}},
				}},
				{name: "流量切换调度", stepType: "serial", tasks: []taskDef{
					{name: "备用容量确认", stepType: "serial", subs: []subTaskDef{
						{"备节点状态检查", "parallel"},
						{"容量余量评估", "serial"},
					}},
					{name: "流量逐步切换", stepType: "serial", subs: []subTaskDef{
						{"小流量验证", "serial"},
						{"全量流量切换", "serial"},
					}},
				}},
				{name: "核心服务恢复", stepType: "serial", tasks: []taskDef{
					{name: "服务实例拉起", stepType: "parallel", subs: []subTaskDef{
						{"核心服务重启", "serial"},
						{"依赖服务确认", "serial"},
					}},
					{name: "服务健康验证", stepType: "serial", subs: []subTaskDef{
						{"健康检查通过", "serial"},
						{"核心链路验证", "serial"},
					}},
				}},
				{name: "数据一致性校验", stepType: "serial", tasks: []taskDef{
					{name: "数据状态核对", stepType: "parallel", subs: []subTaskDef{
						{"主备数据比对", "serial"},
						{"缓存数据核对", "serial"},
						{"消息队列对账", "serial"},
					}},
					{name: "数据修复处理", stepType: "serial", subs: []subTaskDef{
						{"差异数据定位", "serial"},
						{"修复脚本执行", "serial"},
					}},
				}},
				{name: "全量流量回切", stepType: "serial", tasks: []taskDef{
					{name: "回切准备", stepType: "serial", subs: []subTaskDef{
						{"回切方案确认", "serial"},
						{"监控基线准备", "serial"},
					}},
					{name: "流量分批回切", stepType: "serial", subs: []subTaskDef{
						{"首批流量回切", "serial"},
						{"剩余流量回切", "parallel"},
					}},
				}},
			},
		},
		// 阶段4：复盘总结改进（3 环节）
		{
			name: "复盘总结改进", stepType: "serial",
			stages: []stageDef{
				{name: "故障时间线复盘", stepType: "serial", tasks: []taskDef{
					{name: "操作日志收集", stepType: "serial", subs: []subTaskDef{
						{"处置日志归档", "serial"},
						{"关键节点标注", "parallel"},
					}},
					{name: "时间线梳理", stepType: "serial", subs: []subTaskDef{
						{"事件顺序整理", "serial"},
						{"关键决策复盘", "serial"},
					}},
				}},
				{name: "根因深度定位", stepType: "serial", tasks: []taskDef{
					{name: "技术根因分析", stepType: "parallel", subs: []subTaskDef{
						{"代码层根因", "serial"},
						{"架构层根因", "serial"},
						{"运维层根因", "serial"},
					}},
					{name: "流程根因分析", stepType: "serial", subs: []subTaskDef{
						{"流程缺陷识别", "serial"},
						{"改进机会梳理", "serial"},
					}},
				}},
				{name: "改进措施落地", stepType: "serial", tasks: []taskDef{
					{name: "技术改进执行", stepType: "serial", subs: []subTaskDef{
						{"改进方案制定", "parallel"},
						{"改进任务排期", "serial"},
					}},
					{name: "流程改进执行", stepType: "serial", subs: []subTaskDef{
						{"预案更新", "serial"},
						{"培训计划制定", "serial"},
					}},
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

				for _, sub := range task.subs {
					steps = append(steps, stepTemplateReq{
						Name: sub.name, Seq: seq, StepType: sub.stepType,
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
