// Package awtrix provides utilities for awtrix related logic.
package awtrix

// Config defines the configuration options for an Awtrix device.
type Config struct {
	// https://blueforcer.github.io/awtrix3/#/api?id=json-properties-1
	TimeAppEnabled     *bool   `json:"TIM,omitempty"`
	WeekdayAppEnabled  *bool   `json:"WD,omitempty"`
	DateAppEnabled     *bool   `json:"DAT,omitempty"`
	HumidityAppEnabled *bool   `json:"HUM,omitempty"`
	TempAppEnabled     *bool   `json:"TEMP,omitempty"`
	BatteryAppEnabled  *bool   `json:"BAT,omitempty"`
	Overlay            Overlay `json:"OVERLAY,omitempty"`
	TransitionEffect   *int    `json:"TEFF,omitempty"`
}

// Overlay represents the set of available overlays for Awtrix devices.
type Overlay string

//nolint:revive//reason: Overlay is a transparent enum and does not require documentation.
const (
	// Clear will remove any previously set overlays.
	Clear   Overlay = "clear"
	Drizzle Overlay = "drizzle"
	Rain    Overlay = "rain"
	Storm   Overlay = "storm"
	Thunder Overlay = "thunder"
	Snow    Overlay = "snow"
	Frost   Overlay = "frost"
)
