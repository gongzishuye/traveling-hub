package event

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
)

var typePattern = regexp.MustCompile(`^[a-z0-9_.-]{1,64}$`)

type Draft struct {
	Type         string
	WorldVersion int64
	OccurredAt   time.Time
	Payload      json.RawMessage
	Source       string
	Audience     []uuid.UUID
}

type Event struct {
	ID           uuid.UUID       `json:"id"`
	Type         string          `json:"type"`
	WorldVersion int64           `json:"world_version"`
	OccurredAt   time.Time       `json:"occurred_at"`
	Payload      json.RawMessage `json:"payload"`
}

type cursor struct {
	Version int64  `json:"v"`
	Time    string `json:"t"`
	ID      string `json:"i"`
}

func encodeCursor(e Event) string {
	data, _ := json.Marshal(cursor{Version: e.WorldVersion, Time: e.OccurredAt.UTC().Format(time.RFC3339Nano), ID: e.ID.String()})
	return base64.RawURLEncoding.EncodeToString(data)
}

func decodeCursor(value string) (int64, time.Time, uuid.UUID, error) {
	if value == "" {
		return 0, time.Unix(0, 0).UTC(), uuid.Nil, nil
	}
	data, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return 0, time.Time{}, uuid.Nil, fmt.Errorf("invalid cursor")
	}
	var parsed cursor
	if err := json.Unmarshal(data, &parsed); err != nil {
		return 0, time.Time{}, uuid.Nil, fmt.Errorf("invalid cursor")
	}
	at, err := time.Parse(time.RFC3339Nano, parsed.Time)
	if err != nil {
		return 0, time.Time{}, uuid.Nil, fmt.Errorf("invalid cursor")
	}
	id, err := uuid.Parse(parsed.ID)
	if err != nil {
		return 0, time.Time{}, uuid.Nil, fmt.Errorf("invalid cursor")
	}
	return parsed.Version, at, id, nil
}
