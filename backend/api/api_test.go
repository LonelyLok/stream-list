package api

import (
	"testing"

	"example.com/backend/util"
)

func TestGetAllUpcomingStreams(t *testing.T) {
	util.EnvSetUp()
	GetAllUpcomingStreams()
}
