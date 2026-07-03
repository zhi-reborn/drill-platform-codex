package main

// 大型系统故障应急演练模板种子脚本（环节名统一 12 字，每阶段 10-16 环节）
//
// 用法：
//   1. 确保后端服务已启动
//   2. 运行： go run cmd/seed-template-detailed/main.go
//   3. 指向其他后端： BASE_URL=http://host:port/api/v1 go run cmd/seed-template-detailed/main.go
//
// 模板结构（符合规格）：
//   - 4 个阶段，每阶段 10-16 个环节（共 49 个环节）
//   - 每个环节名统一 12 个字（rune 计数，启动时内置校验，不满足则报错退出）
//   - 每个环节下 2-3 个任务
//   - 每个任务下 2-3 个子任务
//   - 阶段、环节：串行（serial）
//   - 任务、子任务：串行/并行混合（serial / parallel）
//
// 复用方式：修改下方 generateSteps() 中的 phases 数据即可。脚本启动时会校验
// 所有环节名均为 12 字（rune 计数，中文按字计），便于维护时及时发现长度错误。

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
	templateName = "大型系统故障应急演练模板"
	stageNameLen = 12 // 环节名统一字数（rune 计数）
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

	deleteTemplateByName(token, templateName)

	templateID, err := createTemplate(token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "创建模板失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("模板已创建, ID=%d\n", templateID)

	steps, err := generateSteps()
	if err != nil {
		fmt.Fprintf(os.Stderr, "生成步骤失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("生成 %d 个步骤（环节名已校验均为 %d 字）\n", len(steps), stageNameLen)

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
		"description": "面向大型系统的4阶段故障应急演练模板，环节名统一12字，每阶段10-16个环节，覆盖故障感知、应急响应、处置恢复、复盘改进，阶段与环节串行，任务/子任务串并行混合",
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
	stageCount := 0
	taskCount := 0
	subCount := 0
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
		os.Exit(1)
	}

	// 统计各层级数量
	children := map[int][]int{}
	for i, s := range result.Data.Steps {
		if s.ParentStepID != nil {
			children[*s.ParentStepID] = append(children[*s.ParentStepID], i)
		}
	}
	for _, s := range result.Data.Steps {
		if s.ParentStepID == nil {
			if idxs, ok := children[s.ID]; ok {
				stageCount += len(idxs)
				for _, si := range idxs {
					if tidxs, ok2 := children[result.Data.Steps[si].ID]; ok2 {
						taskCount += len(tidxs)
						for _, ti := range tidxs {
							if sidxs, ok3 := children[result.Data.Steps[ti].ID]; ok3 {
								subCount += len(sidxs)
							}
						}
					}
				}
			}
		}
	}
	fmt.Printf("环节数: %d\n", stageCount)
	fmt.Printf("任务数: %d\n", taskCount)
	fmt.Printf("子任务数: %d\n", subCount)

	// 打印树形结构（仅阶段和环节层，避免输出过长）
	for _, s := range result.Data.Steps {
		if s.ParentStepID == nil {
			fmt.Printf("■ %s [%s]\n", s.Name, s.StepType)
			if idxs, ok := children[s.ID]; ok {
				for _, si := range idxs {
					st := result.Data.Steps[si]
					taskN := 0
					if tidxs, ok2 := children[st.ID]; ok2 {
						taskN = len(tidxs)
					}
					fmt.Printf("  ├─ %s [%s] (%d任务)\n", st.Name, st.StepType, taskN)
				}
			}
		}
	}
	fmt.Printf("\n模板创建完成！ID=%d\n", templateID)
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

// generateSteps 生成 4阶段 × (10-16)环节 × 2-3任务 × 2-3子任务 的完整步骤树
//
// 层级：phase(阶段) → stage(环节) → task(任务) → sub(子任务, 叶子)
// 约定：
//   - parent_step_id 传 1-based 位置索引（非 DB ID），repository 层做位置→ID 映射
//   - 步骤必须按 Seq 升序、深度优先排列（父在子前）
//   - step_type 同时表达"步骤类型"和"串/并行"：serial=串行, parallel=并行
//   - 所有环节(stage)名必须为 12 字（rune 计数），函数开头校验
func generateSteps() ([]stepTemplateReq, error) {
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
		// ═══════════════════════════════════════════════════════════════
		// 阶段1：故障感知与预警（12 环节）
		// ═══════════════════════════════════════════════════════════════
		{
			name: "故障感知与预警", stepType: "serial",
			stages: []stageDef{
				{name: "监控告警的实时触发与确认", stepType: "serial", tasks: []taskDef{
					{name: "告警接收", stepType: "serial", subs: []subTaskDef{
						{"告警信息解析", "serial"},
						{"告警来源确认", "parallel"},
						{"告警级别初判", "serial"},
					}},
					{name: "告警确认", stepType: "serial", subs: []subTaskDef{
						{"值班人员响应", "serial"},
						{"告警有效性验证", "serial"},
					}},
				}},
				{name: "告警信息的分级与分类处理", stepType: "serial", tasks: []taskDef{
					{name: "信息分级", stepType: "parallel", subs: []subTaskDef{
						{"紧急程度判定", "serial"},
						{"影响范围分级", "serial"},
					}},
					{name: "分类路由", stepType: "serial", subs: []subTaskDef{
						{"告警类型归类", "serial"},
						{"处置团队路由", "parallel"},
						{"升级条件评估", "serial"},
					}},
				}},
				{name: "故障现象的初步研判与定位", stepType: "serial", tasks: []taskDef{
					{name: "现象收集", stepType: "parallel", subs: []subTaskDef{
						{"用户反馈汇总", "serial"},
						{"监控异常提取", "serial"},
						{"日志异常筛查", "parallel"},
					}},
					{name: "现象研判", stepType: "serial", subs: []subTaskDef{
						{"故障特征归纳", "serial"},
						{"影响系统确认", "serial"},
					}},
				}},
				{name: "监控指标的异常趋势与研判", stepType: "serial", tasks: []taskDef{
					{name: "指标提取", stepType: "parallel", subs: []subTaskDef{
						{"核心指标拉取", "serial"},
						{"关联指标收集", "serial"},
					}},
					{name: "趋势研判", stepType: "serial", subs: []subTaskDef{
						{"指标趋势对比", "serial"},
						{"异常模式识别", "parallel"},
						{"阈值突破确认", "serial"},
					}},
				}},
				{name: "业务影响的快速评估与判定", stepType: "serial", tasks: []taskDef{
					{name: "影响评估", stepType: "serial", subs: []subTaskDef{
						{"受影响业务确认", "serial"},
						{"用户规模评估", "parallel"},
					}},
					{name: "等级判定", stepType: "serial", subs: []subTaskDef{
						{"损失程度评估", "serial"},
						{"等级标准对照", "serial"},
						{"响应级别确定", "serial"},
					}},
				}},
				{name: "故障等级的确认与升级处理", stepType: "serial", tasks: []taskDef{
					{name: "等级确认", stepType: "serial", subs: []subTaskDef{
						{"多方信息汇总", "parallel"},
						{"等级最终确认", "serial"},
					}},
					{name: "升级处理", stepType: "serial", subs: []subTaskDef{
						{"升级条件触发", "serial"},
						{"上级通知发送", "serial"},
						{"升级记录归档", "serial"},
					}},
				}},
				{name: "用户反馈的收集与分类汇总", stepType: "serial", tasks: []taskDef{
					{name: "反馈收集", stepType: "parallel", subs: []subTaskDef{
						{"客服渠道收集", "serial"},
						{"线上反馈收集", "serial"},
						{"社交媒体监测", "parallel"},
					}},
					{name: "分类汇总", stepType: "serial", subs: []subTaskDef{
						{"反馈类型归类", "serial"},
						{"影响范围统计", "serial"},
					}},
				}},
				{name: "日志异常的筛查与关联分析", stepType: "serial", tasks: []taskDef{
					{name: "日志筛查", stepType: "parallel", subs: []subTaskDef{
						{"应用日志筛查", "serial"},
						{"系统日志筛查", "serial"},
						{"访问日志筛查", "parallel"},
					}},
					{name: "关联分析", stepType: "serial", subs: []subTaskDef{
						{"时间线关联", "serial"},
						{"因果链推断", "serial"},
					}},
				}},
				{name: "故障范围的边界界定与确认", stepType: "serial", tasks: []taskDef{
					{name: "边界界定", stepType: "serial", subs: []subTaskDef{
						{"受影响模块确认", "parallel"},
						{"未受影响模块确认", "serial"},
					}},
					{name: "范围确认", stepType: "serial", subs: []subTaskDef{
						{"边界验证测试", "serial"},
						{"范围文档记录", "serial"},
					}},
				}},
				{name: "内部团队的即时通报与分发", stepType: "serial", tasks: []taskDef{
					{name: "通报编发", stepType: "serial", subs: []subTaskDef{
						{"通报内容编写", "serial"},
						{"通报级别确认", "serial"},
					}},
					{name: "通报分发", stepType: "parallel", subs: []subTaskDef{
						{"技术团队通知", "serial"},
						{"运维团队通知", "serial"},
						{"管理层通知", "serial"},
					}},
				}},
				{name: "管理汇报与状态同步的处理", stepType: "serial", tasks: []taskDef{
					{name: "汇报准备", stepType: "serial", subs: []subTaskDef{
						{"状态信息汇总", "parallel"},
						{"汇报材料编写", "serial"},
					}},
					{name: "汇报同步", stepType: "serial", subs: []subTaskDef{
						{"管理层汇报", "serial"},
						{"状态同步更新", "serial"},
					}},
				}},
				{name: "外部通报与相关方评估处理", stepType: "serial", tasks: []taskDef{
					{name: "评估通报", stepType: "serial", subs: []subTaskDef{
						{"通报必要性评估", "serial"},
						{"相关方影响评估", "parallel"},
					}},
					{name: "通报执行", stepType: "serial", subs: []subTaskDef{
						{"通报内容审定", "serial"},
						{"外部通报发送", "serial"},
						{"通报反馈跟踪", "serial"},
					}},
				}},
			},
		},
		// ═══════════════════════════════════════════════════════════════
		// 阶段2：应急响应启动（10 环节）
		// ═══════════════════════════════════════════════════════════════
		{
			name: "应急响应启动", stepType: "serial",
			stages: []stageDef{
				{name: "应急小组的快速组建与到位", stepType: "serial", tasks: []taskDef{
					{name: "小组组建", stepType: "serial", subs: []subTaskDef{
						{"应急指挥确定", "serial"},
						{"核心成员召集", "parallel"},
					}},
					{name: "到位确认", stepType: "serial", subs: []subTaskDef{
						{"人员到位确认", "serial"},
						{"职责分工明确", "serial"},
						{"通讯方式确认", "serial"},
					}},
				}},
				{name: "指挥体系的建立与运行保障", stepType: "serial", tasks: []taskDef{
					{name: "体系建立", stepType: "serial", subs: []subTaskDef{
						{"指挥层级确立", "serial"},
						{"决策权限明确", "serial"},
					}},
					{name: "运行保障", stepType: "parallel", subs: []subTaskDef{
						{"指挥场所准备", "serial"},
						{"通讯保障建立", "serial"},
						{"信息流转机制", "serial"},
					}},
				}},
				{name: "角色职责的分配与明确确认", stepType: "serial", tasks: []taskDef{
					{name: "角色分配", stepType: "serial", subs: []subTaskDef{
						{"应急角色定义", "serial"},
						{"人员角色匹配", "parallel"},
					}},
					{name: "职责确认", stepType: "serial", subs: []subTaskDef{
						{"职责清单下发", "serial"},
						{"职责确认签字", "serial"},
					}},
				}},
				{name: "通信渠道的建立与测试验证", stepType: "serial", tasks: []taskDef{
					{name: "渠道建立", stepType: "parallel", subs: []subTaskDef{
						{"应急通信群建立", "serial"},
						{"语音会议搭建", "serial"},
						{"视频会议搭建", "parallel"},
					}},
					{name: "测试验证", stepType: "serial", subs: []subTaskDef{
						{"通信连通测试", "serial"},
						{"备用渠道验证", "serial"},
					}},
				}},
				{name: "故障根因的深度研判与分析", stepType: "serial", tasks: []taskDef{
					{name: "根因研判", stepType: "parallel", subs: []subTaskDef{
						{"技术根因排查", "serial"},
						{"关联系统排查", "serial"},
						{"故障链路确认", "serial"},
					}},
					{name: "深度分析", stepType: "serial", subs: []subTaskDef{
						{"根因假设验证", "serial"},
						{"根因确认报告", "serial"},
					}},
				}},
				{name: "恢复方案的设计与评估确认", stepType: "serial", tasks: []taskDef{
					{name: "方案设计", stepType: "serial", subs: []subTaskDef{
						{"恢复方案制定", "parallel"},
						{"备选方案准备", "serial"},
					}},
					{name: "评估确认", stepType: "serial", subs: []subTaskDef{
						{"方案风险评估", "serial"},
						{"方案可行性确认", "serial"},
						{"最终方案确定", "serial"},
					}},
				}},
				{name: "方案评审与审批确认的处理", stepType: "serial", tasks: []taskDef{
					{name: "方案评审", stepType: "serial", subs: []subTaskDef{
						{"评审会议组织", "serial"},
						{"评审意见收集", "parallel"},
						{"方案修订完善", "serial"},
					}},
					{name: "审批处理", stepType: "serial", subs: []subTaskDef{
						{"审批申请提交", "serial"},
						{"管理层审批获取", "serial"},
						{"执行指令下发", "serial"},
					}},
				}},
				{name: "技术资源的紧急调配与到位", stepType: "serial", tasks: []taskDef{
					{name: "资源调配", stepType: "parallel", subs: []subTaskDef{
						{"服务器资源申请", "serial"},
						{"网络设备调配", "serial"},
						{"恢复工具准备", "serial"},
					}},
					{name: "到位确认", stepType: "serial", subs: []subTaskDef{
						{"资源到位验证", "serial"},
						{"资源可用测试", "serial"},
					}},
				}},
				{name: "人员资源的协调与到岗确认", stepType: "serial", tasks: []taskDef{
					{name: "人员协调", stepType: "serial", subs: []subTaskDef{
						{"到岗人员通知", "parallel"},
						{"替班人员安排", "serial"},
					}},
					{name: "到岗确认", stepType: "serial", subs: []subTaskDef{
						{"到岗情况统计", "serial"},
						{"人员就位确认", "serial"},
					}},
				}},
				{name: "外部支持的协调与确认到位", stepType: "serial", tasks: []taskDef{
					{name: "外部协调", stepType: "serial", subs: []subTaskDef{
						{"供应商联系", "serial"},
						{"外部专家协调", "parallel"},
					}},
					{name: "确认到位", stepType: "serial", subs: []subTaskDef{
						{"外部支持确认", "serial"},
						{"协作机制建立", "serial"},
					}},
				}},
			},
		},
		// ═══════════════════════════════════════════════════════════════
		// 阶段3：故障处置恢复（16 环节）
		// ═══════════════════════════════════════════════════════════════
		{
			name: "故障处置恢复", stepType: "serial",
			stages: []stageDef{
				{name: "故障节点的隔离与摘除处理", stepType: "serial", tasks: []taskDef{
					{name: "节点隔离", stepType: "parallel", subs: []subTaskDef{
						{"异常节点识别", "serial"},
						{"隔离策略制定", "serial"},
						{"隔离操作执行", "serial"},
					}},
					{name: "摘除处理", stepType: "serial", subs: []subTaskDef{
						{"流量摘除执行", "serial"},
						{"节点标记下线", "serial"},
					}},
				}},
				{name: "故障扩散的阻断与控制确认", stepType: "serial", tasks: []taskDef{
					{name: "扩散阻断", stepType: "serial", subs: []subTaskDef{
						{"扩散路径分析", "parallel"},
						{"阻断措施执行", "serial"},
					}},
					{name: "控制确认", stepType: "serial", subs: []subTaskDef{
						{"控制效果验证", "serial"},
						{"扩散趋势确认", "serial"},
					}},
				}},
				{name: "备用容量的确认与评估分析", stepType: "serial", tasks: []taskDef{
					{name: "容量确认", stepType: "parallel", subs: []subTaskDef{
						{"备节点状态检查", "serial"},
						{"资源余量盘点", "serial"},
					}},
					{name: "评估分析", stepType: "serial", subs: []subTaskDef{
						{"容量需求评估", "serial"},
						{"承载能力分析", "serial"},
					}},
				}},
				{name: "流量的逐步切换与验证确认", stepType: "serial", tasks: []taskDef{
					{name: "流量切换", stepType: "serial", subs: []subTaskDef{
						{"小流量验证", "serial"},
						{"中流量切换", "serial"},
						{"全量流量切换", "parallel"},
					}},
					{name: "验证确认", stepType: "serial", subs: []subTaskDef{
						{"切换效果验证", "serial"},
						{"系统稳定性确认", "serial"},
					}},
				}},
				{name: "核心服务的重启与拉起验证", stepType: "serial", tasks: []taskDef{
					{name: "服务重启", stepType: "parallel", subs: []subTaskDef{
						{"核心服务拉起", "serial"},
						{"依赖服务确认", "serial"},
					}},
					{name: "拉起验证", stepType: "serial", subs: []subTaskDef{
						{"服务状态检查", "serial"},
						{"健康端点验证", "serial"},
					}},
				}},
				{name: "依赖服务的确认与恢复处理", stepType: "serial", tasks: []taskDef{
					{name: "依赖确认", stepType: "serial", subs: []subTaskDef{
						{"依赖关系梳理", "parallel"},
						{"依赖状态检查", "serial"},
					}},
					{name: "恢复处理", stepType: "serial", subs: []subTaskDef{
						{"依赖服务恢复", "serial"},
						{"恢复顺序确认", "serial"},
					}},
				}},
				{name: "健康检查的执行与通过确认", stepType: "serial", tasks: []taskDef{
					{name: "检查执行", stepType: "parallel", subs: []subTaskDef{
						{"存活检查执行", "serial"},
						{"就绪检查执行", "serial"},
						{"启动检查执行", "serial"},
					}},
					{name: "通过确认", stepType: "serial", subs: []subTaskDef{
						{"检查结果汇总", "serial"},
						{"健康状态确认", "serial"},
					}},
				}},
				{name: "核心链路的验证与确认处理", stepType: "serial", tasks: []taskDef{
					{name: "链路验证", stepType: "serial", subs: []subTaskDef{
						{"核心链路测试", "parallel"},
						{"关键路径验证", "serial"},
					}},
					{name: "确认处理", stepType: "serial", subs: []subTaskDef{
						{"验证结果分析", "serial"},
						{"链路状态确认", "serial"},
					}},
				}},
				{name: "数据一致性的校验核对处理", stepType: "serial", tasks: []taskDef{
					{name: "一致性校验", stepType: "parallel", subs: []subTaskDef{
						{"主备数据比对", "serial"},
						{"缓存数据核对", "serial"},
						{"消息队列对账", "serial"},
					}},
					{name: "核对处理", stepType: "serial", subs: []subTaskDef{
						{"差异记录归档", "serial"},
						{"核对结果确认", "serial"},
					}},
				}},
				{name: "差异数据的定位与修复处理", stepType: "serial", tasks: []taskDef{
					{name: "差异定位", stepType: "serial", subs: []subTaskDef{
						{"差异数据扫描", "parallel"},
						{"差异原因分析", "serial"},
					}},
					{name: "修复处理", stepType: "serial", subs: []subTaskDef{
						{"修复脚本执行", "serial"},
						{"修复结果验证", "serial"},
					}},
				}},
				{name: "缓存数据的清理与重建处理", stepType: "serial", tasks: []taskDef{
					{name: "缓存清理", stepType: "serial", subs: []subTaskDef{
						{"脏数据识别", "parallel"},
						{"缓存清理执行", "serial"},
					}},
					{name: "重建处理", stepType: "serial", subs: []subTaskDef{
						{"缓存预热执行", "serial"},
						{"重建结果验证", "serial"},
					}},
				}},
				{name: "消息队列的对账与处理确认", stepType: "serial", tasks: []taskDef{
					{name: "队列对账", stepType: "parallel", subs: []subTaskDef{
						{"积压消息确认", "serial"},
						{"消息顺序核对", "serial"},
					}},
					{name: "处理确认", stepType: "serial", subs: []subTaskDef{
						{"积压消息处理", "serial"},
						{"队列状态确认", "serial"},
					}},
				}},
				{name: "全量流量的分批回切与处理", stepType: "serial", tasks: []taskDef{
					{name: "分批回切", stepType: "serial", subs: []subTaskDef{
						{"首批流量回切", "serial"},
						{"中间批次回切", "serial"},
						{"末批流量回切", "parallel"},
					}},
					{name: "回切处理", stepType: "serial", subs: []subTaskDef{
						{"回切监控确认", "serial"},
						{"异常回滚准备", "serial"},
					}},
				}},
				{name: "回切过程的监控与确认处理", stepType: "serial", tasks: []taskDef{
					{name: "过程监控", stepType: "parallel", subs: []subTaskDef{
						{"流量监控", "serial"},
						{"性能监控", "serial"},
						{"错误率监控", "serial"},
					}},
					{name: "确认处理", stepType: "serial", subs: []subTaskDef{
						{"监控数据汇总", "serial"},
						{"稳定状态确认", "serial"},
					}},
				}},
				{name: "降级策略的逐步解除与确认", stepType: "serial", tasks: []taskDef{
					{name: "逐步解除", stepType: "serial", subs: []subTaskDef{
						{"降级清单梳理", "serial"},
						{"降级逐项解除", "parallel"},
					}},
					{name: "解除确认", stepType: "serial", subs: []subTaskDef{
						{"功能完整性验证", "serial"},
						{"解除状态确认", "serial"},
					}},
				}},
				{name: "恢复状态的通报与确认处理", stepType: "serial", tasks: []taskDef{
					{name: "状态通报", stepType: "serial", subs: []subTaskDef{
						{"恢复报告编写", "serial"},
						{"通报内容审定", "parallel"},
					}},
					{name: "确认处理", stepType: "serial", subs: []subTaskDef{
						{"内部通报发送", "serial"},
						{"外部通报发送", "serial"},
						{"恢复状态确认", "serial"},
					}},
				}},
			},
		},
		// ═══════════════════════════════════════════════════════════════
		// 阶段4：复盘总结改进（11 环节）
		// ═══════════════════════════════════════════════════════════════
		{
			name: "复盘总结改进", stepType: "serial",
			stages: []stageDef{
				{name: "处置日志的归档与整理汇总", stepType: "serial", tasks: []taskDef{
					{name: "日志归档", stepType: "parallel", subs: []subTaskDef{
						{"操作日志归档", "serial"},
						{"通信记录归档", "serial"},
						{"决策记录归档", "serial"},
					}},
					{name: "整理汇总", stepType: "serial", subs: []subTaskDef{
						{"日志分类整理", "serial"},
						{"汇总报告生成", "serial"},
					}},
				}},
				{name: "关键节点的标注与回顾分析", stepType: "serial", tasks: []taskDef{
					{name: "节点标注", stepType: "serial", subs: []subTaskDef{
						{"关键节点识别", "parallel"},
						{"节点标注记录", "serial"},
					}},
					{name: "回顾分析", stepType: "serial", subs: []subTaskDef{
						{"节点回顾梳理", "serial"},
						{"决策时机分析", "serial"},
					}},
				}},
				{name: "故障时间线的梳理分析总结", stepType: "serial", tasks: []taskDef{
					{name: "时间线梳理", stepType: "serial", subs: []subTaskDef{
						{"事件顺序整理", "serial"},
						{"时间节点标注", "parallel"},
					}},
					{name: "分析总结", stepType: "serial", subs: []subTaskDef{
						{"关键节点分析", "serial"},
						{"时间线总结报告", "serial"},
					}},
				}},
				{name: "关键决策的复盘与评估分析", stepType: "serial", tasks: []taskDef{
					{name: "决策复盘", stepType: "serial", subs: []subTaskDef{
						{"决策点梳理", "serial"},
						{"决策依据复盘", "parallel"},
					}},
					{name: "评估分析", stepType: "serial", subs: []subTaskDef{
						{"决策效果评估", "serial"},
						{"改进机会分析", "serial"},
					}},
				}},
				{name: "技术根因的深度定位与分析", stepType: "serial", tasks: []taskDef{
					{name: "深度定位", stepType: "parallel", subs: []subTaskDef{
						{"代码层根因定位", "serial"},
						{"配置层根因定位", "serial"},
						{"基础设施根因", "serial"},
					}},
					{name: "定位分析", stepType: "serial", subs: []subTaskDef{
						{"根因链路分析", "serial"},
						{"根因确认报告", "serial"},
					}},
				}},
				{name: "架构层面的根因分析与处理", stepType: "serial", tasks: []taskDef{
					{name: "架构分析", stepType: "serial", subs: []subTaskDef{
						{"架构缺陷识别", "parallel"},
						{"单点故障分析", "serial"},
					}},
					{name: "分析处理", stepType: "serial", subs: []subTaskDef{
						{"架构改进建议", "serial"},
						{"处理方案制定", "serial"},
					}},
				}},
				{name: "运维层面的根因分析与处理", stepType: "serial", tasks: []taskDef{
					{name: "运维分析", stepType: "serial", subs: []subTaskDef{
						{"运维流程分析", "parallel"},
						{"运维工具分析", "serial"},
					}},
					{name: "分析处理", stepType: "serial", subs: []subTaskDef{
						{"运维改进建议", "serial"},
						{"处理方案制定", "serial"},
					}},
				}},
				{name: "流程缺陷的识别与梳理分析", stepType: "serial", tasks: []taskDef{
					{name: "缺陷识别", stepType: "parallel", subs: []subTaskDef{
						{"响应流程缺陷", "serial"},
						{"通报流程缺陷", "serial"},
						{"决策流程缺陷", "serial"},
					}},
					{name: "梳理分析", stepType: "serial", subs: []subTaskDef{
						{"缺陷影响评估", "serial"},
						{"改进方向梳理", "serial"},
					}},
				}},
				{name: "改进措施的制定与排期确认", stepType: "serial", tasks: []taskDef{
					{name: "措施制定", stepType: "serial", subs: []subTaskDef{
						{"技术改进制定", "parallel"},
						{"流程改进制定", "serial"},
						{"管理改进制定", "serial"},
					}},
					{name: "排期确认", stepType: "serial", subs: []subTaskDef{
						{"优先级评估", "serial"},
						{"改进排期确认", "serial"},
					}},
				}},
				{name: "应急预案的更新与完善处理", stepType: "serial", tasks: []taskDef{
					{name: "预案更新", stepType: "serial", subs: []subTaskDef{
						{"预案内容更新", "parallel"},
						{"联系人清单更新", "serial"},
					}},
					{name: "完善处理", stepType: "serial", subs: []subTaskDef{
						{"预案评审验证", "serial"},
						{"预案发布归档", "serial"},
					}},
				}},
				{name: "培训计划的制定与落实跟进", stepType: "serial", tasks: []taskDef{
					{name: "计划制定", stepType: "serial", subs: []subTaskDef{
						{"培训需求梳理", "parallel"},
						{"培训计划编写", "serial"},
					}},
					{name: "落实跟进", stepType: "serial", subs: []subTaskDef{
						{"培训任务分配", "serial"},
						{"落实进度跟踪", "serial"},
					}},
				}},
			},
		},
	}

	// 校验所有环节名均为 stageNameLen 字（rune 计数，中文按字计）
	for pi, phase := range phases {
		for si, stage := range phase.stages {
			if n := len([]rune(stage.name)); n != stageNameLen {
				return nil, fmt.Errorf("阶段%d %q 的环节#%d %q 长度为 %d 字，要求 %d 字",
					pi+1, phase.name, si+1, stage.name, n, stageNameLen)
			}
		}
	}

	// 统计环节数
	totalStages := 0
	for _, phase := range phases {
		totalStages += len(phase.stages)
	}
	fmt.Printf("校验通过：%d 个阶段，共 %d 个环节，环节名均为 %d 字\n",
		len(phases), totalStages, stageNameLen)

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

	return steps, nil
}

func intPtr(v int) *int { return &v }
