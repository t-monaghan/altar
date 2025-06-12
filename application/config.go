package application

import "github.com/t-monaghan/altar/utils"

// DisableAllDefaultApps configures the broker to diable all default apps on startup.
func DisableAllDefaultApps() func(*utils.AwtrixConfig) {
	return func(cfg *utils.AwtrixConfig) {
		defaultApps := []func(*utils.AwtrixConfig){
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
func DisableDefaultTimeApp() func(*utils.AwtrixConfig) {
	return func(cfg *utils.AwtrixConfig) {
		disable := false
		cfg.TimeAppEnabled = &disable
	}
}

// DisableDefaultWeekdayApp disables the default time app on the awtrix display on broker startup.
func DisableDefaultWeekdayApp() func(*utils.AwtrixConfig) {
	return func(cfg *utils.AwtrixConfig) {
		disable := false
		cfg.WeekdayAppEnabled = &disable
	}
}

// DisableDefaultDateApp disables the default date app on the awtrix display on broker startup.
func DisableDefaultDateApp() func(*utils.AwtrixConfig) {
	return func(cfg *utils.AwtrixConfig) {
		disable := false
		cfg.DateAppEnabled = &disable
	}
}

// DisableDefaultHumidityApp disables the default humidity app on the awtrix display on broker startup.
func DisableDefaultHumidityApp() func(*utils.AwtrixConfig) {
	return func(cfg *utils.AwtrixConfig) {
		disable := false
		cfg.HumidityAppEnabled = &disable
	}
}

// DisableDefaultTempApp disables the default temperature app on the awtrix display on broker startup.
func DisableDefaultTempApp() func(*utils.AwtrixConfig) {
	return func(cfg *utils.AwtrixConfig) {
		disable := false
		cfg.TempAppEnabled = &disable
	}
}

// DisableDefaultBatteryApp disables the default battery app on the awtrix display on broker startup.
func DisableDefaultBatteryApp() func(*utils.AwtrixConfig) {
	return func(cfg *utils.AwtrixConfig) {
		disable := false
		cfg.BatteryAppEnabled = &disable
	}
}
