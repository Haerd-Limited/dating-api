package dto

type PresenceRequest struct {
	ResourceType string `json:"resource_type" validate:"required"`
	ResourceID   string `json:"resource_id" validate:"required"`
}

type EventResponse struct {
	Type         string `json:"type"`
	ResourceType string `json:"resource_type,omitempty"`
	ResourceID   string `json:"resource_id,omitempty"`
	ActorName    string `json:"actor_name,omitempty"`
	Status       string `json:"status,omitempty"`
	OccurredAt   string `json:"occurred_at"`
}

type ListEventsResponse struct {
	Events []EventResponse `json:"events"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
}
