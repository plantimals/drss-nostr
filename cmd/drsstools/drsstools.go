package main

import (
	"flag"
	"fmt"
	"net/url"

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

func main() {
	config := parseFlags()
	drss, err := drssnostr.RSSToDRSS(config.FeedURL)
	if err != nil {
		panic(err)
	}
	fmt.Println(drss.DisplayName)
	/*json, err := cid.MarshalJSON()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json))*/

}
