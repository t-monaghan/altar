package utils

import (
	"net/http"
	"time"
)

type AltarHandler interface {
	ShouldPushToAwtrix() bool
	Fetch(client *http.Client) error
	GetData() any
	SetPushOnNextCall(bool)
	GetName() string
	GetPollRate() time.Duration
	GetGlobalConfig() AwtrixConfig
}
