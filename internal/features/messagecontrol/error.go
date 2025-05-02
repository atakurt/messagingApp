package messagecontrol

import "fmt"

type SchedulerError struct {
	Operation string
	Err       error
}

func (e *SchedulerError) Error() string {
	return fmt.Sprintf("scheduler operation '%s' failed: %v", e.Operation, e.Err)
}

func (e *SchedulerError) Unwrap() error {
	return e.Err
}
