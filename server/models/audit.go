package models

import "time"

type Audit struct {
	ID        int       `json:"id"`
	Operation string    `json:"operation"`
	UserID    int       `json:"user_id"`
	Timestamp time.Time `json:"timestamp"`
	RecordData
}
