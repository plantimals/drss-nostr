package drssnostr

import (
	"context"
	"time"

	nostr "github.com/fiatjaf/go-nostr"
	"github.com/mmcdole/gofeed"
	log "github.com/sirupsen/logrus"
)

type PublicKey string

type Feed struct {
	DisplayName string
	PubKey      PublicKey
	RSS         *gofeed.Feed
	Items       []*nostr.Event
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

func RSSToDRSS(RSSurl string) (*Feed, error) {
	feed, err := GetRSSFeed(RSSurl)
	if err != nil {
		return nil, err
	}
	pk := nostr.GeneratePrivateKey()
	pubk, err := nostr.GetPublicKey(pk)
	if err != nil {
		return nil, err
	}
	return &Feed{
		DisplayName: feed.Title,
		PubKey:      PublicKey(pubk),
		RSS:         feed,
	}, nil
}

func RSSItemToEvent(item *gofeed.Item, privateKey string, pubKey PublicKey) (*nostr.Event, error) {
	/*content := ""
	if item.Content != "" {
		content = item.Content
	} else {
		content = item.Description
	}*/
	content := item.Description
	if len(content) > 250 {
		content = content[:250]
		log.Info("shortened description")
	}
	n := nostr.Event{
		CreatedAt: time.Now(),
		Kind:      nostr.KindTextNote,
		Tags:      make(nostr.Tags, 0),
		Content:   content,
		PubKey:    string(pubKey),
	}
	n.ID = string(n.Serialize())
	n.Sign(privateKey)
	return &n, nil
}
