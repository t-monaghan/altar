package broker

import (
	"github.com/t-monaghan/altar/utils/awtrix"
)

// DisableAllDefaultApps configures the broker to diable all default Awtrix apps on startup.
func DisableAllDefaultApps() func(*awtrix.Config) {
	return func(cfg *awtrix.Config) {
		defaultApps := []func(*awtrix.Config){
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

// DisableDefaultTimeApp disables the default time app on the awtrix device on broker startup.
func DisableDefaultTimeApp() func(*awtrix.Config) {
	return func(cfg *awtrix.Config) {
		disable := false
		cfg.TimeAppEnabled = &disable
	}
}

// DisableDefaultWeekdayApp disables the default time app on the awtrix device on broker startup.
func DisableDefaultWeekdayApp() func(*awtrix.Config) {
	return func(cfg *awtrix.Config) {
		disable := false
		cfg.WeekdayAppEnabled = &disable
	}
}

// DisableDefaultDateApp disables the default date app on the awtrix device on broker startup.
func DisableDefaultDateApp() func(*awtrix.Config) {
	return func(cfg *awtrix.Config) {
		disable := false
		cfg.DateAppEnabled = &disable
	}
}

// DisableDefaultHumidityApp disables the default humidity app on the awtrix device on broker startup.
func DisableDefaultHumidityApp() func(*awtrix.Config) {
	return func(cfg *awtrix.Config) {
		disable := false
		cfg.HumidityAppEnabled = &disable
	}
}

// DisableDefaultTempApp disables the default temperature app on the awtrix device on broker startup.
func DisableDefaultTempApp() func(*awtrix.Config) {
	return func(cfg *awtrix.Config) {
		disable := false
		cfg.TempAppEnabled = &disable
	}
}

// DisableDefaultBatteryApp disables the default battery app on the awtrix device on broker startup.
func DisableDefaultBatteryApp() func(*awtrix.Config) {
	return func(cfg *awtrix.Config) {
		disable := false
		cfg.BatteryAppEnabled = &disable
	}
}
