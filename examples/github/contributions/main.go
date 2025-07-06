// Package contributions is intended for use with `gh-altar` to display the GitHub contributions on an Awtrix device via
// an altar broker.
package contributions

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/t-monaghan/altar/application"
)

//nolint:gochecknoglobals
var (
	contributionsChannel chan []int
	once                 sync.Once
	channelInitialized   bool
)

// Handler receives contributions data from `gh altar contributions` and passes it to Fetcher.
func Handler(rsp http.ResponseWriter, req *http.Request) {
	if !channelInitialized {
		initChannel()
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		slog.Error("github contributions handler failed to read body", "error", err)
		rsp.WriteHeader(http.StatusBadRequest)

		return
	}

	var contributions []int

	err = json.Unmarshal(body, &contributions)
	if err != nil {
		slog.Error("github contributions handler failed to unmarshal request", "error", err)
		rsp.WriteHeader(http.StatusBadRequest)

		return
	}

	select {
	case contributionsChannel <- contributions:
	default:
		slog.Warn("github contributions channel is full, dropping message")
	}

	rsp.WriteHeader(http.StatusOK)
}

// Fetcher receives data from the handler and prepares it to be posted by altar's broker.
func Fetcher(app *application.Application, _ *http.Client) error {
	if !channelInitialized {
		initChannel()
	}

	var rawCount []int
	select {
	case rawCount = <-contributionsChannel:
		slog.Debug("contributions fetcher received contributions count", "length-of-count", len(rawCount))
	default:
		app.PushOnNextCall = false

		return nil
	}

	app.PushOnNextCall = true

	graph := contributionGraphsDrawInstruction(rawCount)

	firstWeekOfMonth := firstWeekOfMonthDrawInstruction()

	app.Data.Draw = &[]application.DrawInstructions{
		{Bitmap: &graph},
		{Bitmap: &firstWeekOfMonth},
	}

	return nil
}

const black = 0x000000
const darkestGreen = 0x1D2F21
const darkGreen = 0x254727
const green = 0x307732
const brightGreen = 0x3AA63C
const dimWhite = 0x888888
const red = 0xFF0000

func contributionGraphsDrawInstruction(allContributions []int) application.ImageAndPosition {
	indexBackTo := len(allContributions) - daysWillFitOnDisplay()
	displayableContributions := allContributions[indexBackTo:]
	busiestDay := slices.Max(displayableContributions)
	transformed := transformRightThenDownToDownThenRight(displayableContributions)

	painted := make([]int, widthOfDisplay*daysInAWeek)

	for pos, contributionValue := range transformed {
		var colour int

		switch {
		case contributionValue == 0:
			colour = black
		case contributionValue <= busiestDay/8:
			colour = darkestGreen
		case contributionValue <= busiestDay/3:
			colour = darkGreen
		case contributionValue <= busiestDay/3*2:
			colour = green
		case contributionValue < busiestDay:
			colour = brightGreen
		case contributionValue == busiestDay:
			colour = dimWhite
		default:
			colour = red

			slog.Error("github contribution count did not bin correctly", "value", contributionValue, "max", busiestDay)
		}

		painted[pos] = colour
	}

	return application.ImageAndPosition{
		XPos:   0,
		Ypos:   0,
		Width:  widthOfDisplay,
		Height: heightOfDisplay - 1,
		Image:  painted,
	}
}

const blue = 0x2A93C2
const hoursInADay = 24

func firstWeekOfMonthDrawInstruction() application.ImageAndPosition {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	startOfThisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	drawing := make([]int, widthOfDisplay)

	for i := range drawing {
		drawing[i] = black
	}

	for i := range 8 {
		pastMonthStart := startOfThisMonth.AddDate(0, -i, 0)
		daysSince := int(today.Sub(pastMonthStart).Hours() / hoursInADay)

		weeksSinceStartOfMonth := daysSince / daysInAWeek
		if weeksSinceStartOfMonth < widthOfDisplay {
			drawing[weeksSinceStartOfMonth] = blue
		}
	}
	// contribution grid is in reverse chronological order, we reverse the drawing to match this.
	slices.Reverse(drawing)

	return application.ImageAndPosition{
		XPos:   0,
		Ypos:   heightOfDisplay - 1,
		Width:  widthOfDisplay,
		Height: 1,
		Image:  drawing}
}

const widthOfDisplay = 32

const heightOfDisplay = 8

const daysInAWeek = 7

func transformRightThenDownToDownThenRight(contributions []int) []int {
	vertWeeks := make([]int, widthOfDisplay*daysInAWeek)

	weeksHandled := 0

	for week := range slices.Chunk(contributions, daysInAWeek) {
		for i, day := range week {
			vertWeeks[i*widthOfDisplay+weeksHandled] = day
		}

		weeksHandled++
	}

	return vertWeeks
}

func daysWillFitOnDisplay() int {
	daysLeftInWeek := daysInAWeek - int(time.Now().Weekday()) - 1 // subtract one as github includes today's contributions
	awtrixDayDisplayCount := widthOfDisplay * daysInAWeek
	displayDays := awtrixDayDisplayCount - daysLeftInWeek

	return displayDays
}

const channelBufferSize = 5

func initChannel() {
	once.Do(func() {
		contributionsChannel = make(chan []int, channelBufferSize)
		channelInitialized = true
	})
}

// Reset clears the state of the channel used to communicate between the api handler and the altar fetcher.
func Reset() {
	if channelInitialized {
		for len(contributionsChannel) > 0 {
			<-contributionsChannel
		}
	}
}
