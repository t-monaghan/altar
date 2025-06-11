package weather

import (
	"fmt"
	"net/http"

	"github.com/t-monaghan/altar/application"
)

func RainChanceFetcher(a *application.Application, c *http.Client) error {
	// TODO: query if currently raining
	nextRain, foundRain, err := weeklyRainForecast(c)
	if err != nil {
		return err
	}

	if !foundRain {
		a.Data.Text = "sunny week"
		return nil
	}

	a.Data.Text = fmt.Sprintf("%v%% %v", nextRain.PrecipitationProbability, nextRain.Time.Format("3PM Mon"))

	return nil
}
