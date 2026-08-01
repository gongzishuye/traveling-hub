package journey

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestSnapshotHidesReturnTimeUntilJourneyHasReturned(t *testing.T) {
	departed := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	returnAt := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	travelling := DailyJourney{
		ID: uuid.New(), FrogID: uuid.New(), Status: Travelling,
		TemplateID: "willow-pond-reed", PostcardID: "willow-pond", FoodID: "light-meal",
		DepartedAt: departed, ReturnAt: returnAt, NextStage: 2,
	}
	snapshot, err := BuildSnapshot(travelling.FrogID, departed.Add(4*time.Hour), &travelling, nil)
	if err != nil {
		t.Fatalf("BuildSnapshot() error = %v", err)
	}
	if snapshot.Phase != Travelling || snapshot.Journey == nil || snapshot.Journey.ReturnAt != nil {
		t.Fatalf("travelling snapshot = %#v", snapshot)
	}
	if len(snapshot.Events) != 2 || snapshot.Events[0].Stage != 0 || snapshot.Events[1].Stage != 1 {
		t.Fatalf("travelling events = %#v", snapshot.Events)
	}

	returnedAt := returnAt
	returned := travelling
	returned.Status = Returned
	returned.ReturnedAt = &returnedAt
	returned.NextStage = 3
	snapshot, err = BuildSnapshot(returned.FrogID, returnedAt, &returned, []string{"willow-pond"})
	if err != nil {
		t.Fatalf("BuildSnapshot() returned error = %v", err)
	}
	if snapshot.Phase != Returned || snapshot.Journey == nil || snapshot.Journey.ReturnAt == nil || !snapshot.Journey.ReturnAt.Equal(returnAt) {
		t.Fatalf("returned snapshot = %#v", snapshot)
	}
	if len(snapshot.Events) != 4 || snapshot.Events[3].Stage != 3 || snapshot.AlbumPostcardIDs[0] != "willow-pond" {
		t.Fatalf("returned snapshot = %#v", snapshot)
	}
}

func TestSnapshotRejectsUnknownTemplate(t *testing.T) {
	journey := DailyJourney{ID: uuid.New(), FrogID: uuid.New(), Status: Travelling, TemplateID: "unknown", NextStage: 1}
	if _, err := BuildSnapshot(journey.FrogID, time.Now(), &journey, nil); err == nil {
		t.Fatal("BuildSnapshot() accepted an unknown template")
	}
}
