// Package weather defines some example applications that display weather information and forecasting.
package weather

import (
	"fmt"
	"net/http"

	"github.com/t-monaghan/altar/application"
)

// RainChanceFetcher displays information about precipitation in Melbourne.
func RainChanceFetcher(app *application.Application, client *http.Client) (application.AwtrixConfig, error) {
	precip, err := currentPrecipitation(client)
	if err != nil {
		return application.AwtrixConfig{}, fmt.Errorf("error querying current precipitation: %w", err)
	}

	if precip > 0 {
		app.Data.Text = fmt.Sprintf("Raining: %.0fmm", precip)
		app.Data.Overlay = application.Rain

		return application.AwtrixConfig{Overlay: application.Rain}, nil
	}
	// removes any previous application of the rain effect
	app.Data.Overlay = ""

	nextRain, foundRain, err := weeklyRainForecast(client)
	if err != nil {
		return application.AwtrixConfig{}, err
	}

	if !foundRain {
		app.Data.Text = "sunny week"

		return application.AwtrixConfig{Overlay: application.Clear}, nil
	}

	app.Data.Text = fmt.Sprintf("%v%% %v", nextRain.PrecipitationProbability, nextRain.Time.Format("3PM Mon"))

	return application.AwtrixConfig{Overlay: application.Clear}, nil
}
