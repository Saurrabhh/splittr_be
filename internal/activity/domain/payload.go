package domain

import (
	"encoding/json"
	"fmt"
)

// ActivityPayload is a sealed interface for type-safe polymorphic activity metadata.
type ActivityPayload interface {
	PayloadType() EntityType
	isActivityPayload()
}

// ExpensePayload represents metadata snapshot for EXPENSE entity activities.
type ExpensePayload struct {
	Type    EntityType  `json:"type"`
	Expense interface{} `json:"expense"`
	Splits  interface{} `json:"splits,omitempty"`
}

func (ExpensePayload) PayloadType() EntityType { return EntityTypeExpense }
func (ExpensePayload) isActivityPayload()      {}

// SettlementPayload represents metadata snapshot for SETTLEMENT entity activities.
type SettlementPayload struct {
	Type    EntityType  `json:"type"`
	Expense interface{} `json:"expense"`
	Split   interface{} `json:"split,omitempty"`
}

func (SettlementPayload) PayloadType() EntityType { return EntityTypeSettlement }
func (SettlementPayload) isActivityPayload()      {}

// MemberPayload represents metadata snapshot for MEMBER entity activities.
type MemberPayload struct {
	Type   EntityType  `json:"type"`
	Member interface{} `json:"member"`
}

func (MemberPayload) PayloadType() EntityType { return EntityTypeMember }
func (MemberPayload) isActivityPayload()      {}

// GroupPayload represents metadata snapshot for GROUP entity activities.
type GroupPayload struct {
	Type    EntityType  `json:"type"`
	Group   interface{} `json:"group"`
	Members interface{} `json:"members,omitempty"`
}

func (GroupPayload) PayloadType() EntityType { return EntityTypeGroup }
func (GroupPayload) isActivityPayload()      {}

// UnmarshalPayload deserializes raw JSON into the correct concrete ActivityPayload based on EntityType.
func UnmarshalPayload(entityType EntityType, data []byte) (ActivityPayload, error) {
	if len(data) == 0 || string(data) == "null" {
		return nil, nil
	}
	switch entityType {
	case EntityTypeExpense:
		var p ExpensePayload
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, err
		}
		p.Type = EntityTypeExpense
		return p, nil
	case EntityTypeSettlement:
		var p SettlementPayload
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, err
		}
		p.Type = EntityTypeSettlement
		return p, nil
	case EntityTypeMember:
		var p MemberPayload
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, err
		}
		p.Type = EntityTypeMember
		return p, nil
	case EntityTypeGroup:
		var p GroupPayload
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, err
		}
		p.Type = EntityTypeGroup
		return p, nil
	default:
		return nil, fmt.Errorf("unknown entity type: %s", entityType)
	}
}
