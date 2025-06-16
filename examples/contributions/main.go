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

	"cloud.google.com/go/civil"
	"github.com/t-monaghan/altar/application"
)

//nolint:gochecknoglobals
var (
	contributionsChannel chan []int
	once                 sync.Once
	channelInitialized   bool
)

// Handler is receives contributions data from `gh altar contributions` and passes it to Fetcher.
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

const lightGreyRawValue = 251604817    // #EFF2F51
const lightestGreenRawValue = 11333307 // #ACEEBB
const midGreenRawValue = 4899435       // #4AC26B
const highestGreenValue = 1139497      // #116329
const goldValue = 16508676             // #FBE704
const redValue = 16711680              // #FF0000

func contributionGraphsDrawInstruction(allContributions []int) application.ImageAndPosition {
	displayableContributions := allContributions[daysWillFitOnDisplay():]
	busiestDay := slices.Max(displayableContributions)
	transformed := transformRightThenDownToDownThenRight(displayableContributions)

	painted := make([]int, 224)

	for pos, contributionValue := range transformed {
		var colour int

		switch {
		case contributionValue == 0:
			colour = lightGreyRawValue
		case contributionValue < (busiestDay / binCount):
			colour = lightestGreenRawValue
		case contributionValue < (busiestDay/binCount)*2:
			colour = midGreenRawValue
		case contributionValue < busiestDay:
			colour = highestGreenValue
		case contributionValue == busiestDay:
			colour = goldValue
		default:
			colour = redValue

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

func firstWeekOfMonthDrawInstruction() application.ImageAndPosition {
	todaysDate := civil.DateOf(time.Now())
	startOfThisMonth := civil.Date{Year: todaysDate.Year, Month: todaysDate.Month, Day: 1}
	drawing := make([]int, widthOfDisplay)

	for i := range 8 {
		weeksSinceStartOfMonth := todaysDate.DaysSince(startOfThisMonth.AddMonths(-i)) / daysInAWeek
		if weeksSinceStartOfMonth < widthOfDisplay {
			drawing[weeksSinceStartOfMonth] = 2790338 // #2A93C2
		}
	}
	// Contribution grid is in reverse chronological order, we reverse the drawing to match this.
	slices.Reverse(drawing)

	return application.ImageAndPosition{XPos: 0, Ypos: heightOfDisplay, Width: widthOfDisplay, Height: 1, Image: drawing}
}

const binCount = 5

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
