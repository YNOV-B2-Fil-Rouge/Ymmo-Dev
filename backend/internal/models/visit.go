package models

import "time"

// Visit status values (must match the ENUM in db/schema.sql).
const (
	VisitStatusRequested = "REQUESTED"
	VisitStatusConfirmed = "CONFIRMED"
	VisitStatusCancelled = "CANCELLED"
	VisitStatusCompleted = "COMPLETED"
)

// Visit maps the `visits` table: a client asks to visit a property, handled
// by the property's agent.
type Visit struct {
	ID          uint      `gorm:"primaryKey;column:id" json:"id"`
	PropertyID  uint      `gorm:"column:property_id" json:"property_id"`
	ClientID    uint      `gorm:"column:client_id" json:"client_id"`
	AgentID     uint      `gorm:"column:agent_id" json:"agent_id"`
	ScheduledAt time.Time `gorm:"column:scheduled_at" json:"scheduled_at"`
	Status      string    `gorm:"column:status" json:"status"`
	Notes       *string   `gorm:"column:notes" json:"notes,omitempty"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
}

func (Visit) TableName() string { return "visits" }
