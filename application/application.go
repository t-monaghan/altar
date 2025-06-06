// Package application provides the ability to create applications for use in altar's broker
//
// The behaviour for an application is defined in it's fetcher
package application

import (
	"log/slog"
	"time"
)

// AppData is Altar's presentation of a custom Awtrix application.
//
//nolint:tagliatelle
type AppData struct {
	// includes all fields from docs linked, except for "draw" and "effect settings"
	// https://github.com/Blueforcer/awtrix3/blob/main/docs/api.md#json-properties
	Text         string   `json:"text,omitempty"`
	TextCase     *int     `json:"textCase,omitempty"`
	TopText      *bool    `json:"topText,omitempty"`
	TextOffset   *int     `json:"textOffset,omitempty"`
	Center       *bool    `json:"center,omitempty"`
	Color        []int    `json:"color,omitempty"`    // RGB color values [R,G,B]
	Gradient     [][]int  `json:"gradient,omitempty"` // Array of RGB colors [[R,G,B], [R,G,B]]
	BlinkText    *int     `json:"blinkText,omitempty"`
	FadeText     *int     `json:"fadeText,omitempty"`
	Background   []int    `json:"background,omitempty"` // RGB color values [R,G,B]
	Rainbow      *bool    `json:"rainbow,omitempty"`
	Icon         string   `json:"icon,omitempty"`
	PushIcon     *int     `json:"pushIcon,omitempty"`
	Repeat       *int     `json:"repeat,omitempty"`
	Duration     *int     `json:"duration,omitempty"`
	Hold         *bool    `json:"hold,omitempty"`
	Sound        string   `json:"sound,omitempty"`
	Rtttl        string   `json:"rtttl,omitempty"`
	LoopSound    *bool    `json:"loopSound,omitempty"`
	Bar          []int    `json:"bar,omitempty"`
	Line         []int    `json:"line,omitempty"`
	Autoscale    *bool    `json:"autoscale,omitempty"`
	BarBC        []int    `json:"barBC,omitempty"` // RGB color values [R,G,B]
	Progress     *int     `json:"progress,omitempty"`
	ProgressC    []int    `json:"progressC,omitempty"`  // RGB color values [R,G,B]
	ProgressBC   []int    `json:"progressBC,omitempty"` // RGB color values [R,G,B]
	Pos          *int     `json:"pos,omitempty"`
	Lifetime     *int     `json:"lifetime,omitempty"`
	LifetimeMode *int     `json:"lifetimeMode,omitempty"`
	Stack        *bool    `json:"stack,omitempty"`
	Wakeup       *bool    `json:"wakeup,omitempty"`
	NoScroll     *bool    `json:"noScroll,omitempty"`
	Clients      []string `json:"clients,omitempty"`
	ScrollSpeed  *int     `json:"scrollSpeed,omitempty"`
	Effect       string   `json:"effect,omitempty"`
	Save         *bool    `json:"save,omitempty"`
	Overlay      string   `json:"overlay,omitempty"`
}

// Application is Altar's approach of managing the data retrieval and storage required of a custom Awtrix application.
type Application struct {
	Name           string
	fetcher        func(*Application) error
	Data           AppData
	PollRate       time.Duration
	lastPolled     time.Time
	PushOnNextCall bool
}

const defaultPollRate = time.Second * 10

// NewApplication Instantiates a new Altar application.
func NewApplication(name string, fetcher func(*Application) error) Application {
	return Application{
		Name:           name,
		fetcher:        fetcher,
		Data:           AppData{},
		PollRate:       defaultPollRate,
		PushOnNextCall: false,
	}
}

// SetPollRateByRateLimit is a helper function that sets the application's poll rate
// when given the count of requests per duration.
func (a *Application) SetPollRateByRateLimit(requests int, perDuration time.Duration) {
	a.PollRate = time.Duration(requests / int(perDuration))
}

// ShouldFetch defines whether an application should be fetched again according to it's poll rate.
func (a *Application) ShouldFetch() bool {
	return time.Since(a.lastPolled) > a.PollRate
}

// ShouldPushToAwtrix defines whether an application should push it's data to Awtrix.
func (a *Application) ShouldPushToAwtrix() bool {
	return a.PushOnNextCall
}

// Fetch uses the application's fetcher to query for new data.
func (a *Application) Fetch() error {
	if !a.ShouldFetch() {
		slog.Debug("skipping app fetch", "app", a.Name,
			"seconds-since-last-fetch", time.Since(a.lastPolled).Seconds(), "poll-rate-seconds", a.PollRate.Seconds())

		return nil
	}

	slog.Debug("fetching for app", "app", a.Name,
		"seconds-since-last-fetch", time.Since(a.lastPolled).Seconds(), "poll-rate-seconds", a.PollRate.Seconds())

	a.lastPolled = time.Now()
	a.PushOnNextCall = true

	return a.fetcher(a)
}

// GetData returns the application's current data.
func (a *Application) GetData() AppData {
	return a.Data
}
