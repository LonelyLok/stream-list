package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

var (
	maxResults = flag.Int64("max-results", 100, "Max YouTube results")
)

func prettyPrintStruct(v interface{}) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(string(b))
}

func test() {
	fmt.Println("Hello, World!")
	flag.Parse()

	API_KEY := os.Getenv("YOUTUBE_API_KEY")

	ctx := context.Background()

	service, err := youtube.NewService(ctx, option.WithAPIKey(API_KEY))

	if err != nil {
		log.Fatalf("Error creating new Youtube client: %v", err)
	}

	call := service.Search.List([]string{"snippet", "liveStreamingDetails"}).
		ChannelId("UCL_qhgtOy0dy1Agp8vkySQg").
		Type("video").
		EventType("upcoming").
		MaxResults(*maxResults)

	response, err := call.Do()

	if err != nil {
		log.Fatalf("Error making search API call: %v", err)
	}

	for _, item := range response.Items {
		prettyPrintStruct(item)
	}
}
