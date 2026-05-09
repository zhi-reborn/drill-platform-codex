package flowengine

import "errors"

var (
	ErrInstanceNotFound   = errors.New("flow instance not found")
	ErrStepNotFound       = errors.New("step not found")
	ErrInvalidStatus      = errors.New("invalid status transition")
	ErrInstanceNotRunning = errors.New("flow instance is not running")
	ErrStepNotActive      = errors.New("step is not in running status")
	ErrPreStepsNotDone    = errors.New("predecessor steps not completed")
	ErrInvalidFlowDef     = errors.New("invalid flow definition: no steps")
)
