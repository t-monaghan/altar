package main

import "time"

// WeatherResponse represents the complete response from the weather API
type WeatherResponse struct {
	Latitude             float64     `json:"latitude"`
	Longitude            float64     `json:"longitude"`
	GenerationTimeMs     float64     `json:"generationtime_ms"`
	UTCOffsetSeconds     int         `json:"utc_offset_seconds"`
	Timezone             string      `json:"timezone"`
	TimezoneAbbreviation string      `json:"timezone_abbreviation"`
	Elevation            float64     `json:"elevation"`
	HourlyUnits          HourlyUnits `json:"hourly_units"`
	Hourly               HourlyData  `json:"hourly"`
}

// HourlyUnits represents the units used for hourly data
type HourlyUnits struct {
	Time                     string `json:"time"`
	PrecipitationProbability string `json:"precipitation_probability"`
}

// HourlyData represents the hourly forecast data
type HourlyData struct {
	Time                     []string `json:"time"`                      // ISO8601 time strings
	PrecipitationProbability []int    `json:"precipitation_probability"` // percentage values
}

// GetHourlyForecast returns hourly forecast data as paired time and probability values
// for easier consumption
func (wr *WeatherResponse) GetHourlyForecast() []HourlyForecast {
	result := make([]HourlyForecast, len(wr.Hourly.Time))

	for i := 0; i < len(wr.Hourly.Time); i++ {
		// Parse ISO8601 time string to time.Time
		t, _ := time.Parse("2006-01-02T15:04", wr.Hourly.Time[i])

		result[i] = HourlyForecast{
			Time:                     t,
			PrecipitationProbability: wr.Hourly.PrecipitationProbability[i],
		}
	}

	return result
}

// HourlyForecast represents a single hourly forecast entry with parsed time
type HourlyForecast struct {
	Time                     time.Time
	PrecipitationProbability int // percentage
}
