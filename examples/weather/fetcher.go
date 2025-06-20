// Package weather defines some example applications that display weather information and forecasting.
package weather

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/t-monaghan/altar/application"
	"github.com/t-monaghan/altar/utils/awtrix"
)

// Fetcher displays information about precipitation in Melbourne.
//
//nolint:funlen
func Fetcher(app *application.Application, client *http.Client) error {
	precip, err := currentPrecipitation(client)
	if err != nil {
		return fmt.Errorf("error querying current precipitation: %w", err)
	}

	if precip > 0 {
		app.Data.Text = fmt.Sprintf("Raining: %.0fmm", precip)
		app.Data.Overlay = awtrix.Rain
		app.GlobalConfig.Overlay = awtrix.Rain

		return nil
	}
	// removes any previous application of the rain effect
	app.Data.Overlay = ""
	app.GlobalConfig.Overlay = awtrix.Clear

	nextRain, foundRain, err := weeklyRainForecast(client)
	if err != nil {
		return err
	}

	if !foundRain {
		app.Data.Text = "sunny week"

		return nil
	}

	colouredText := []application.TextWithColour{}
	rainChanceString := strconv.Itoa(nextRain.PrecipitationProbability) + "% "

	const blueHex = "#3396FF"

	const whiteHex = "#FFFFFF"

	colouredText = append(colouredText, application.TextWithColour{
		Colour: blueHex,
		Text:   rainChanceString})

	timeUntilRain := time.Until(nextRain.Time)

	var readableTime string

	switch {
	case timeUntilRain < time.Minute:
		readableTime = "in 1 min"
	case timeUntilRain < time.Hour:
		readableTime = fmt.Sprintf("in %.0f mins", timeUntilRain.Minutes())
	case timeUntilRain < 2*time.Hour:
		readableTime = "in 1 hour"
	case timeUntilRain < 6*time.Hour:
		readableTime = fmt.Sprintf("in %.0f hours", timeUntilRain.Hours())
	default:
		readableTime = nextRain.Time.Format("3PM Mon")
	}

	colouredText = append(colouredText, application.TextWithColour{
		Colour: whiteHex,
		Text:   readableTime,
	})

	app.Data.Text = colouredText

	return nil
}
