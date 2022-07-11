package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"

	"github.com/fiatjaf/go-nostr"
	drssnostr "github.com/plantimals/drss-nostr"
)

type config struct {
	StoragePath  string
	FeedURL      string
	DumpSchema   bool
	DumpJSONFeed bool
}

func parseFlags() *config {
	var storagePath string
	var feedURL string
	var dumpSchema bool
	var dumpJSONFeed bool
	flag.StringVar(&storagePath, "storage", "./feed", "path to construct feed")
	flag.StringVar(&feedURL, "feedURL", "https://ipfs.io/blog/index.xml", "feed URL")
	flag.BoolVar(&dumpSchema, "schema", false, "dump the IPFeed jsonschema")
	flag.BoolVar(&dumpJSONFeed, "toJSON", false, "dump the contents of the provided URL in jsonfeed format to stdout")
	flag.Parse()

	_, err := url.ParseRequestURI(feedURL)
	if err != nil {
		panic(err)
	}
	return &config{
		StoragePath:  storagePath,
		FeedURL:      feedURL,
		DumpSchema:   dumpSchema,
		DumpJSONFeed: dumpJSONFeed,
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
	config := parseFlags()
	drss, err := drssnostr.RSSToDRSS(config.FeedURL)
	if err != nil {
		panic(err)
	}
	fmt.Println(drss.DisplayName)
	fmt.Println(drss.PubKey)
	privKey := nostr.GeneratePrivateKey()
	for _, item := range drss.RSS.Items {
		ev, err := drssnostr.RSSItemToEvent(item, privKey, drss.PubKey)
		if err != nil {
			panic(err)
		}
		j, err := ev.MarshalJSON()
		if err != nil {
			panic(err)
		}
		event, _ := PrettyString(string(j))
		fmt.Println(event)
	}
}
