package domain

import "time"

// ExportPayload is the full GDPR data export response (Right to Access / Data Portability).
type ExportPayload struct {
	ExportedAt           time.Time       `json:"exported_at"`
	Account              AccountExport   `json:"account"`
	Profile              ProfileExport   `json:"profile"`
	Preferences          interface{}     `json:"preferences"`
	Photos               []PhotoExport   `json:"photos"`
	VoicePrompts         []VoicePromptExport `json:"voice_prompts"`
	Swipes               []SwipeExport   `json:"swipes"`
	Matches              []MatchExport   `json:"matches"`
	Conversations        []ConversationExport `json:"conversations"`
	Feedback             []FeedbackExport `json:"feedback"`
	Events               []EventExport   `json:"events"`
	InsightSnapshots     []InsightSnapshotExport `json:"insight_snapshots"`
	VerificationAttempts []VerificationAttemptExport `json:"verification_attempts"`
	Blocks               []BlockExport   `json:"blocks"`
	Reports              []ReportExport  `json:"reports"`
	MatchingAnswers      []UserAnswerExport `json:"matching_answers"`
}

// AccountExport is the user account record (no sensitive auth details).
type AccountExport struct {
	ID                   string     `json:"id"`
	Email                *string    `json:"email,omitempty"`
	FirstName            string     `json:"first_name"`
	LastName             *string    `json:"last_name,omitempty"`
	Phone                *string    `json:"phone,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
	OnboardingStep       string     `json:"onboarding_step"`
	HowDidYouHearAboutUs *string    `json:"how_did_you_hear_about_us,omitempty"`
}

// ProfileExport is the user profile for export.
type ProfileExport struct {
	UserID            string     `json:"user_id"`
	DisplayName       string     `json:"display_name"`
	Birthdate        *time.Time `json:"birthdate,omitempty"`
	HeightCM         *int       `json:"height_cm,omitempty"`
	Geo              string     `json:"geo"`
	City             *string    `json:"city,omitempty"`
	Country          *string    `json:"country,omitempty"`
	GenderID         *int       `json:"gender_id,omitempty"`
	DatingIntentionID *int      `json:"dating_intention_id,omitempty"`
	ReligionID       *int       `json:"religion_id,omitempty"`
	EducationLevelID *int       `json:"education_level_id,omitempty"`
	PoliticalBeliefID *int      `json:"political_belief_id,omitempty"`
	Work             *string    `json:"work,omitempty"`
	JobTitle         *string    `json:"job_title,omitempty"`
	University       *string    `json:"university,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	CoverMediaURL    *string    `json:"cover_media_url,omitempty"`
	Emoji            *string    `json:"emoji,omitempty"`
	SexualityID      *int       `json:"sexuality_id,omitempty"`
	Verified         string     `json:"verified"`
}

// PhotoExport is a single photo record.
type PhotoExport struct {
	ID        int64     `json:"id"`
	URL       string    `json:"url"`
	Position  *int      `json:"position,omitempty"`
	IsPrimary bool      `json:"is_primary"`
	CreatedAt time.Time `json:"created_at"`
}

// VoicePromptExport is a voice prompt record.
type VoicePromptExport struct {
	ID         int64     `json:"id"`
	PromptType *int      `json:"prompt_type,omitempty"`
	AudioURL   string    `json:"audio_url"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
}

// SwipeExport is a swipe (like/pass/superlike) record.
type SwipeExport struct {
	ID             int64     `json:"id"`
	ActorID        string    `json:"actor_id"`
	TargetID       string    `json:"target_id"`
	Action         string    `json:"action"`
	MessageType    *string   `json:"message_type,omitempty"`
	Message        *string   `json:"message,omitempty"`
	PromptID       *int64    `json:"prompt_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// MatchExport is a match record.
type MatchExport struct {
	ID         string     `json:"id"`
	UserA      string     `json:"user_a"`
	UserB      string     `json:"user_b"`
	CreatedAt  time.Time  `json:"created_at"`
	RevealedAt *time.Time `json:"revealed_at,omitempty"`
	Status     string     `json:"status"`
	DateMode   bool       `json:"date_mode"`
}

// ConversationExport is a conversation with its messages.
type ConversationExport struct {
	ID        string          `json:"id"`
	UserA     string          `json:"user_a"`
	UserB     string          `json:"user_b"`
	Messages  []MessageExport `json:"messages"`
}

// MessageExport is a single message.
type MessageExport struct {
	ID             int64     `json:"id"`
	SenderID       string    `json:"sender_id"`
	Type           string    `json:"type"`
	TextBody       *string   `json:"text_body,omitempty"`
	MediaKey       *string   `json:"media_key,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	EditedAt       *time.Time `json:"edited_at,omitempty"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`
}

// FeedbackExport is a feedback submission.
type FeedbackExport struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Title     *string   `json:"title,omitempty"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// EventExport is an analytics event.
type EventExport struct {
	ID         string                 `json:"id"`
	OccurredAt time.Time              `json:"occurred_at"`
	Name       string                 `json:"name"`
	Props      map[string]interface{} `json:"props,omitempty"`
	Version    int                    `json:"version"`
}

// InsightSnapshotExport is an insight snapshot.
type InsightSnapshotExport struct {
	ID          string                 `json:"id"`
	Key         string                 `json:"key"`
	PeriodStart time.Time              `json:"period_start"`
	PeriodEnd   time.Time              `json:"period_end"`
	Scope       string                 `json:"scope"`
	Payload     map[string]interface{} `json:"payload,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

// VerificationAttemptExport is a verification attempt.
type VerificationAttemptExport struct {
	ID             string    `json:"id"`
	Type           string    `json:"type"`
	Status         string    `json:"status"`
	SessionID      *string   `json:"session_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// BlockExport is a block record.
type BlockExport struct {
	BlockerUserID string    `json:"blocker_user_id"`
	BlockedUserID string    `json:"blocked_user_id"`
	Reason        *string   `json:"reason,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// ReportExport is a report where the user is reporter or reported.
type ReportExport struct {
	ID             string    `json:"id"`
	ReporterUserID string    `json:"reporter_user_id"`
	ReportedUserID string    `json:"reported_user_id"`
	Category       string    `json:"category"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

// UserAnswerExport is a matching question answer.
type UserAnswerExport struct {
	UserID     string    `json:"user_id"`
	QuestionID int64     `json:"question_id"`
	AnswerID   int64     `json:"answer_id"`
	Importance string    `json:"importance"`
	UpdatedAt  time.Time `json:"updated_at"`
}
