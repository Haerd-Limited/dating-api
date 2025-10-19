package realtime

import "github.com/google/uuid"

func NewEventID() string {
	id, _ := uuid.NewV7() // handle err if you prefer; V7 is time-ordered
	return id.String()
}
