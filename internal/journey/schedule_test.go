package journey

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPlanDailyJourneyLocksMorningDepartureAndNightReturn(t *testing.T) {
	now := time.Date(2026, 8, 1, 0, 1, 0, 0, time.UTC)
	planned, err := PlanDailyJourney(uuid.New(), now, 1, 125)
	if err != nil {
		t.Fatalf("PlanDailyJourney() error = %v", err)
	}
	if planned.DepartedAt.Format(time.RFC3339) != "2026-08-01T00:00:00Z" {
		t.Fatalf("departed_at = %s", planned.DepartedAt)
	}
	if planned.ReturnAt.Format(time.RFC3339) != "2026-08-01T12:05:00Z" {
		t.Fatalf("return_at = %s", planned.ReturnAt)
	}
	if planned.TemplateID != "willow-pond-boat" || planned.NextStage != 1 || planned.Status != Travelling {
		t.Fatalf("planned journey = %#v", planned)
	}
}

func TestPlanDailyJourneyRejectsInvalidSelection(t *testing.T) {
	if _, err := PlanDailyJourney(uuid.New(), time.Now(), -1, 0); err == nil {
		t.Fatal("PlanDailyJourney() accepted an invalid template selection")
	}
	if _, err := PlanDailyJourney(uuid.New(), time.Now(), 0, 240); err == nil {
		t.Fatal("PlanDailyJourney() accepted an invalid return minute")
	}
}
