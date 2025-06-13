package weather

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func currentPrecipitation(client *http.Client) (float64, error) {
	req, err := http.NewRequestWithContext(context.Background(),
		http.MethodGet, "https://api.open-meteo.com/v1/forecast", nil)
	if err != nil {
		return 0, fmt.Errorf("error creating request for current precipitation: %w", err)
	}

	query := req.URL.Query()
	query.Add("latitude", os.Getenv("LATITUDE"))
	query.Add("longitude", os.Getenv("LONGITUDE"))
	query.Add("current", "precipitation")

	zone, _ := time.Now().Zone()
	slog.Info(zone)
	query.Add("timezone", zone)
	req.URL.RawQuery = query.Encode()

	response, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("error requesting current precipitation: %w", err)
	}

	defer func() {
		closeErr := response.Body.Close()
		if closeErr != nil {
			err = fmt.Errorf("error closing response body: %w", closeErr)
		}
	}()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return 0, fmt.Errorf("error reading response of current precipitation: %w", err)
	}

	precip, err := extractPrecipitation(body)
	if err != nil {
		return 0, fmt.Errorf("error reading precipitation data: %w", err)
	}

	return precip, nil
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

// extractPrecipitation takes the response from a current precipitation api call
// and extracts the current rainfall in mm.
func extractPrecipitation(jsonData []byte) (float64, error) {
	var weatherData precipitationResponse

	err := json.Unmarshal(jsonData, &weatherData)
	if err != nil {
		return 0, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if weatherData.CurrentData.Precipitation == 0 && (len(jsonData) == 0 || string(jsonData) == "{}") {
		return 0, ErrNoPrecipitationData
	}

	return weatherData.CurrentData.Precipitation, nil
}
