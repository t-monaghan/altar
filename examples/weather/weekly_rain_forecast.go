package weather

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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
		return HourlyForecast{}, false, fmt.Errorf("error creating request for rain forecast app: %w", err)
	}

	q := req.URL.Query()
	q.Add("latitude", "-37.814")
	q.Add("longitude", "144.9633")
	q.Add("hourly", "precipitation_probability")
	q.Add("timezone", "Australia/Sydney")
	req.URL.RawQuery = q.Encode()

	response, err := client.Do(req)
	if err != nil {
		return HourlyForecast{}, false, fmt.Errorf("error performing request against rain forecast app: %w", err)
	}

	defer func() {
		closeErr := response.Body.Close()
		if closeErr != nil {
			err = fmt.Errorf("error closing response body: %w", closeErr)
		}
	}()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return HourlyForecast{}, false, fmt.Errorf("error reading response of rain forecast app: %w", err)
	}

	forecast := &forecastResponse{}
	err = json.Unmarshal(body, forecast)

	if err != nil {
		return HourlyForecast{}, false, fmt.Errorf("failed to unmarshal weather response: %w", err)
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

type precipitationResponse struct {
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Timezone    string  `json:"timezone"`
	CurrentData struct {
		Time          string  `json:"time"`
		Interval      int     `json:"interval"`
		Precipitation float64 `json:"precipitation"`
	} `json:"current"`
	CurrentUnits struct {
		Time          string `json:"time"`
		Interval      string `json:"interval"`
		Precipitation string `json:"precipitation"`
	} `json:"current_units"`
}

// ErrNoPrecipitationData is returned when the weather api returns no precipitation data.
var ErrNoPrecipitationData = errors.New("precipitation data not found in JSON response")

func extractPrecipitation(jsonData string) (float64, error) {
	var weatherData precipitationResponse

	err := json.Unmarshal([]byte(jsonData), &weatherData)
	if err != nil {
		return 0, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if weatherData.CurrentData.Precipitation == 0 && (len(jsonData) == 0 || jsonData == "{}") {
		return 0, ErrNoPrecipitationData
	}

	return weatherData.CurrentData.Precipitation, nil
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
