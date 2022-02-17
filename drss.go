package drssnostr

import (
	"context"
	"time"

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

func RSSToDRSS(RSSurl string) (*Feed, error) {
	feed, err := GetRSSFeed(RSSurl)
	if err != nil {
		return nil, err
	}
	return &Feed{
		DisplayName: feed.Title,
		PubKey:      PublicKey(feed.Link),
		RSS:         feed,
	}, nil
}

func GetRSSFeed(url string) (*gofeed.Feed, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	fp := gofeed.NewParser()
	feed, err := fp.ParseURLWithContext(url, ctx)
	if err != nil {
		return nil, err
	}
	return feed, nil
}
