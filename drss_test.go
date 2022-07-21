package drssnostr

import (
	"testing"

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
