// Package application provides the ability to create applications for use in altar's broker
//
// The behaviour for an application is defined in it's fetcher
package application

import (
	"fmt"
)

// AppData is Altar's presentation of a custom Awtrix application.
type AppData struct {
	Text string `json:"text"`
}

// Application is Altar's approach of managing the data retrieval and storage required of a custom Awtrix application.
type Application struct {
	Name    string
	fetcher func() (string, error)
	data    AppData
}

// NewApplication Instantiates a new Altar application.
func NewApplication(name string, fetcher func() (string, error)) Application {
	return Application{
		Name:    name,
		fetcher: fetcher,
		data:    AppData{Text: ""},
	}
}

// Fetch uses the application's fetcher to query for new data.
func (a *Application) Fetch() error {
	val, err := a.fetcher()
	if err != nil {
		return fmt.Errorf("failed to fetch for %v: %w", a.Name, err)
	}

	// TODO: only update changed fields (fields will expand to icon etc)
	a.data = AppData{Text: val}

	return nil
}

// GetData returns the application's current data.
func (a *Application) GetData() AppData {
	return a.data
}
