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
	isHolo    bool
}

var streamers = []Streamer{
	{
		"UCL_qhgtOy0dy1Agp8vkySQg",
		"Mori Calliope",
		"https://static.wikia.nocookie.net/virtualyoutuber/images/c/cd/Mori_Calliope_-_Icon.png",
		true,
	},
	{
		"UC8rcEBzJSleTkf_-agPM20g",
		"IRyS",
		"https://static.wikia.nocookie.net/virtualyoutuber/images/f/ff/IRyS_-_Icon.png",
		true,
	},
	{
		"UCHsx4Hqa-1ORjQTh9TYDhww",
		"Takanashi Kiara",
		"https://static.wikia.nocookie.net/virtualyoutuber/images/9/9a/Takanashi_Kiara_-_Icon.png",
		true,
	},
	{
		"UCgmPnx-EEeOrZSg5Tiw7ZRQ",
		"Hakos Baelz",
		"https://static.wikia.nocookie.net/virtualyoutuber/images/0/0a/Hakos_Baelz_-_Icon.png",
		true,
	},
	{
		"UCgnfPPb9JI3e9A4cXHnWbyg",
		"Shiori Novella",
		"https://static.wikia.nocookie.net/virtualyoutuber/images/5/53/Shiori_Novella_-_Icon.png",
		true,
	},
	{
		"UCDHABijvPBnJm7F-KlNME3w",
		"Gigi Murin",
		"https://static.wikia.nocookie.net/virtualyoutuber/images/2/2d/Gigi_Murin_-_Icon.png",
		true,
	},
	{
		"UCIfAvpeIWGHb0duCkMkmm2Q",
		"Nimi Nightmare",
		"https://static.wikia.nocookie.net/virtualyoutuber/images/f/f3/Nimi_Nightmare_Portrait.jpg",
		false,
	},
}

func Map[T, U any](slice []T, f func(T) U) []U {
	result := make([]U, len(slice))
	for i, item := range slice {
		result[i] = f(item)
	}
	return result
}

func Filter[T any](slice []T, f func(T) bool) []T {
	filtered := make([]T, 0, len(slice))
	for _, item := range slice {
		if f(item) {
			filtered = append(filtered, item)
		}
	}
	return filtered
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

func contains(slice []string, element string) bool {
	for _, v := range slice {
		if v == element {
			return true
		}
	}
	return false
}

func TestFun() {
	fmt.Println("Hello, World!")
}

func GetAllUpcomingStreams() interface{} {
	resultsMap := map[string]StreamerInfo{}

	API_KEY := os.Getenv("YOUTUBE_API_KEY")

	ctx := context.Background()

	service, err := youtube.NewService(ctx, option.WithAPIKey(API_KEY))

	if err != nil {
		log.Fatalf("Error creating new Youtube client: %v", err)
	}

	videoIds := []string{}

	videoLists := []interface{}{}

	validChannelIds := Map(streamers, func(streamer Streamer) string { return streamer.channelId })

	notHoloOnlyStreamers := Filter(streamers, func(s Streamer) bool {
		return !s.isHolo
	})

	for _, streamer := range streamers {
		info := StreamerInfo{
			ChannelID: streamer.channelId,
			Name:      streamer.name,
			IconURL:   streamer.iconURL,
			Videos:    videoLists,
		}
		resultsMap[streamer.channelId] = info
	}

	for _, eventType := range []string{"upcoming", "live"} {
		call := service.Search.List([]string{"id", "snippet"}).Q("HoloLive EN").Type("video").EventType(eventType).MaxResults(*maxResults)
		response, err := call.Do()
		if err != nil {
			log.Printf("Error making search API call: %v", err)
			return resultsMap
		}
		for _, item := range response.Items {
			if !(contains(validChannelIds, item.Snippet.ChannelId)) {
				continue
			}
			if contains(videoIds, item.Id.VideoId) {
				continue
			}
			videoIds = append(videoIds, item.Id.VideoId)
		}
	}

	for _, streamer := range notHoloOnlyStreamers {
		for _, eventType := range []string{"upcoming", "live"} {
			call := service.Search.List([]string{"id", "snippet"}).
				ChannelId(streamer.channelId).
				Type("video").
				EventType(eventType).
				MaxResults(*maxResults)

			response, err := call.Do()

			if err != nil {
				log.Printf("Error making search API call: %v", err)
				continue
			}

			for _, item := range response.Items {
				if contains(videoIds, item.Id.VideoId) {
					continue
				}
				videoIds = append(videoIds, item.Id.VideoId)
			}
		}
	}

	if len(videoIds) > 0 {
		videoCall := service.Videos.List([]string{"snippet", "liveStreamingDetails"}).Id(videoIds...)

		response, err := videoCall.Do()

		if err != nil {
			log.Printf("Error making Video List API call: %v", err)
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
