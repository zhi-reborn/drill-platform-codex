package dto

import "drill-platform/internal/domain/entity"

type LoginRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=100"`
}

type LoginResponse struct {
	Token      string `json:"token"`
	UserID     uint64 `json:"user_id"`
	Username   string `json:"username"`
	RealName   string `json:"real_name"`
	Role       string `json:"role"`
	Department string `json:"department"`
}

type CreateDrillRequest struct {
	TemplateID   uint64              `json:"template_id" binding:"required"`
	Name         string              `json:"name" binding:"required,max=200"`
	PlannedStart string              `json:"planned_start"`
	Assignees    map[uint64][]uint64 `json:"assignees"`
}

type CreateDrillInstanceRequest struct {
	TemplateID uint   `json:"template_id" binding:"required"`
	Name       string `json:"name" binding:"required,max=200"`
}

type CreateTemplateRequest struct {
	Name        string              `json:"name" binding:"required,max=200"`
	Category    string              `json:"category" binding:"required,max=50"`
	Description string              `json:"description"`
	Steps       []StepTemplateRequest `json:"steps"`
}

type UpdateTemplateRequest struct {
	Name        string `json:"name" binding:"required,max=200"`
	Category    string `json:"category" binding:"required,max=50"`
	Description string `json:"description"`
}

type StepTemplateRequest struct {
	Name                string  `json:"name" binding:"required,max=200"`
	Seq                 int     `json:"seq" binding:"required"`
	StepType            string  `json:"step_type" binding:"required,oneof=serial parallel any_of condition"`
	TimeoutMinutes      int     `json:"timeout_minutes"`
	PreStepIDs          []int64 `json:"pre_step_ids"`
	GuideContent        string  `json:"guide_content"`
	IsBlocking          int8    `json:"is_blocking"`
	DefaultAssigneeRole string `json:"default_assignee_role"`
}

type UpdateDrillStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=running paused terminated"`
}

type CompleteStepRequest struct {
	StepID uint   `json:"step_id" binding:"required"`
	Result string `json:"result"`
}

type ReportIssueRequest struct {
	StepID uint   `json:"step_id" binding:"required"`
	Reason string `json:"reason" binding:"required"`
}

type AssignStepRequest struct {
	DrillID uint   `json:"drill_id" binding:"required"`
	StepID  uint   `json:"step_id" binding:"required"`
	UserIDs []uint `json:"user_ids" binding:"required,min=1"`
}

type PageQuery struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Keyword  string `form:"keyword"`
	Category string `form:"category"`
	Status   string `form:"status"`
}

func (q *PageQuery) Normalize() {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 || q.PageSize > 100 {
		q.PageSize = 20
	}
}

type PaginationRequest struct {
	Page     int    `form:"page" binding:"min=1"`
	PageSize int    `form:"page_size" binding:"min=1,max=100"`
	Keyword  string `form:"keyword"`
	Category string `form:"category"`
	Status   string `form:"status"`
}

type PaginationResponse struct {
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	Items    interface{} `json:"items"`
}

type WebSocketMessage struct {
	EventType string      `json:"event_type"`
	DrillID   uint        `json:"drill_id"`
	Payload   interface{} `json:"payload"`
	Timestamp int64       `json:"timestamp"`
}

type NotificationQuery struct {
	Page       int  `form:"page"`
	PageSize   int  `form:"page_size"`
	UnreadOnly bool `form:"unread_only"`
}

func (q *NotificationQuery) Normalize() {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 || q.PageSize > 100 {
		q.PageSize = 10
	}
}

type NotificationListResponse struct {
	Total int64                    `json:"total"`
	Items []entity.Notification `json:"items"`
}
