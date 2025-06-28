// Package weather defines some example applications that display weather information and forecasting.
package weather

import (
	"fmt"
	"net/http"
	"time"

	"github.com/t-monaghan/altar/application"
	"github.com/t-monaghan/altar/utils/awtrix"
)

const blueHex = "#3396FF"
const whiteHex = "#FFFFFF"

// Fetcher displays information about precipitation in Melbourne.
func Fetcher(app *application.Application, client *http.Client) error {
	precip, err := currentPrecipitation(client)
	if err != nil {
		return fmt.Errorf("error querying current precipitation: %w", err)
	}

	thirty := 30
	app.Data.ScrollSpeed = &thirty

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

	readableTime := nextRainInWords(nextRain)

	colouredText = append(colouredText, application.TextWithColour{
		Colour: whiteHex,
		Text:   readableTime,
	})

	app.Data.Text = colouredText
	app.Data.Overlay = awtrix.Rain

	return nil
}

func nextRainInWords(nextRain HourlyForecast) string {
	var readableTime string

	timeUntilRain := time.Until(nextRain.Time)

	switch {
	case timeUntilRain < time.Minute:
		readableTime = "1 min"
	case timeUntilRain < time.Hour:
		readableTime = fmt.Sprintf("%.0f mins", timeUntilRain.Minutes())
	case timeUntilRain < 2*time.Hour:
		readableTime = "1 hour"
	case timeUntilRain < 6*time.Hour:
		readableTime = fmt.Sprintf("%.0f hours", timeUntilRain.Hours())
	default:
		readableTime = nextRain.Time.Format("3PM Mon")
	}

	return readableTime
}
