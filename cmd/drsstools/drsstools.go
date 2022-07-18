package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"

	drssnostr "github.com/plantimals/drss-nostr"
)

type config struct {
	FeedURL     string
	PrivateKey  string
	PublicKeys  []string
	Relays      []string
	DisplayName string
}

type StringList []string

func (s *StringList) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *StringList) Set(value string) error {
	*s = append(*s, value)
	return nil // no error to return
}

var publicKeys, relays StringList

func parseFlags() *config {
	var feedURL string
	var privateKey string
	var displayName string
	flag.StringVar(&feedURL, "feedURL", "https://ipfs.io/blog/index.xml", "feed URL")
	flag.StringVar(&privateKey, "privateKey", "6f6285da349cc629bda7dd72f96dee872c3bfd93f31c2ab5e4ead47588d870b7", "private key")
	flag.StringVar(&displayName, "displayName", "plantimals", "display name")
	flag.Var(&publicKeys, "publicKeys", "public keys")
	flag.Var(&relays, "relays", "relay URLs")
	flag.Parse()

	_, err := url.ParseRequestURI(feedURL)
	if err != nil {
		panic(err)
	}
	return &config{
		FeedURL:    feedURL,
		PrivateKey: privateKey,
		PublicKeys: publicKeys,
		Relays:     relays,
	}
}

func PrettyString(str string) (string, error) {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(str), "", "    "); err != nil {
		return "", err
	}
	return prettyJSON.String(), nil
}

func main() {
	conf := parseFlags()

	dfeed := &drssnostr.DRSSFeed{
		DisplayName: conf.DisplayName,
		PubKeys:     conf.PublicKeys,
		PrivKey:     conf.PrivateKey,
		Relays:      conf.Relays,
		FeedURL:     conf.FeedURL,
	}
	if err := dfeed.AddRelays(); err != nil {
		panic(err)
	}
	drss2rss(dfeed)
	rss2drss(dfeed)
}

func drss2rss(f *drssnostr.DRSSFeed) {
	if err := f.DRSSToRSS(); err != nil {
		panic(err)
	}
	atom, err := f.RSS.ToAtom()
	if err != nil {
		panic(err)
	}
	fmt.Println(atom)
}

func rss2drss(f *drssnostr.DRSSFeed) {
	if err := f.RSSToDRSS(); err != nil {
		panic(err)
	}
	for _, ev := range f.Events {
		eventJson, err := json.Marshal(ev)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(eventJson))
	}
}
