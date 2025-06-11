// Package weather defines some example applications that display weather information and forecasting.
package weather

import (
	"fmt"
	"net/http"

	"github.com/t-monaghan/altar/application"
)

// RainChanceFetcher displays information about precipitation in Melbourne.
func RainChanceFetcher(app *application.Application, c *http.Client) error {
	// TODO: query if currently raining
	nextRain, foundRain, err := weeklyRainForecast(c)
	if err != nil {
		return err
	}

	if !foundRain {
		app.Data.Text = "sunny week"

		return nil
	}

	app.Data.Text = fmt.Sprintf("%v%% %v", nextRain.PrecipitationProbability, nextRain.Time.Format("3PM Mon"))

	return nil
}
