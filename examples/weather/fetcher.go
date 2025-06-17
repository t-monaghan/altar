// Package weather defines some example applications that display weather information and forecasting.
package weather

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/t-monaghan/altar/application"
	"github.com/t-monaghan/altar/utils"
)

// Fetcher displays information about precipitation in Melbourne.
func Fetcher(app *application.Application, client *http.Client) error {
	precip, err := currentPrecipitation(client)
	if err != nil {
		return fmt.Errorf("error querying current precipitation: %w", err)
	}

	if precip > 0 {
		app.Data.Text = fmt.Sprintf("Raining: %.0fmm", precip)
		app.Data.Overlay = utils.Rain
		app.GlobalConfig.Overlay = utils.Rain

		return nil
	}
	// removes any previous application of the rain effect
	app.Data.Overlay = ""
	app.GlobalConfig.Overlay = utils.Clear

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
	colouredText = append(colouredText, application.TextWithColour{
		Colour: whiteHex,
		Text:   nextRain.Time.Format("3PM Mon")})

	app.Data.Text = colouredText

	return nil
}
