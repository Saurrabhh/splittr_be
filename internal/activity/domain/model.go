package domain

import (
	"encoding/json"
	"time"
)

// ActorInfo encapsulates performer details for an activity log.
type ActorInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
} // @name ActorInfo

// Activity represents an audit log entry for actions performed in the system.
type Activity struct {
	ID          string          `json:"id"`
	GroupID     *string         `json:"groupId,omitempty"`
	Actor       ActorInfo       `json:"actor"`
	ActionType  ActionType      `json:"actionType"`
	EntityType  EntityType      `json:"entityType"`
	EntityID    *string         `json:"entityId,omitempty"`
	Description string          `json:"description"`
	Payload     ActivityPayload `json:"payload,omitempty" swaggertype:"object"`
	CreatedAt   time.Time       `json:"createdAt"`
} // @name Activity

// UnmarshalJSON customizes deserialization of Activity to unmarshal payload into its concrete type.
func (a *Activity) UnmarshalJSON(b []byte) error {
	type Alias Activity
	aux := &struct {
		RawPayload json.RawMessage `json:"payload,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(a),
	}
	if err := json.Unmarshal(b, aux); err != nil {
		return err
	}
	if len(aux.RawPayload) > 0 {
		payload, err := UnmarshalPayload(a.EntityType, aux.RawPayload)
		if err != nil {
			return err
		}
		a.Payload = payload
	}
	return nil
}
