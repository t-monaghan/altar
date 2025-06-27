package weather

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"time"
)

// HourlyForecast represents a single hourly forecast entry with parsed time.
type HourlyForecast struct {
	Time                     time.Time
	PrecipitationProbability int // percentage
}

// ErrEmptyResponse describes when the weather api returns an empty body.
var ErrEmptyResponse = errors.New("did not receive a response body from weather api")

func weeklyRainForecast(client *http.Client) (HourlyForecast, bool, error) {
	req, err := http.NewRequestWithContext(context.Background(),
		http.MethodGet, "https://api.open-meteo.com/v1/forecast", nil)
	if err != nil {
		return HourlyForecast{}, false, fmt.Errorf("error creating request for weekly rain forecast: %w", err)
	}

	query := req.URL.Query()
	query.Add("latitude", os.Getenv("LATITUDE"))
	query.Add("longitude", os.Getenv("LONGITUDE"))
	query.Add("timezone", os.Getenv("WEATHER_TIMEZONE"))
	query.Add("hourly", "precipitation_probability")

	req.URL.RawQuery = query.Encode()

	response, err := client.Do(req)
	if err != nil {
		return HourlyForecast{}, false, fmt.Errorf("error requesting weekly rain forecast: %w", err)
	}

	defer func() {
		closeErr := response.Body.Close()
		if closeErr != nil {
			err = fmt.Errorf("error closing response body: %w", closeErr)
		}
	}()

	forecast, err := readForecastResponse(response)
	if err != nil {
		return HourlyForecast{}, false, fmt.Errorf("error reading forecast response: %w", err)
	}

	if len(forecast.Hourly.PrecipitationProbability) == 0 {
		return HourlyForecast{}, false, ErrEmptyResponse
	}

	hourly := forecast.getHourlyForecast()

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

	return nextRain, foundRain, nil
}

type forecastResponse struct {
	Latitude             float64                `json:"latitude"`
	Longitude            float64                `json:"longitude"`
	GenerationTimeMs     float64                `json:"generationtime_ms"`
	UTCOffsetSeconds     int                    `json:"utc_offset_seconds"`
	Timezone             string                 `json:"timezone"`
	TimezoneAbbreviation string                 `json:"timezone_abbreviation"`
	Elevation            float64                `json:"elevation"`
	HourlyUnits          hourlyUnits            `json:"hourly_units"`
	Hourly               hourlyForecastResponse `json:"hourly"`
}

func readForecastResponse(response *http.Response) (*forecastResponse, error) {
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response of rain forecast: %w", err)
	}

	forecast := &forecastResponse{}

	err = json.Unmarshal(body, forecast)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal rain forecast: %w", err)
	}

	return forecast, nil
}

type hourlyUnits struct {
	Time                     string `json:"time"`
	PrecipitationProbability string `json:"precipitation_probability"`
}

func (wr *forecastResponse) getHourlyForecast() []HourlyForecast {
	result := make([]HourlyForecast, len(wr.Hourly.Time))

	for i := range wr.Hourly.Time {
		t, _ := time.Parse("2006-01-02T15:04", wr.Hourly.Time[i])

		result[i] = HourlyForecast{
			Time:                     t,
			PrecipitationProbability: wr.Hourly.PrecipitationProbability[i],
		}
	}

	return result
}

// hourlyForecastResponse is defined as we transform the response struct
// containing lists into a list containing the structs.
type hourlyForecastResponse struct {
	Time                     []string `json:"time"`                      // ISO8601 time strings
	PrecipitationProbability []int    `json:"precipitation_probability"` // percentage values
}
