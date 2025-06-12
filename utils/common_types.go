package utils

// AwtrixConfig defines the configuration Altar can perform on an Awtrix device.
type AwtrixConfig struct {
	// https://blueforcer.github.io/awtrix3/#/api?id=json-properties-1
	TimeAppEnabled     *bool   `json:"TIM,omitempty"`
	WeekdayAppEnabled  *bool   `json:"WD,omitempty"`
	DateAppEnabled     *bool   `json:"DAT,omitempty"`
	HumidityAppEnabled *bool   `json:"HUM,omitempty"`
	TempAppEnabled     *bool   `json:"TEMP,omitempty"`
	BatteryAppEnabled  *bool   `json:"BAT,omitempty"`
	Overlay            Overlay `json:"OVERLAY,omitempty"`
}

// Overlay represents the enumarable options for Awtrix app and global overlays.
type Overlay string

//nolint:revive
const (
	Rain  Overlay = "rain"
	Clear Overlay = "clear"
)
