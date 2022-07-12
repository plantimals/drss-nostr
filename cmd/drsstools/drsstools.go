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
	StoragePath  string
	FeedURL      string
	PrivateKey   string
	DumpSchema   bool
	DumpJSONFeed bool
}

func parseFlags() *config {
	var storagePath string
	var feedURL string
	var privateKey string
	var dumpSchema bool
	var dumpJSONFeed bool
	flag.StringVar(&storagePath, "storage", "./feed", "path to construct feed")
	flag.StringVar(&feedURL, "feedURL", "https://ipfs.io/blog/index.xml", "feed URL")
	flag.StringVar(&privateKey, "privateKey", "6f6285da349cc629bda7dd72f96dee872c3bfd93f31c2ab5e4ead47588d870b7", "private key")
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
		PrivateKey:   privateKey,
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
	drss, err := drssnostr.RSSToDRSS(config.FeedURL, config.PrivateKey)
	if err != nil {
		panic(err)
	}

	for _, item := range drss.Events {
		j, err := item.MarshalJSON()
		if err != nil {
			panic(err)
		}
		event, _ := PrettyString(string(j))
		fmt.Println(event)
	}
}
