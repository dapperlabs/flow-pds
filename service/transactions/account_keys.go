package transactions

import (
	"github.com/google/uuid"
	"github.com/onflow/flow-go-sdk"
	"gorm.io/gorm"
)

type AccountKey struct {
	gorm.Model

	ID             uuid.UUID `gorm:"column:id;primary_key;type:uuid;"`
	State          KeyState
	SequenceNumber uint64
	KeyIndex       uint64

	FlowKey *flow.AccountKey `gorm:"-"`
}

func (AccountKey) TableName() string {
	return "account_keys"
}

func (t *AccountKey) BeforeCreate(tx *gorm.DB) (err error) {
	t.ID = uuid.New()
	return nil
}

type KeyState string

const (
	KeyStateInUse     KeyState = "in_use"
	KeyStateAvailable KeyState = "available"
)
