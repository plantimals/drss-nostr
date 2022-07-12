package drssnostr

import (
	"context"
	"time"

	nostr "github.com/fiatjaf/go-nostr"
	"github.com/mmcdole/gofeed"
)

type Feed struct {
	DisplayName  string
	PubKey       string
	RSS          *gofeed.Feed
	Events       []*nostr.Event
	LastItemGUID string
}

type DRSSIdentity struct {
	PrivKey string
	PubKey  string
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

func RSSToDRSS(RSSurl string, privKey string) (*Feed, error) {
	feed, err := GetRSSFeed(RSSurl)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	nostrEvents := make([]*nostr.Event, 0)

	for _, item := range feed.Items {
		ev, err := RSSItemToEvent(item, privKey)
		if err != nil {
			panic(err)
		}
		nostrEvents = append(nostrEvents, ev)
	}

	pubKey, err := nostr.GetPublicKey(privKey)
	if err != nil {
		return nil, err
	}

	return &Feed{
		DisplayName: feed.Title,
		PubKey:      pubKey,
		RSS:         feed,
		Events:      nostrEvents,
	}, nil
}

func RSSItemToEvent(item *gofeed.Item, privateKey string) (*nostr.Event, error) {
	content := item.Content
	if len(content) == 0 {
		content = item.Description
	}

	if len(content) > 250 {
		content += content[0:249] + "…"
	}
	content += "\n\n" + item.Link

	pubkey, err := nostr.GetPublicKey(privateKey)
	if err != nil {
		return nil, err
	}

	createdAt := time.Now()
	if item.UpdatedParsed != nil {
		createdAt = *item.UpdatedParsed
	}
	if item.PublishedParsed != nil {
		createdAt = *item.PublishedParsed
	}

	n := nostr.Event{
		CreatedAt: createdAt,
		Kind:      nostr.KindTextNote,
		Tags:      nostr.Tags{},
		Content:   content,
		PubKey:    pubkey,
	}
	n.ID = string(n.Serialize())
	n.Sign(privateKey)
	return &n, nil
}

func DRSSToRSS(events []*nostr.Event) (*gofeed.Feed, error) {
	feed := &gofeed.Feed{
		Items: make([]*gofeed.Item, 0),
	}
	for _, ev := range events {
		item, err := EventToItem(ev)
		if err != nil {
			return nil, err
		}
		feed.Items = append(feed.Items, item)
	}
	return feed, nil
}

func EventToItem(event *nostr.Event) (*gofeed.Item, error) {
	item := &gofeed.Item{
		Content: event.Content,
	}
	return item, nil
}
