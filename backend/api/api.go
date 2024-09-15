package api

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
	maxResults = flag.Int64("max-results", 50, "Max YouTube results")
)

var streamers = []Streamer{
	{
		"UCL_qhgtOy0dy1Agp8vkySQg",
		"Mori Calliope",
	},
	{
		"UC8rcEBzJSleTkf_-agPM20g",
		"IRyS",
	},
	{
		"UCHsx4Hqa-1ORjQTh9TYDhww",
		"Takanashi Kiara",
	},
	{
		"UCgmPnx-EEeOrZSg5Tiw7ZRQ",
		"Hakos Baelz",
	},
	{
		"UCgnfPPb9JI3e9A4cXHnWbyg",
		"Shiori Novella",
	},
}

type Streamer struct {
	channelId string
	name      string
}

func prettyPrintStruct(v interface{}) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(string(b))
}

type VideoInfo struct {
	ID                 string                    `json:"id"`
	Title              string                    `json:"title"`
	ScheduledStartTime string                    `json:"scheduledStartTime"`
	Thumbnails         *youtube.ThumbnailDetails `json:"thumbnails"`
}

type StreamerInfo struct {
	ChannelID string      `json:"channelId"`
	Name      string      `json:"name"`
	Videos    interface{} `json:"videos"`
}

func GetAllUpcomingStreams() []interface{} {

	API_KEY := os.Getenv("YOUTUBE_API_KEY")

	ctx := context.Background()

	service, err := youtube.NewService(ctx, option.WithAPIKey(API_KEY))

	if err != nil {
		log.Fatalf("Error creating new Youtube client: %v", err)
	}

	results := []interface{}{}

	for _, streamer := range streamers {
		call := service.Search.List([]string{"id", "snippet"}).
			ChannelId(streamer.channelId).
			Type("video").
			EventType("upcoming").
			MaxResults(*maxResults)

		response, err := call.Do()

		if err != nil {
			log.Fatalf("Error making search API call: %v", err)
		}

		videoIds := []string{}

		for _, item := range response.Items {
			videoIds = append(videoIds, item.Id.VideoId)
		}

		videoLists := []interface{}{}

		if len(videoIds) > 0 {
			videoCall := service.Videos.List([]string{"snippet", "liveStreamingDetails"}).Id(videoIds...)

			response, err := videoCall.Do()

			if err != nil {
				log.Fatalf("Error making search API call: %v", err)
			}

			for _, item := range response.Items {
				obj := VideoInfo{
					ID:                 item.Id,
					Title:              item.Snippet.Title,
					ScheduledStartTime: item.LiveStreamingDetails.ScheduledStartTime,
					Thumbnails:         item.Snippet.Thumbnails,
				}
				videoLists = append(videoLists, obj)
			}
		}
		info := StreamerInfo{
			ChannelID: streamer.channelId,
			Name:      streamer.name,
			Videos:    videoLists,
		}
		results = append(results, info)
	}

	return results
}
