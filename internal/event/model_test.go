package event

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCursorRoundTrip(t *testing.T) {
	want := Event{ID: uuid.New(), WorldVersion: 4, OccurredAt: time.Date(2026, 7, 30, 0, 0, 0, 123, time.UTC), Payload: json.RawMessage(`{}`)}
	version, at, id, err := decodeCursor(encodeCursor(want))
	if err != nil || version != want.WorldVersion || !at.Equal(want.OccurredAt) || id != want.ID {
		t.Fatalf("cursor round trip = (%d, %v, %v, %v)", version, at, id, err)
	}
}
