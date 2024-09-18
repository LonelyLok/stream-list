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

type Streamer struct {
	channelId string
	name      string
	iconURL   string
}

var streamers = []Streamer{
	{
		"UCL_qhgtOy0dy1Agp8vkySQg",
		"Mori Calliope",
		"https://yt3.googleusercontent.com/8B_T08sx8R7XVi5Mwx_l9sjQm5FGWGspeujSvVDvd80Zyr-3VvVTRGVLOnBrqNRxZp6ZeXAV=s120-c-k-c0x00ffffff-no-rj",
	},
	{
		"UC8rcEBzJSleTkf_-agPM20g",
		"IRyS",
		"https://yt3.googleusercontent.com/cDSMiVy3Xa49Ci_YyouVNzfCwVXKRYmOeywWQ_UFKzvAp6tvyeMtXMyzWzQ2u8ft4EENsJKt7A=s120-c-k-c0x00ffffff-no-rj",
	},
	{
		"UCHsx4Hqa-1ORjQTh9TYDhww",
		"Takanashi Kiara",
		"https://yt3.googleusercontent.com/w7TKJYU7zmamFmf-WxfahCo_K7Bg2__Pk-CCBNnbewMG-77OZLqJO9MLvDAmH9nEkZH8OkWgSQ=s120-c-k-c0x00ffffff-no-rj",
	},
	{
		"UCgmPnx-EEeOrZSg5Tiw7ZRQ",
		"Hakos Baelz",
		"https://yt3.googleusercontent.com/9FFCT3cu9FxyLJUUFovpPI7WRj0I7_KuApwkEaLsD0NHVVL2OPTtGn8Qga5YFbeC_47-MoEXrA=s176-c-k-c0x00ffffff-no-rj-mo",
	},
	{
		"UCgnfPPb9JI3e9A4cXHnWbyg",
		"Shiori Novella",
		"https://yt3.ggpht.com/ZlovVsPyh8NgS37S4dfONiCBySiboGPbT9cYuirb8JM3JhSnqlpJk-8SQUEA7jPfqXpMvjaa=s176-c-k-c0x00ffffff-no-rj-mo",
	},
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
	ID                   string                    `json:"id"`
	Title                string                    `json:"title"`
	ScheduledStartTime   string                    `json:"scheduledStartTime"`
	Thumbnails           *youtube.ThumbnailDetails `json:"thumbnails"`
	LiveBroadcastContent string                    `json:"liveBroadcastContent"`
	ChannelID            string                    `json:"channelId"`
}

type StreamerInfo struct {
	ChannelID string        `json:"channelId"`
	Name      string        `json:"name"`
	IconURL   string        `json:"iconURL"`
	Videos    []interface{} `json:"videos"`
}

func GetAllUpcomingStreams() interface{} {

	API_KEY := os.Getenv("YOUTUBE_API_KEY")

	ctx := context.Background()

	service, err := youtube.NewService(ctx, option.WithAPIKey(API_KEY))

	if err != nil {
		log.Fatalf("Error creating new Youtube client: %v", err)
	}

	resultsMap := map[string]StreamerInfo{}

	videoIds := []string{}

	videoLists := []interface{}{}

	for _, streamer := range streamers {
		call := service.Search.List([]string{"id", "snippet"}).
			ChannelId(streamer.channelId).
			Type("video").
			EventType("upcoming").
			MaxResults(*maxResults)

		response, err := call.Do()

		if err != nil {
			log.Printf("Error making search API call: %v", err)
			continue
		}

		for _, item := range response.Items {
			videoIds = append(videoIds, item.Id.VideoId)
		}

		info := StreamerInfo{
			ChannelID: streamer.channelId,
			Name:      streamer.name,
			IconURL:   streamer.iconURL,
			Videos:    videoLists,
		}
		resultsMap[streamer.channelId] = info
	}

	if len(videoIds) > 0 {
		videoCall := service.Videos.List([]string{"snippet", "liveStreamingDetails"}).Id(videoIds...)

		response, err := videoCall.Do()

		if err != nil {
			log.Fatalf("Error making search API call: %v", err)
		}

		for _, item := range response.Items {
			obj := VideoInfo{
				ID:                   item.Id,
				Title:                item.Snippet.Title,
				ScheduledStartTime:   item.LiveStreamingDetails.ScheduledStartTime,
				Thumbnails:           item.Snippet.Thumbnails,
				LiveBroadcastContent: item.Snippet.LiveBroadcastContent,
				ChannelID:            item.Snippet.ChannelId,
			}
			info := resultsMap[obj.ChannelID]
			info.Videos = append(info.Videos, obj)
			resultsMap[obj.ChannelID] = info
		}
	}

	return resultsMap
}
