package models

import (
	"time"

	"github.com/google/uuid"
)

type Block struct {
	Name   string                 `json:"name"`
	Params map[string]interface{} `json:"params,omitempty"`
}

type Wod struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	Level       string    `json:"level"`
	DurationMin int       `json:"duration_min"`
	Equipment   []string  `json:"equipment,omitempty"`
	Seed        string    `json:"seed"`
	Blocks      []Block   `json:"blocks"`
}
