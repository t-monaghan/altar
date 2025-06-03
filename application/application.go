package application

import (
	"fmt"
)

type AppData struct {
	Text string `json:"text"`
}

type Application struct {
	Name    string
	fetcher func() (string, error)
	Data    AppData
}

func NewApplication(name string, fetcher func() (string, error)) Application {
	return Application{
		Name:    name,
		fetcher: fetcher,
	}
}

func (a *Application) Fetch() error {
	val, err := a.fetcher()
	if err != nil {
		return fmt.Errorf("failed to fetch for %v: %w", a.Name, err)
	}
	a.Data = AppData{Text: val}
	return nil
}
