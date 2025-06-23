package utils

import (
	"net/http"
	"time"

	"github.com/t-monaghan/altar/utils/awtrix"
)

// Routine defines the requirements for an object to be managed by an altar broker.
type Routine interface {
	Fetch(client *http.Client) error
	GetData() any
	ShouldPushToAwtrix() bool
	GetName() string
	GetPollRate() time.Duration
	GetGlobalConfig() awtrix.Config
}
