package common

import "fmt"

type InvalidPackStateError struct {
	Expected []PackState
	Actual   PackState
}

func (m *InvalidPackStateError) Error() string {
	return fmt.Sprintf("pack in unexpected state expected: %+v actual:%s", m.Expected, m.Actual)
}

func (m *InvalidPackStateError) Is(target error) bool {
	_, ok := target.(*InvalidPackStateError)
	return ok
}
