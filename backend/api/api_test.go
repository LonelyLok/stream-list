package api

import (
	"testing"

	"example.com/backend/util"
)

func TestGetAllUpcomingStreams(t *testing.T) {
	util.EnvSetUp()
	TestFun()
}

func TestScrapeStreams(t *testing.T) {
	videos, err := scrapeStreams("UCgnfPPb9JI3e9A4cXHnWbyg") // Mori Calliope
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("found %d streams", len(videos))

	for _, v := range videos {
		t.Logf("%s | %s | %s | %s", v.ID, v.LiveBroadcastContent, v.ScheduledStartTime, v.Title)
	}
}
