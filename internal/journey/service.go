package journey

import (
	"context"
	cryptorand "crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/faria/traveling-hub/internal/world"
	"github.com/google/uuid"
)

// Selector controls only the one-time plan selection. Its output is persisted
// at creation, so replacing it never changes an existing journey.
type Selector interface {
	Select(frogID uuid.UUID, localDay time.Time) (templateIndex int, returnMinute int, err error)
}

type SelectorFunc func(frogID uuid.UUID, localDay time.Time) (templateIndex int, returnMinute int, err error)

func (f SelectorFunc) Select(frogID uuid.UUID, localDay time.Time) (int, int, error) {
	return f(frogID, localDay)
}

type Service struct {
	repository Repository
	clock      world.Clock
	selector   Selector
}

func NewService(repository Repository, clock world.Clock, selector Selector) Service {
	if clock == nil {
		clock = world.SystemClock()
	}
	if selector == nil {
		selector = cryptoSelector{}
	}
	return Service{repository: repository, clock: clock, selector: selector}
}

func (s Service) ReconcileAndSnapshot(ctx context.Context, frogID uuid.UUID) (Snapshot, error) {
	return s.ReconcileAt(ctx, frogID, s.clock.Now())
}

// ReconcileAt is used by both a shared worker and request reads. It never
// creates a goroutine per frog; every side effect is safe to retry.
func (s Service) ReconcileAt(ctx context.Context, frogID uuid.UUID, now time.Time) (Snapshot, error) {
	if frogID == uuid.Nil {
		return Snapshot{}, fmt.Errorf("frog ID is required")
	}
	if now.IsZero() {
		return Snapshot{}, fmt.Errorf("current time is required")
	}
	if _, err := s.repository.AdvanceNarrativeStages(ctx, frogID, now); err != nil {
		return Snapshot{}, err
	}
	localNow := now.In(shanghai)
	today := localMidnight(now)
	ensureJourney := func(day time.Time) error {
		existing, err := s.repository.ForDate(ctx, frogID, day)
		if err != nil {
			return err
		}
		if existing != nil {
			return nil
		}
		templateIndex, returnMinute, err := s.selector.Select(frogID, day)
		if err != nil {
			return fmt.Errorf("select journey plan: %w", err)
		}
		planned, err := PlanDailyJourney(frogID, day, templateIndex, returnMinute)
		if err != nil {
			return err
		}
		if _, _, err := s.repository.CreateIfAbsent(ctx, planned); err != nil {
			return err
		}
		return nil
	}
	if localNow.Before(today.Add(8 * time.Hour)) {
		existed, err := s.repository.FrogExistedBefore(ctx, frogID, today)
		if err != nil {
			return Snapshot{}, err
		}
		if existed {
			if err := ensureJourney(today.AddDate(0, 0, -1)); err != nil {
				return Snapshot{}, err
			}
		}
	} else if err := ensureJourney(today); err != nil {
		return Snapshot{}, err
	}
	if _, err := s.repository.AdvanceNarrativeStages(ctx, frogID, now); err != nil {
		return Snapshot{}, err
	}
	current, err := s.repository.LatestForDisplay(ctx, frogID, localDate(now))
	if err != nil {
		return Snapshot{}, err
	}
	album, err := s.repository.AlbumPostcardIDs(ctx, frogID)
	if err != nil {
		return Snapshot{}, err
	}
	return BuildSnapshot(frogID, now, current, album)
}

type cryptoSelector struct{}

func (cryptoSelector) Select(_ uuid.UUID, _ time.Time) (int, int, error) {
	template, err := cryptorand.Int(cryptorand.Reader, big.NewInt(int64(len(Templates()))))
	if err != nil {
		return 0, 0, fmt.Errorf("select template: %w", err)
	}
	minute, err := cryptorand.Int(cryptorand.Reader, big.NewInt(4*60))
	if err != nil {
		return 0, 0, fmt.Errorf("select return minute: %w", err)
	}
	return int(template.Int64()), int(minute.Int64()), nil
}
