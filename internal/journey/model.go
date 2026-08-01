package journey

import (
	"time"

	"github.com/google/uuid"
)

type Template struct {
	ID         string
	FoodID     string
	PostcardID string
	Events     [3]string
}

type Status string

const (
	Home       Status = "home"
	Travelling Status = "travelling"
	Returned   Status = "returned"
)

type DailyJourney struct {
	ID         uuid.UUID
	FrogID     uuid.UUID
	LocalDate  time.Time
	Status     Status
	TemplateID string
	PostcardID string
	FoodID     string
	DepartedAt time.Time
	ReturnAt   time.Time
	NextStage  int
	ReturnedAt *time.Time
}

// Snapshot is the server-owned view of one traveler's current or most recent
// daily journey. It deliberately has no command fields: callers can observe
// the rhythm but cannot advance it.
type Snapshot struct {
	FrogID           uuid.UUID       `json:"frog_id"`
	ServerTime       time.Time       `json:"server_time"`
	LocalDate        string          `json:"local_date"`
	Phase            Status          `json:"phase"`
	Journey          *JourneyView    `json:"journey,omitempty"`
	Events           []SnapshotEvent `json:"events"`
	AlbumPostcardIDs []string        `json:"album_postcard_ids"`
}

type JourneyView struct {
	TemplateID string     `json:"template_id"`
	PostcardID string     `json:"postcard_id"`
	FoodID     string     `json:"food_id"`
	DepartedAt time.Time  `json:"departed_at"`
	ReturnAt   *time.Time `json:"return_at,omitempty"`
}

type SnapshotEvent struct {
	Stage      int       `json:"stage"`
	OccurredAt time.Time `json:"occurred_at"`
	Text       string    `json:"text"`
}
