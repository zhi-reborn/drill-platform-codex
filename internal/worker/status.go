package worker

type Status string

const (
	StatusStandby     Status = "standby"
	StatusRecovering  Status = "recovering"
	StatusLeaderReady Status = "leader-ready"
	StatusStopping    Status = "stopping"
)
