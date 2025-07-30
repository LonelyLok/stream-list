package api

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

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
	{
		"UC6T7TJZbW6nO-qsc5coo8Pg",
		"Dooby3D",
		"https://static.wikia.nocookie.net/virtualyoutuber/images/1/10/Dooby3d.png",
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

type Thumbnail struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type Thumbnails struct {
	Medium *Thumbnail `json:"medium"`
}

// Stream holds the fields we care about for each upcoming stream.
type VideoInfoV2 struct {
	ID                   string     `json:"id"`
	Title                string     `json:"title"`
	ScheduledStartTime   string     `json:"scheduledStartTime"`
	Thumbnails           Thumbnails `json:"thumbnails"`
	LiveBroadcastContent string     `json:"liveBroadcastContent"`
	ChannelID            string     `json:"channelId"`
}

type StreamerInfo struct {
	ChannelID string        `json:"channelId"`
	Name      string        `json:"name"`
	IconURL   string        `json:"iconURL"`
	Videos    []interface{} `json:"videos"`
}

type StreamerInfoV2 struct {
	ChannelID string        `json:"channelId"`
	Name      string        `json:"name"`
	IconURL   string        `json:"iconURL"`
	Videos    []VideoInfoV2 `json:"videos"`
}

func contains(slice []string, element string) bool {
	for _, v := range slice {
		if v == element {
			return true
		}
	}
	return false
}

var httpClient = &http.Client{
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
	},
	Timeout: 15 * time.Second,
}

func TestFun() {
	fmt.Println("Hello, World!")
}

func fetchStartTime(videoID string) (time.Time, error) {
	url := "https://www.youtube.com/watch?v=" + videoID

	// 1) build a new GET request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return time.Time{}, err
	}

	// 2) force no-cache
	req.Header.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Expires", "0")

	// 3) execute it
	resp, err := httpClient.Do(req)
	if err != nil {
		return time.Time{}, err
	}
	defer resp.Body.Close()

	// 4) read the body as before
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return time.Time{}, err
	}

	// regex extract ytInitialPlayerResponse = { … };
	re := regexp.MustCompile(`(?s)ytInitialPlayerResponse\s*=\s*(\{.*?\});`)
	m := re.FindStringSubmatch(string(body))
	if len(m) < 2 {
		return time.Time{}, fmt.Errorf("player JSON not found")
	}

	var pr map[string]interface{}
	if err := json.Unmarshal([]byte(m[1]), &pr); err != nil {
		return time.Time{}, err
	}

	// dig into microformat → playerMicroformatRenderer
	mf, ok := pr["microformat"].(map[string]interface{})
	if !ok {
		return time.Time{}, fmt.Errorf("microformat not found")
	}
	pmr, ok := mf["playerMicroformatRenderer"].(map[string]interface{})
	if !ok {
		return time.Time{}, fmt.Errorf("playerMicroformatRenderer not found")
	}

	// pull the liveBroadcastDetails.startTimestamp
	lbd, ok := pmr["liveBroadcastDetails"].(map[string]interface{})
	if !ok {
		return time.Time{}, fmt.Errorf("liveBroadcastDetails not found")
	}
	ts, ok := lbd["startTimestamp"].(string)
	if !ok {
		return time.Time{}, fmt.Errorf("startTimestamp not a string")
	}

	// parse the RFC3339 timestamp
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

func scrapeStreams(channelID string) ([]VideoInfoV2, error) {
	url := fmt.Sprintf("https://www.youtube.com/channel/%s/streams", channelID)
	results := []VideoInfoV2{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	// tell any intermediate caches (and the server) not to return a cached response
	req.Header.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Expires", "0")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch page: %w", err)
	}
	defer resp.Body.Close()

	htmlBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return results, fmt.Errorf("read body: %w", err)
	}
	html := string(htmlBytes)

	// 1) Extract the ytInitialData JSON blob
	re := regexp.MustCompile(`(?s)ytInitialData\s*=\s*(\{.*?\});`)
	matches := re.FindStringSubmatch(html)
	if len(matches) < 2 {
		return results, fmt.Errorf("ytInitialData not found")
	}

	var initialData map[string]interface{}
	if err := json.Unmarshal([]byte(matches[1]), &initialData); err != nil {
		return results, fmt.Errorf("parse JSON: %w", err)
	}

	// fmt.Printf("%#v\n", initialData)

	// 2) Drill into contents.twoColumnBrowseResultsRenderer.tabs
	contents, _ := initialData["contents"].(map[string]interface{})
	browse, _ := contents["twoColumnBrowseResultsRenderer"].(map[string]interface{})
	tabs, _ := browse["tabs"].([]interface{})

	// 3) Locate the Streams tab
	var streamTab map[string]interface{}
	for _, t := range tabs {
		tr, ok := t.(map[string]interface{})["tabRenderer"].(map[string]interface{})
		if !ok {
			continue
		}
		title, _ := tr["title"].(string)
		if title == "Live" || tr["selected"] == true {
			streamTab, _ = tr["content"].(map[string]interface{})
			break
		}
	}
	if streamTab == nil {
		return results, fmt.Errorf("streams tab not found")
	}

	// 4) Extract the grid of items
	grid, _ := streamTab["richGridRenderer"].(map[string]interface{})
	items, _ := grid["contents"].([]interface{})

	for _, c := range items {
		itemMap, _ := c.(map[string]interface{})["richItemRenderer"].(map[string]interface{})
		content, _ := itemMap["content"].(map[string]interface{})
		vr, ok := content["videoRenderer"].(map[string]interface{})
		if !ok {
			continue
		}

		id := vr["videoId"].(string)
		titleRuns := vr["title"].(map[string]interface{})["runs"].([]interface{})
		title := titleRuns[0].(map[string]interface{})["text"].(string)

		thumbsRaw, _ := vr["thumbnail"].(map[string]interface{})["thumbnails"].([]interface{})

		var t Thumbnails

		for _, raw := range thumbsRaw {

			m, ok := raw.(map[string]interface{})
			if !ok {
				continue
			}

			url, ok1 := m["url"].(string)
			if !ok1 {
				continue
			}
			mediumURL := strings.Replace(url, "hqdefault.jpg", "mqdefault.jpg", 1)
			thumb := &Thumbnail{
				URL:    mediumURL,
				Width:  320,
				Height: 180,
			}
			t.Medium = thumb
			break
		}

		isLive := false
		if overlays, ok := vr["thumbnailOverlays"].([]interface{}); ok {
			for _, o := range overlays {
				if m, ok := o.(map[string]interface{})["thumbnailOverlayTimeStatusRenderer"].(map[string]interface{}); ok {
					if m["style"] == "LIVE" {
						isLive = true
						break
					}
				}
			}
		}

		up, isUpcoming := vr["upcomingEventData"].(map[string]interface{})

		if !isUpcoming && !isLive {
			continue
		}
		var scheduledTime time.Time
		if isUpcoming {
			tsStr, _ := up["startTime"].(string)
			tsInt, _ := strconv.ParseInt(tsStr, 10, 64)
			scheduled := time.Unix(tsInt, 0)
			scheduledTime = scheduled
		} else if isLive {
			startTime, err := fetchStartTime(id)
			if err != nil {
				fmt.Println("Error fetching start time:", err)
			}
			scheduledTime = startTime
		}

		if scheduledTime.IsZero() {
			fmt.Println("No scheduled time found for video:", id)
			continue
		}

		var status string
		if isUpcoming {
			status = "upcoming"
		} else {
			status = "live"
		}

		results = append(results, VideoInfoV2{
			ID:                   id,
			Title:                title,
			ScheduledStartTime:   scheduledTime.UTC().Format(time.RFC3339),
			LiveBroadcastContent: status,
			ChannelID:            channelID,
			Thumbnails:           t,
		})
	}
	return results, nil
}

func GetAllUpcomingStreamsByScraping() map[string]StreamerInfoV2 {
	var wg sync.WaitGroup
	var mu sync.Mutex
	results := make(map[string]StreamerInfoV2, len(streamers))

	for _, streamer := range streamers {
		wg.Add(1)
		go func(s Streamer) {
			defer wg.Done()

			videos, err := scrapeStreams(s.channelId)
			if err != nil {
				fmt.Println("Error scraping streams for", s.name, ":", err)
				return
			}

			info := StreamerInfoV2{
				ChannelID: s.channelId,
				Name:      s.name,
				IconURL:   s.iconURL,
				Videos:    videos,
			}

			mu.Lock()
			results[s.channelId] = info
			mu.Unlock()
		}(streamer)
	}

	wg.Wait()
	return results
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
