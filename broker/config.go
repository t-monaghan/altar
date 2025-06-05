package broker

// AwtrixConfig defines the configuration Altar can perform on an Awtrix device
//
//nolint:tagliatelle
type AwtrixConfig struct {
	// https://blueforcer.github.io/awtrix3/#/api?id=json-properties-1
	TimeAppEnabled *bool `json:"TIM,omitempty"`
}

// DisableAllDefaultApps configures the broker to diable all default apps on startup.
func DisableAllDefaultApps() func(*AwtrixConfig) {
	return func(cfg *AwtrixConfig) {
		defaultApps := []func(*AwtrixConfig){
			DisableDefaultTimeApp(),
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
