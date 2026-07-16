package models

import "time"

// Group represents a row in the `groups` table.
type Group struct {
	GroupID   string    `json:"group_id"`
	GroupName string    `json:"group_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
