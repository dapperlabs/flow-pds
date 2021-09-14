package common

import (
	"database/sql/driver"
	"fmt"

	"github.com/onflow/flow-go-sdk"
)

type FlowAddress flow.Address

func (a FlowAddress) MarshalJSON() ([]byte, error) {
	return flow.Address(a).MarshalJSON()
}

func (a *FlowAddress) UnmarshalJSON(data []byte) error {
	b := flow.Address(*a)
	b.UnmarshalJSON(data)
	*a = FlowAddress(b)
	return nil
}

func (a FlowAddress) Value() (driver.Value, error) {
	return flow.Address(a).Bytes(), nil
}

func (a *FlowAddress) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal FlowAddress value: %v", value)
	}
	*a = FlowAddress(flow.BytesToAddress(bytes))
	return nil
}
