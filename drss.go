package drssnostr

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	nostr "github.com/fiatjaf/go-nostr"
	"github.com/gorilla/feeds"  //composes RSS from nostr events
	"github.com/mmcdole/gofeed" //parses RSS from nostr events
	log "github.com/sirupsen/logrus"
)

// DRSSFeed is a collection of data required to go from RSS to nostr and back
type DRSSFeed struct {
	DisplayName  string           `json:"display_name,omitempty"`
	PubKeys      []string         `json:"pub_keys,omitempty"`
	PrivKey      string           `json:"priv_key,omitempty"`
	Relays       []string         `json:"relays,omitempty"`
	Pools        *nostr.RelayPool `json:"-"`
	FeedURL      string           `json:"feed_url,omitempty"`
	RSS          *feeds.Feed      `json:"-"`
	Events       []*nostr.Event   `json:"-"`
	LastItemGUID string           `json:"-"`
}

// NewFeed parses a json representation of a DRSSFeed and returns a DRSSFeed
func NewFeed(j []byte) (*DRSSFeed, error) {
	var feed DRSSFeed
	err := json.Unmarshal(j, &feed)
	if err != nil {
		return nil, err
	}
	if feed.Relays != nil {
		feed.AddRelays()
	}
	return &feed, nil
}

// DRSSFeed dump the feed to a json string
func (f *DRSSFeed) ToString() (string, error) {
	answer, err := json.Marshal(f)
	if err != nil {
		return "", err
	}
	return string(answer), nil
}

// AddRelays adds relays to the feed
func (f *DRSSFeed) AddRelays(relays ...string) error {
	if f.Relays == nil {
		f.Relays = make([]string, 0)
	}
	for _, r := range relays {
		f.Relays = append(f.Relays, r)
	}
	f.Pools = nostr.NewRelayPool()
	for _, r := range f.Relays {
		err := f.Pools.Add(r, nostr.SimplePolicy{Read: true, Write: true})
		if err != nil {
			return nil
		}
	}
	f.Pools.SecretKey = &f.PrivKey
	return nil
}

// GetRSSFeed pulls an RSS feed from the URL and parses it into a feeds.Feed
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

// RSSToDRSS converts an RSS feed to a DRSS feed, a collection of nostr events
func (f *DRSSFeed) RSSToDRSS() error {

	//check that feed object has necessary inputs for this operation
	if f.PrivKey == "" {
		return fmt.Errorf("no priv key")
	} else if f.FeedURL == "" {
		return fmt.Errorf("no feed url")
	} else if f.Relays == nil {
		return fmt.Errorf("no relays")
	}

	feed, err := GetRSSFeed(f.FeedURL)
	if err != nil {
		return err
	}

	nostrEvents := make([]*nostr.Event, 0)

	for _, item := range feed.Items {
		ev, err := RSSItemToEvent(item, f.PrivKey)
		if err != nil {
			panic(err)
		}
		nostrEvents = append(nostrEvents, ev)
	}

	pubKey, err := nostr.GetPublicKey(f.PrivKey)
	if err != nil {
		return err
	}

	f.PubKeys = append(f.PubKeys, pubKey)
	f.Events = nostrEvents

	return nil
}

// PublishNostr publishes the nostr events in the feed to the relays
func (f *DRSSFeed) PublishNostr() error {
	//check that feed object has necessary inputs for this operation
	if f.Events == nil || len(f.Events) == 0 {
		return fmt.Errorf("no events")
	}
	if f.Pools == nil {
		return fmt.Errorf("no pools")
	}
	pk, err := nostr.GetPublicKey(f.PrivKey)
	if err != nil {
		return err
	}
	log.Info("Publishing nostr events under public key: " + pk)

	for _, ev := range f.Events {
		_, _, err := f.Pools.PublishEvent(ev)
		log.Info("Published event: " + ev.ID)
		if err != nil {
			return err
		}
	}
	time.Sleep(3 * time.Second)
	return nil
}

// RSSItemToEvent converts a RSS item and private key into a nostr event
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

func (f *DRSSFeed) GetEvents() error {

	sub := f.Pools.Sub(nostr.Filters{{
		Authors: nostr.StringList(f.PubKeys),
		Kinds:   nostr.IntList{nostr.KindTextNote},
	}})

	events := make([]*nostr.Event, 0)
	//launch a goroutine to listen to the relay
	go func() {
		for e := range sub.UniqueEvents {
			events = append(events, &e)
		}
	}()

	//wait to receive all events then close the subscription
	time.Sleep(1 * time.Second)
	sub.Unsub()

	f.Events = events
	return nil
}

// DRSSToRSS converts a DRSS feed to a RSS feed
// takes n public keys and compiles them into a feed
func (f *DRSSFeed) DRSSToRSS() error {

	if err := f.GetEvents(); err != nil {
		return err
	}

	items := make([]*feeds.Item, 0)
	for _, ev := range f.Events {
		item, err := EventToItem(ev)
		if err != nil {
			return err
		}
		items = append(items, item)
	}

	f.RSS = &feeds.Feed{
		Title:       f.DisplayName,
		Created:     time.Now(),
		Link:        &feeds.Link{Href: fmt.Sprintf("https://nostr.com/p/%s", f.PubKeys[0])},
		Description: fmt.Sprintf("drss feed generated from nostr events by the public key: %s", f.PubKeys[0]),
		Items:       items,
	}
	return nil
}

// EventToItem converts a nostr event to a RSS item
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
