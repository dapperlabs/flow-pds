package app

import (
	"fmt"

	"github.com/flow-hydraulics/flow-pds/service/common"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CirculatingPackContract represents the contract of a pack NFT that has been
// put into circulation.
// We need to monitor each circulating packs events.
type CirculatingPackContract struct {
	gorm.Model
	ID uuid.UUID `gorm:"column:id;primary_key;type:uuid;"`

	Name    string             `gorm:"column:name;uniqueIndex:name_address"`
	Address common.FlowAddress `gorm:"column:address;uniqueIndex:name_address"`

	StartAtBlock uint64 `gorm:"column:start_at_block"`
}

func (CirculatingPackContract) TableName() string {
	return "circulating_packs"
}

func (c *CirculatingPackContract) BeforeCreate(tx *gorm.DB) (err error) {
	c.ID = uuid.New()
	return nil
}

func (c CirculatingPackContract) String() string {
	return AddressLocation{Name: c.Name, Address: c.Address}.String()
}

func (c CirculatingPackContract) EventName(event string) string {
	return fmt.Sprintf("%s.%s", c, event)
}

type CirculatingPackContractBlockCursor struct {
	gorm.Model

	ID           uuid.UUID `gorm:"column:id;primary_key;type:uuid;"`
	EventName    string    `gorm:"unique"`
	StartAtBlock uint64    `gorm:"column:start_at_block"`
}

func (CirculatingPackContractBlockCursor) TableName() string {
	return "circulating_packs_block_cursor"
}

func (c *CirculatingPackContractBlockCursor) BeforeCreate(tx *gorm.DB) (err error) {
	c.ID = uuid.New()
	return nil
}
