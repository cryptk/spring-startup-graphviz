package grapher

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/goccy/go-graphviz/cgraph"
)

type StartupDuration struct {
	time.Duration
}

type StartupResponse struct {
	SpringBootVersion string          `json:"springBootVersion"`
	Timeline          StartupTimeline `json:"timeline"`
}

type StartupTimeline struct {
	StartTime string         `json:"startTime"`
	Events    []StartupEvent `json:"events"`
}

type StartupEvent struct {
	StartTime   string          `json:"startTime"`
	EndTime     string          `json:"endTime"`
	Duration    StartupDuration `json:"duration"`
	StartupStep StartupStep     `json:"startupStep"`
}

type StartupStep struct {
	Name     string       `json:"name"`
	ID       int          `json:"id"`
	ParentID int          `json:"parentID"`
	Tags     []StartupTag `json:"tags"`
}

type StartupTag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type graphNode struct {
	node         *cgraph.Node
	event        StartupEvent
	shouldRender bool
	parent       *graphNode
}

// This converts a Period Time string in the format "PT...S" into a time.Duration
func (sd *StartupDuration) UnmarshalJSON(data []byte) error {

	var durationString string
	if err := json.Unmarshal(data, &durationString); err != nil {
		return err
	}

	durationString = strings.Replace(durationString, "PT", "", 1)
	durationString = strings.Replace(durationString, "S", "s", 1)
	duration, err := time.ParseDuration(durationString)
	if err != nil {
		return err
	}

	sd.Duration = duration
	return nil
}
