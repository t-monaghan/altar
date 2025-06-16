// Package application provides the ability to create applications for use in altar's broker
//
// The behaviour for an application is defined in it's fetcher
package application

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/t-monaghan/altar/utils"
)

// AppData is Altar's presentation of a custom Awtrix application.
//
//nolint:lll
type AppData struct {
	// includes all fields from docs linked, except for "draw" and "effect settings"
	// https://github.com/Blueforcer/awtrix3/blob/main/docs/api.md#json-properties
	Text         string              `json:"text,omitempty"`
	TextCase     *int                `json:"textCase,omitempty"`
	TopText      *bool               `json:"topText,omitempty"`
	TextOffset   *int                `json:"textOffset,omitempty"`
	Center       *bool               `json:"center,omitempty"`
	Color        []int               `json:"color,omitempty"`    // RGB color values [R,G,B]
	Gradient     [][]int             `json:"gradient,omitempty"` // Array of RGB colors [[R,G,B], [R,G,B]]
	BlinkText    *int                `json:"blinkText,omitempty"`
	FadeText     *int                `json:"fadeText,omitempty"`
	Background   []int               `json:"background,omitempty"` // RGB color values [R,G,B]
	Rainbow      *bool               `json:"rainbow,omitempty"`
	Icon         string              `json:"icon,omitempty"`
	PushIcon     *int                `json:"pushIcon,omitempty"`
	Repeat       *int                `json:"repeat,omitempty"`
	Duration     *int                `json:"duration,omitempty"`
	Bar          []int               `json:"bar,omitempty"`
	Line         []int               `json:"line,omitempty"`
	Autoscale    *bool               `json:"autoscale,omitempty"`
	BarBC        []int               `json:"barBC,omitempty"` // RGB color values [R,G,B]
	Progress     *int                `json:"progress,omitempty"`
	ProgressC    []int               `json:"progressC,omitempty"`  // RGB color values [R,G,B]
	ProgressBC   []int               `json:"progressBC,omitempty"` // RGB color values [R,G,B]
	Pos          *int                `json:"pos,omitempty"`
	Lifetime     *int                `json:"lifetime,omitempty"`
	LifetimeMode *int                `json:"lifetimeMode,omitempty"` // TODO: check nil = shows the app, 0 = deletes the app, 1 = marks it as staled with a red rectangle around the app.
	NoScroll     *bool               `json:"noScroll,omitempty"`
	ScrollSpeed  *int                `json:"scrollSpeed,omitempty"`
	Effect       string              `json:"effect,omitempty"`
	Overlay      utils.Overlay       `json:"overlay,omitempty"`
	Draw         *[]DrawInstructions `json:"draw,omitempty"`
}

// DrawInstructions is the container for the different image instructions possible in awtrix.
type DrawInstructions struct {
	Bitmap *ImageAndPosition `json:"db,omitempty"`
}

// ImageAndPosition defines how to draw an image based on pixel colours in their raw numeric value.
type ImageAndPosition struct {
	XPos   int
	Ypos   int
	Width  int
	Height int
	Image  []int
}

// MarshalJSON is used here as ImageAndPosition has a curious format defined by awtrix.
//
//nolint:wrapcheck //MarshalJSON is being overloaded, so we have no error to wrap around the returned error.
func (f *ImageAndPosition) MarshalJSON() ([]byte, error) {
	return json.Marshal([]any{
		f.XPos,
		f.Ypos,
		f.Width,
		f.Height,
		f.Image,
	})
}

// Application is Altar's approach of managing the data retrieval and storage required of a custom Awtrix application.
type Application struct {
	Name           string
	fetcher        func(*Application, *http.Client) error
	Data           AppData
	GlobalConfig   utils.AwtrixConfig
	PollRate       time.Duration
	lastPolled     time.Time
	PushOnNextCall bool
	HTTPClient     *http.Client
}

// NewApplication Instantiates a new altar application.
func NewApplication(name string, fetcher func(*Application, *http.Client) error) Application {
	return Application{
		Name:           name,
		fetcher:        fetcher,
		Data:           AppData{},
		GlobalConfig:   utils.AwtrixConfig{},
		PollRate:       utils.DefaultPollRate,
		PushOnNextCall: false,
	}
}

// SetPollRateByRateLimit is a helper function that sets the application's poll rate
// when given the count of requests per duration.
func (a *Application) SetPollRateByRateLimit(requests uint32, duration time.Duration) {
	a.PollRate = duration / time.Duration(requests)
}

// ShouldFetch defines whether an application should be fetched again according to it's poll rate.
func (a *Application) ShouldFetch() bool {
	return time.Since(a.lastPolled) > a.PollRate
}

// ShouldPushToAwtrix defines whether an application should push it's data to Awtrix.
func (a *Application) ShouldPushToAwtrix() bool {
	return a.PushOnNextCall
}

// GetName returns the name of the app.
func (a *Application) GetName() string {
	return a.Name
}

// GetPollRate returns the app's poll rate.
func (a *Application) GetPollRate() time.Duration {
	return a.PollRate
}

// Fetch uses the application's fetcher to query for new data.
func (a *Application) Fetch(client *http.Client) error {
	if !a.ShouldFetch() {
		slog.Debug("skipping app fetch", "app", a.Name,
			"seconds-since-last-fetch", time.Since(a.lastPolled).Seconds(), "poll-rate-seconds", a.PollRate.Seconds())

		return nil
	}

	slog.Debug("fetching for app", "app", a.Name,
		"seconds-since-last-fetch", time.Since(a.lastPolled).Seconds(), "poll-rate-seconds", a.PollRate.Seconds())

	a.lastPolled = time.Now()
	a.PushOnNextCall = true

	return a.fetcher(a, client)
}

// GetData returns the application's current data.
func (a *Application) GetData() any {
	return a.Data
}

// GetGlobalConfig returns the global awtrix config this application has set.
func (a *Application) GetGlobalConfig() utils.AwtrixConfig {
	return a.GlobalConfig
}
