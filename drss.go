package drssnostr

import (
	"context"
	"fmt"
	"time"

	nostr "github.com/fiatjaf/go-nostr"
	"github.com/gorilla/feeds"
	"github.com/mmcdole/gofeed"
	log "github.com/sirupsen/logrus"
)

type Feed struct {
	DisplayName  string
	PubKey       string
	RSS          *gofeed.Feed
	Events       []*nostr.Event
	LastItemGUID string
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

func DRSSToRSS(title string) (string, error) {
	rp := nostr.NewRelayPool()
	err := rp.Add("wss://nostr.drss.io", nil)
	if err != nil {
		return "", err
	}
	keys := make([]string, 0)
	keys = append(keys, "dd81a8bacbab0b5c3007d1672fb8301383b4e9583d431835985057223eb298a5")
	sub := rp.Sub(nostr.Filters{{
		Authors: keys,
		Kinds:   nostr.IntList{nostr.KindTextNote},
	}})

	fmt.Println("before pull events")

	//events := make([]*nostr.Event, 0)

	items := make([]*feeds.Item, 0)
	go func() {
		for e := range sub.UniqueEvents {
			fmt.Println("default")
			item, err := EventToItem(&e)
			if err != nil {
				log.Error(err)
				return
			}
			items = append(items, item)

			fmt.Println(item)
		}
		log.Info("done pulling events")
	}()
	time.Sleep(3 * time.Second)
	sub.Unsub()

	if err != nil {
		return "", err
	}

	feed := &feeds.Feed{
		Title:       title,
		Created:     time.Now(),
		Link:        &feeds.Link{Href: fmt.Sprintf("https://nostr.com/p/%s", keys[0])},
		Description: fmt.Sprintf("drss feed generated from nostr events by the public key: %s", keys[0]),
		Items:       items,
	}
	answer, err := feed.ToAtom()
	if err != nil {
		return "", err
	}
	return answer, nil
}

func EventToItem(event *nostr.Event) (*feeds.Item, error) {
	item := &feeds.Item{
		Author:  &feeds.Author{Name: event.PubKey},
		Content: event.Content,
		Created: event.CreatedAt,
		Link:    &feeds.Link{Href: fmt.Sprintf("https://nostr.com/e/%s", event.ID)},
		Id:      event.ID,
	}
	return item, nil
}
