package utils

import (
	"net/http"
	"time"
)

// AltarHandler defines what is required for a broker to manage a handler.
type AltarHandler interface {
	Fetch(client *http.Client) error
	// TODO remove for handling to be done in fetcher with switching on type
	GetData() any
	ShouldPushToAwtrix() bool
	GetName() string
	GetPollRate() time.Duration
	GetGlobalConfig() AwtrixConfig
}
