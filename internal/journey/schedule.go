package journey

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

var shanghai, _ = time.LoadLocation("Asia/Shanghai")

func PlanDailyJourney(frogID uuid.UUID, now time.Time, templateIndex, returnMinute int) (DailyJourney, error) {
	if frogID == uuid.Nil {
		return DailyJourney{}, fmt.Errorf("frog ID is required")
	}
	if templateIndex < 0 || templateIndex >= len(templates) {
		return DailyJourney{}, fmt.Errorf("invalid template selection")
	}
	if returnMinute < 0 || returnMinute >= 4*60 {
		return DailyJourney{}, fmt.Errorf("invalid return minute")
	}
	localNow := now.In(shanghai)
	localMidnight := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, shanghai)
	departed := localMidnight.Add(8 * time.Hour)
	returnAt := localMidnight.Add(18*time.Hour + time.Duration(returnMinute)*time.Minute)
	template := templates[templateIndex]
	return DailyJourney{
		ID: uuid.New(), FrogID: frogID, LocalDate: localMidnight, Status: Travelling,
		TemplateID: template.ID, PostcardID: template.PostcardID, FoodID: template.FoodID,
		DepartedAt: departed.UTC(), ReturnAt: returnAt.UTC(), NextStage: 1,
	}, nil
}
