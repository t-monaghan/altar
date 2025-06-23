// Package notifier provides functionality to write notifiers for altar brokers.
package notifier

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/t-monaghan/altar/utils"
	"github.com/t-monaghan/altar/utils/awtrix"
)

// Notifier is altar's approach of managing the data retrieval and storage required of a custom Awtrix notifier.
type Notifier struct {
	Name           string
	fetcher        func(*Notifier, *http.Client) error
	Data           *NotificationData
	GlobalConfig   awtrix.Config
	PollRate       time.Duration
	HTTPClient     *http.Client
	PushOnNextCall bool
	lastPolled     time.Time
}

// NewNotifier instantiates a new altar notification routine.
func NewNotifier(name string, fetcher func(*Notifier, *http.Client) error) Notifier {
	return Notifier{
		Name:           name,
		Data:           &NotificationData{},
		GlobalConfig:   awtrix.Config{},
		PollRate:       utils.DefaultPollRate,
		PushOnNextCall: false,
		fetcher:        fetcher,
	}
}

// Fetch controls the fetching for a notifier.
func (n *Notifier) Fetch(client *http.Client) error {
	if !n.ShouldFetch() {
		slog.Debug("skipping notifier fetch", "notifier", n.Name,
			"seconds-since-last-fetch", time.Since(n.lastPolled).Seconds(), "poll-rate-seconds", n.PollRate.Seconds())

		return nil
	}

	slog.Debug("fetching for notifier", "notifier", n.Name,
		"seconds-since-last-fetch", time.Since(n.lastPolled).Seconds(), "poll-rate-seconds", n.PollRate.Seconds())

	n.lastPolled = time.Now()

	return n.fetcher(n, client)
}

// NotificationData is altar's presentation of a custom Awtrix notification.
type NotificationData struct {
	Text        string         `json:"text,omitempty"`
	TextCase    *int           `json:"textCase,omitempty"`
	TopText     *bool          `json:"topText,omitempty"`
	TextOffset  *int           `json:"textOffset,omitempty"`
	Center      *bool          `json:"center,omitempty"`
	Color       []int          `json:"color,omitempty"`    // RGB color values [R,G,B]
	Gradient    [][]int        `json:"gradient,omitempty"` // Array of RGB colors [[R,G,B], [R,G,B]]
	BlinkText   *int           `json:"blinkText,omitempty"`
	FadeText    *int           `json:"fadeText,omitempty"`
	Background  []int          `json:"background,omitempty"` // RGB color values [R,G,B]
	Rainbow     *bool          `json:"rainbow,omitempty"`
	Icon        string         `json:"icon,omitempty"`
	PushIcon    *int           `json:"pushIcon,omitempty"`
	Repeat      *int           `json:"repeat,omitempty"`
	Duration    *int           `json:"duration,omitempty"`
	Hold        *bool          `json:"hold,omitempty"`
	Sound       string         `json:"sound,omitempty"`
	Rtttl       string         `json:"rtttl,omitempty"`
	LoopSound   *bool          `json:"loopSound,omitempty"`
	Bar         []int          `json:"bar,omitempty"`
	Line        []int          `json:"line,omitempty"`
	Autoscale   *bool          `json:"autoscale,omitempty"`
	BarBC       []int          `json:"barBC,omitempty"` // RGB color values [R,G,B]
	Progress    *int           `json:"progress,omitempty"`
	ProgressC   []int          `json:"progressC,omitempty"`  // RGB color values [R,G,B]
	ProgressBC  []int          `json:"progressBC,omitempty"` // RGB color values [R,G,B]
	Stack       *bool          `json:"stack,omitempty"`
	Wakeup      *bool          `json:"wakeup,omitempty"`
	NoScroll    *bool          `json:"noScroll,omitempty"`
	Clients     []string       `json:"clients,omitempty"`
	ScrollSpeed *int           `json:"scrollSpeed,omitempty"`
	Effect      string         `json:"effect,omitempty"`
	Overlay     awtrix.Overlay `json:"overlay,omitempty"`
}

// GetName returns the notifier's name.
func (n *Notifier) GetName() string {
	return n.Name
}

// GetPollRate returns the notifier's poll rate.
func (n *Notifier) GetPollRate() time.Duration {
	return n.PollRate
}

// GetData returns the notifiers data, this is the payload sent to the awtrix device.
func (n *Notifier) GetData() any {
	return n.Data
}

// GetGlobalConfig returns the global config this notifier wishes to manipulate.
func (n *Notifier) GetGlobalConfig() awtrix.Config {
	return n.GlobalConfig
}

// ShouldFetch signals whether this notifier should have it's fetch method run.
func (n *Notifier) ShouldFetch() bool {
	return time.Since(n.lastPolled) > n.PollRate
}

// ShouldPushToAwtrix signals whether a broker should push this notifier's data to the awtrix device.
func (n *Notifier) ShouldPushToAwtrix() bool {
	return n.PushOnNextCall
}

// SetPollRateByRateLimit is a helper function that sets the notifiers's poll rate
// when given the count of requests per duration.
func (n *Notifier) SetPollRateByRateLimit(requests uint32, duration time.Duration) {
	n.PollRate = duration / time.Duration(requests)
}
