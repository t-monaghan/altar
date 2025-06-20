package utils

import (
	"net/http"
	"time"

	"github.com/t-monaghan/altar/utils/awtrix"
)

// AltarHandler defines what is required for a broker to manage a handler.
type AltarHandler interface {
	Fetch(client *http.Client) error
	GetData() any
	ShouldPushToAwtrix() bool
	GetName() string
	GetPollRate() time.Duration
	GetGlobalConfig() awtrix.Config
}
