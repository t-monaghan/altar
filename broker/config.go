package broker

// AwtrixConfig defines the configuration Altar can perform on an Awtrix device
//
//nolint:tagliatelle
type AwtrixConfig struct {
	// https://blueforcer.github.io/awtrix3/#/api?id=json-properties-1
	TimeAppEnabled     *bool `json:"TIM,omitempty"`
	WeekdayAppEnabled  *bool `json:"WD,omitempty"`
	DateAppEnabled     *bool `json:"DAT,omitempty"`
	HumidityAppEnabled *bool `json:"HUM,omitempty"`
	TempAppEnabled     *bool `json:"TEMP,omitempty"`
	BatteryAppEnabled  *bool `json:"BAT,omitempty"`
}

// DisableAllDefaultApps configures the broker to diable all default apps on startup.
func DisableAllDefaultApps() func(*AwtrixConfig) {
	return func(cfg *AwtrixConfig) {
		defaultApps := []func(*AwtrixConfig){
			DisableDefaultTimeApp(),
			DisableDefaultWeekdayApp(),
			DisableDefaultDateApp(),
			DisableDefaultHumidityApp(),
			DisableDefaultTempApp(),
			DisableDefaultBatteryApp(),
		}
		for _, fn := range defaultApps {
			fn(cfg)
		}
	}
}

// DisableDefaultTimeApp disables the default time app on the awtrix display on broker startup.
func DisableDefaultTimeApp() func(*AwtrixConfig) {
	return func(cfg *AwtrixConfig) {
		disable := false
		cfg.TimeAppEnabled = &disable
	}
}

// DisableDefaultWeekdayApp disables the default time app on the awtrix display on broker startup.
func DisableDefaultWeekdayApp() func(*AwtrixConfig) {
	return func(cfg *AwtrixConfig) {
		disable := false
		cfg.WeekdayAppEnabled = &disable
	}
}

// DisableDefaultDateApp disables the default date app on the awtrix display on broker startup.
func DisableDefaultDateApp() func(*AwtrixConfig) {
	return func(cfg *AwtrixConfig) {
		disable := false
		cfg.DateAppEnabled = &disable
	}
}

// DisableDefaultHumidityApp disables the default humidity app on the awtrix display on broker startup.
func DisableDefaultHumidityApp() func(*AwtrixConfig) {
	return func(cfg *AwtrixConfig) {
		disable := false
		cfg.HumidityAppEnabled = &disable
	}
}

// DisableDefaultTempApp disables the default temperature app on the awtrix display on broker startup.
func DisableDefaultTempApp() func(*AwtrixConfig) {
	return func(cfg *AwtrixConfig) {
		disable := false
		cfg.TempAppEnabled = &disable
	}
}

// DisableDefaultBatteryApp disables the default battery app on the awtrix display on broker startup.
func DisableDefaultBatteryApp() func(*AwtrixConfig) {
	return func(cfg *AwtrixConfig) {
		disable := false
		cfg.BatteryAppEnabled = &disable
	}
}

