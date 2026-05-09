package websocket

import "time"

// 消息事件类型
const (
	// 系统事件
	EventPing   = "ping"   // 客户端心跳
	EventPong   = "pong"   // 服务端心跳响应
	EventError  = "error"  // 错误消息
	EventInfo   = "info"   // 信息消息

	// 演练状态事件
	EventDrillStarted    = "drill_started"     // 演练开始
	EventDrillPaused     = "drill_paused"      // 演练暂停
	EventDrillResumed    = "drill_resumed"     // 演练恢复
	EventDrillCompleted  = "drill_completed"   // 演练完成
	EventDrillTerminated = "drill_terminated"  // 演练终止

	// 步骤状态事件
	EventStepStarted   = "step_started"   // 步骤开始
	EventStepComplete  = "step_complete"  // 步骤完成
	EventStepIssue     = "step_issue"     // 步骤上报异常
	EventStepSkipped   = "step_skipped"   // 步骤跳过
	EventStepTimeout   = "step_timeout"   // 步骤超时

	// 预警事件
	EventTimeoutWarning = "timeout_warning" // 超时预警
	EventTimeoutAlert   = "timeout_alert"   // 超时告警

	// 控制事件
	EventControlPause     = "control_pause"      // 暂停指令
	EventControlResume    = "control_resume"     // 恢复指令
	EventControlTerminate = "control_terminate"  // 终止指令
	EventControlComment   = "control_comment"    // 指挥评论
)

// 通道类型
const (
	ChannelDisplay = "display" // 大屏通道
	ChannelTasks   = "tasks"   // 任务通道
	ChannelControl = "control" // 指挥通道
)

// WSMessage WebSocket 消息结构
type WSMessage struct {
	EventType string      `json:"event_type"`
	DrillID   uint        `json:"drill_id,omitempty"`
	Payload   interface{} `json:"payload,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// NewMessage 创建新的 WebSocket 消息
func NewMessage(eventType string, drillID uint, payload interface{}) WSMessage {
	return WSMessage{
		EventType: eventType,
		DrillID:   drillID,
		Payload:   payload,
		Timestamp: time.Now().Unix(),
	}
}

// NewPingMessage 创建心跳消息
func NewPingMessage() WSMessage {
	return WSMessage{
		EventType: EventPing,
		Timestamp: time.Now().Unix(),
	}
}

// NewPongMessage 创建心跳响应消息
func NewPongMessage() WSMessage {
	return WSMessage{
		EventType: EventPong,
		Timestamp: time.Now().Unix(),
	}
}

// NewErrorMessage 创建错误消息
func NewErrorMessage(message string) WSMessage {
	return WSMessage{
		EventType: EventError,
		Payload:   map[string]string{"message": message},
		Timestamp: time.Now().Unix(),
	}
}

// 心跳配置
const (
	PingPeriod    = 30 * time.Second  // 服务端发送 ping 的间隔
	PongWait      = 60 * time.Second  // 等待 pong 的超时时间
	WriteWait     = 10 * time.Second  // 写入超时时间
	MaxMessageSize = 4096             // 最大消息大小 (4KB)
)

// 步骤状态变更 Payload 结构
type StepChangePayload struct {
	DrillID       uint   `json:"drill_id"`
	StepID        uint   `json:"step_id"`
	StepName      string `json:"step_name"`
	PreviousStatus string `json:"previous_status"`
	NewStatus     string `json:"new_status"`
	Executor      string `json:"executor,omitempty"`
	Comment       string `json:"comment,omitempty"`
}

// 超时预警 Payload 结构
type TimeoutWarningPayload struct {
	DrillID      uint   `json:"drill_id"`
	StepID       uint   `json:"step_id"`
	StepName     string `json:"step_name"`
	Executor     string `json:"executor"`
	TimeoutAt    int64  `json:"timeout_at"`
	RemainingSec int    `json:"remaining_sec"`
	Level        string `json:"level"` // "warning" | "alert"
}

// 演练状态变更 Payload 结构
type DrillStatusPayload struct {
	DrillID        uint   `json:"drill_id"`
	DrillName      string `json:"drill_name"`
	PreviousStatus string `json:"previous_status"`
	NewStatus      string `json:"new_status"`
	Operator       string `json:"operator,omitempty"`
}

// 控制指令 Payload 结构
type ControlPayload struct {
	DrillID  uint   `json:"drill_id"`
	Action   string `json:"action"`   // pause, resume, terminate
	Operator string `json:"operator"`
	Comment  string `json:"comment,omitempty"`
}

// 任务分配 Payload 结构
type TaskAssignPayload struct {
	UserID      uint   `json:"user_id"`
	DrillID     uint   `json:"drill_id"`
	StepID      uint   `json:"step_id"`
	StepName    string `json:"step_name"`
	DrillName   string `json:"drill_name"`
	Deadline    int64  `json:"deadline"`
	Action      string `json:"action"` // assigned, updated, completed, expired
}
