// An example of Altar's intended usage
//
// To have this run in debug mode with a request logger run `devbox services up`
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"slices"
	"time"

	"github.com/t-monaghan/altar/application"
	"github.com/t-monaghan/altar/broker"
)

var client = http.Client{Timeout: 5 * time.Second}

func helloWorldFetcher(a *application.AppData) error {
	a.Text = "Hello, World!"

	return nil
}

func rainChanceFetcher(a *application.AppData) error {
	// TODO: query if currently raining
	req, err := http.NewRequest(http.MethodGet, "https://api.open-meteo.com/v1/forecast", nil)
	if err != nil {
		return fmt.Errorf("error creating request for rain forecast app: %w", err)
	}

	q := req.URL.Query()
	q.Add("latitude", "-37.814")
	q.Add("longitude", "144.9633")
	q.Add("hourly", "precipitation_probability")
	q.Add("timezone", "Australia/Sydney")
	req.URL.RawQuery = q.Encode()

	response, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error performing request against rain forecast app: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("error reading response of rain forecast app: %w", err)
	}

	forecast := &WeatherResponse{}
	err = json.Unmarshal(body, forecast)

	if err != nil {
		return fmt.Errorf("failed to unmarshal weather response: %w", err)
	}
	hourly := forecast.GetHourlyForecast()

	slices.SortFunc(hourly, func(a, b HourlyForecast) int {
		return int(a.Time.Sub(b.Time))
	})

	nextRain := HourlyForecast{}
	foundRain := false
	for _, hour := range hourly {
		if hour.PrecipitationProbability > 0 {
			nextRain = hour
			foundRain = true
			break
		}
	}

	if foundRain {
		untilNextRain := nextRain.Time.Sub(time.Now())
		if untilNextRain < time.Hour*24 {
			a.Text = fmt.Sprintf("%v%% chance of rain in %v hours", nextRain.PrecipitationProbability, untilNextRain.Round(time.Hour).Hours())
		} else if untilNextRain < time.Hour*24*2 {
			a.Text = fmt.Sprintf("%v%% chance of rain tomorrow", nextRain.PrecipitationProbability)
		} else if untilNextRain < time.Hour*24*7 {
			a.Text = fmt.Sprintf("%v%% chance of rain in %v days", nextRain.PrecipitationProbability, int(untilNextRain.Hours()/24))
		}
	} else {
		a.Text = "sunny week"
	}

	return nil
}

func main() {
	helloWorld := application.NewApplication("Hello World", helloWorldFetcher)
	weatherApp := application.NewApplication("Rain Forecast", rainChanceFetcher)
	appList := []*application.Application{&helloWorld, &weatherApp}
	// TODO: read ip from config (viper?)
	// or allow dynamic address via HTTP request
	broker, err := broker.NewBroker(
		"127.0.0.1",
		appList,
		broker.DisableAllDefaultApps(),
	)
	// TODO: read debug from flag (cobra/viper?)
	broker.Debug = true

	if err != nil {
		slog.Error("error instantiating new broker", "error", err)
		os.Exit(1)
	}

	broker.Start()
}

// TODO: have a server spawn and listen to "/kill" to allow rolling a new broker build
