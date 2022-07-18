package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"

	drssnostr "github.com/plantimals/drss-nostr"
	log "github.com/sirupsen/logrus"
)

type config struct {
	FeedURL     string
	PrivateKey  string
	PublicKeys  []string
	Relays      []string
	DisplayName string
	Cmd         string
}

type StringList []string

func (s *StringList) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *StringList) Set(value string) error {
	*s = append(*s, value)
	return nil // no error to return
}

func parseFlags() *config {
	var feedURL, privateKey, cmd, displayName string

	var publicKeys, relays StringList

	flag.StringVar(&feedURL, "feedURL", "https://ipfs.io/blog/index.xml", "feed URL")
	flag.StringVar(&privateKey, "privateKey", "6f6285da349cc629bda7dd72f96dee872c3bfd93f31c2ab5e4ead47588d870b7", "private key")
	flag.StringVar(&displayName, "displayName", "plantimals", "display name")
	flag.StringVar(&cmd, "cmd", "d2r", "d2r or r2d")
	flag.Var(&publicKeys, "publicKeys", "public keys")
	flag.Var(&relays, "relays", "relay URLs")
	flag.Parse()

	if len(os.Args) < 2 {
		flag.Usage()
		os.Exit(1)
	}

	_, err := url.ParseRequestURI(feedURL)
	if err != nil {
		panic(err)
	}

	return &config{
		FeedURL:     feedURL,
		PrivateKey:  privateKey,
		PublicKeys:  publicKeys,
		Relays:      relays,
		DisplayName: displayName,
		Cmd:         cmd,
	}
}

func PrettyString(str string) (string, error) {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(str), "", "    "); err != nil {
		return "", err
	}
	return prettyJSON.String(), nil
}

func drss2rss(f *drssnostr.DRSSFeed) {
	log.Info("converting DRSS to RSS")
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
	f.PublishNostr()
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
	log.Info(fmt.Sprintf("found %d publicKeys in conf", len(conf.PublicKeys)))
	log.Info(fmt.Sprintf("found %d publicKeys", len(dfeed.PubKeys)))
	switch conf.Cmd {
	case "d2r":
		drss2rss(dfeed)
	case "r2d":
		rss2drss(dfeed)
		feedString, err := dfeed.ToString()
		if err != nil {
			panic(err)
		}
		fmt.Println(PrettyString(feedString))
	}
}
