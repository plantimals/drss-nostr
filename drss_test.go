package drssnostr

import (
	"testing"
	"time"

	nostr "github.com/fiatjaf/go-nostr"
)

//TestUniquify tests that events are collapsed based on their IDs
func TestUniquifyEvents(t *testing.T) {
	events := []*nostr.Event{
		{ID: "1"},
		{ID: "2"},
		{ID: "3"},
		{ID: "1"},
		{ID: "2"},
		{ID: "1"},
	}
	uniquified := UniquifyEvents(events)
	if len(uniquified) != 3 {
		t.Errorf("Expected 3 unique events, got %d", len(uniquified))
	}
}

func TestSortEventsDateDesc(t *testing.T) {
	events := []*nostr.Event{
		{CreatedAt: time.Time(time.Now().Add(1 * time.Hour)), ID: "0"},
		{CreatedAt: time.Time(time.Now()), ID: "1"},
		{CreatedAt: time.Time(time.Now().Add(-1 * time.Hour)), ID: "2"},
	}
	sorted := SortEventsDateDesc(events)
	if sorted[0].ID != "2" {
		t.Errorf("Expected event 2, got %s", sorted[0].ID)
	}
	if sorted[1].ID != "1" {
		t.Errorf("Expected event 1, got %s", sorted[1].ID)
	}
	if sorted[2].ID != "0" {
		t.Errorf("Expected event 0, got %s", sorted[2].ID)
	}
}
