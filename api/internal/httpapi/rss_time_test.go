package httpapi

import (
	"testing"
	"time"
)

func TestDaysSinceLastUpdate(t *testing.T) {
	location := mustLoadBerlinTestLocation(t)
	for _, testCase := range daysSinceLastUpdateCases(location) {
		t.Run(testCase.name, func(t *testing.T) {
			got := daysSinceLastUpdate(testCase.lastChanged, testCase.now, location)
			if got != testCase.wantDays {
				t.Fatalf("expected %d days, got %d", testCase.wantDays, got)
			}
		})
	}
}

func daysSinceLastUpdateCases(location *time.Location) []struct {
	name        string
	lastChanged time.Time
	now         time.Time
	wantDays    int
} {
	cases := []struct {
		name        string
		lastChanged time.Time
		now         time.Time
		wantDays    int
	}{
		{
			name:        "reset immediately after update",
			lastChanged: time.Date(2026, time.March, 10, 13, 0, 0, 0, location),
			now:         time.Date(2026, time.March, 10, 13, 30, 0, 0, location),
			wantDays:    0,
		},
		{
			name:        "increments at first noon checkpoint after update",
			lastChanged: time.Date(2026, time.March, 10, 11, 59, 0, 0, location),
			now:         time.Date(2026, time.March, 10, 12, 1, 0, 0, location),
			wantDays:    1,
		},
		{
			name:        "stays zero before first noon checkpoint",
			lastChanged: time.Date(2026, time.March, 10, 12, 1, 0, 0, location),
			now:         time.Date(2026, time.March, 11, 11, 59, 0, 0, location),
			wantDays:    0,
		},
		{
			name:        "counts consecutive noon checkpoints",
			lastChanged: time.Date(2026, time.March, 10, 12, 1, 0, 0, location),
			now:         time.Date(2026, time.March, 12, 12, 1, 0, 0, location),
			wantDays:    2,
		},
	}
	return append(cases, dstTransitionCase(location))
}

func dstTransitionCase(location *time.Location) struct {
	name        string
	lastChanged time.Time
	now         time.Time
	wantDays    int
} {
	return struct {
		name        string
		lastChanged time.Time
		now         time.Time
		wantDays    int
	}{
		name:        "handles dst transition in berlin",
		lastChanged: time.Date(2026, time.March, 28, 13, 0, 0, 0, location),
		now:         time.Date(2026, time.March, 30, 13, 0, 0, 0, location),
		wantDays:    2,
	}
}

func mustLoadBerlinTestLocation(t *testing.T) *time.Location {
	t.Helper()
	location, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Fatalf("load Europe/Berlin location: %v", err)
	}
	return location
}
