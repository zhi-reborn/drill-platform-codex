package websocket

import "errors"

var (
	ErrClientNotFound   = errors.New("客户端不存在")
	ErrSendChannelFull  = errors.New("发送队列已满")
	ErrInvalidDrillID   = errors.New("无效的演练ID")
	ErrChannelNotExists = errors.New("通道不存在")
)
