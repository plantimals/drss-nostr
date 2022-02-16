package drssnostr

import (
	nostr "github.com/fiatjaf/go-nostr"
	"github.com/mmcdole/gofeed"
)

type PublicKey string

type Feed struct {
	DisplayName string
	PubKey      PublicKey
	RSS         *gofeed.Feed
	Items       []*nostr.Event
}

func rssToDRSS(RSSurl string) (*Feed, error) {
	return nil, nil
}
