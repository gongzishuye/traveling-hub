package journey

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

const departureText = "旅人出发了。"

// BuildSnapshot maps durable journey state to the read-only game contract.
// A journey's locked timestamps, not the browser clock, determine which
// narrative entries are visible.
func BuildSnapshot(frogID uuid.UUID, now time.Time, current *DailyJourney, albumPostcardIDs []string) (Snapshot, error) {
	localNow := now.In(shanghai)
	snapshot := Snapshot{
		FrogID:           frogID,
		ServerTime:       localNow,
		LocalDate:        localNow.Format("2006-01-02"),
		Phase:            Home,
		Events:           []SnapshotEvent{},
		AlbumPostcardIDs: append([]string(nil), albumPostcardIDs...),
	}
	if current == nil {
		return snapshot, nil
	}
	if current.FrogID != frogID {
		return Snapshot{}, fmt.Errorf("journey does not belong to frog")
	}
	template, found := TemplateByID(current.TemplateID)
	if !found || template.PostcardID != current.PostcardID || template.FoodID != current.FoodID {
		return Snapshot{}, fmt.Errorf("unknown or inconsistent journey template %q", current.TemplateID)
	}
	view := &JourneyView{
		TemplateID: current.TemplateID,
		PostcardID: current.PostcardID,
		FoodID:     current.FoodID,
		DepartedAt: current.DepartedAt.In(shanghai),
	}
	if current.Status == Returned {
		returnedAt := current.ReturnAt.In(shanghai)
		view.ReturnAt = &returnedAt
	}
	snapshot.Phase = current.Status
	snapshot.Journey = view
	snapshot.Events = visibleEvents(*current, template)
	return snapshot, nil
}

func visibleEvents(current DailyJourney, template Template) []SnapshotEvent {
	events := []SnapshotEvent{{Stage: 0, OccurredAt: current.DepartedAt.In(shanghai), Text: departureText}}
	visibleTemplateStages := current.NextStage - 1
	if current.Status == Returned {
		visibleTemplateStages = len(template.Events)
	}
	if visibleTemplateStages < 0 {
		visibleTemplateStages = 0
	}
	if visibleTemplateStages > len(template.Events) {
		visibleTemplateStages = len(template.Events)
	}
	for index := 0; index < visibleTemplateStages; index++ {
		events = append(events, SnapshotEvent{
			Stage:      index + 1,
			OccurredAt: stageDueAt(current, index+1).In(shanghai),
			Text:       template.Events[index],
		})
	}
	return events
}

func stageDueAt(journey DailyJourney, stage int) time.Time {
	duration := journey.ReturnAt.Sub(journey.DepartedAt)
	switch stage {
	case 1:
		return journey.DepartedAt.Add(duration / 4)
	case 2:
		return journey.DepartedAt.Add(duration / 2)
	default:
		return journey.ReturnAt
	}
}

func localDate(now time.Time) string { return now.In(shanghai).Format("2006-01-02") }

func localMidnight(now time.Time) time.Time {
	localNow := now.In(shanghai)
	return time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, shanghai)
}
