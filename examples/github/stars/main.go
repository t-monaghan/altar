// Package stars provides an example extension of altar's broker server.
package stars

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sync"

	"github.com/t-monaghan/altar/notifier"
)

//nolint:gochecknoglobals
var (
	starrerChannel     chan string
	once               sync.Once
	channelInitialized bool
)

const channelBufferSize = 5

func initChannel() {
	once.Do(func() {
		starrerChannel = make(chan string, channelBufferSize)
		channelInitialized = true
	})
}

// Fetcher receives data from the handler and prepares it to be posted by altar's broker.
func Fetcher(ntfr *notifier.Notifier, _ *http.Client) error {
	if !channelInitialized {
		initChannel()
	}

	var starrer string
	select {
	case starrer = <-starrerChannel:
		slog.Debug("contributions fetcher received contributions count", "starrer", starrer)
	default:
		ntfr.PushOnNextCall = false

		return nil
	}

	three := 3
	seventyFive := 75
	t := true
	ntfr.PushOnNextCall = true
	ntfr.Data.Text = starrer
	ntfr.Data.Repeat = &three
	ntfr.Data.Rainbow = &t
	ntfr.Data.ScrollSpeed = &seventyFive

	return nil
}

type webhookPayload struct {
	Sender sender `json:"sender"`
	Action string `json:"action"`
}

type sender struct {
	Login string `json:"login"`
}

// Handler receives data from the gh-altar tool and passes it onto Fetcher.
func Handler(rsp http.ResponseWriter, req *http.Request) {
	if !channelInitialized {
		initChannel()
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		slog.Error("github checks handler failed to read body", "error", err)
		rsp.WriteHeader(http.StatusBadRequest)

		return
	}

	sig := req.Header.Get("X-Hub-Signature-256")
	if sig == "" {
		slog.Error("did not find signature in webhook request")
		rsp.WriteHeader(http.StatusBadRequest)

		return
	}

	secret := os.Getenv("STARGAZER_WEBHOOK_SECRET")
	if secret == "" {
		slog.Error("could not find stargazer secret")
		rsp.WriteHeader(http.StatusServiceUnavailable)

		return
	}

	statusCode := validateSignature(sig, secret, body)
	if statusCode != 0 {
		rsp.WriteHeader(statusCode)

		return
	}

	slog.Debug("webhook with matching signature received")

	var webhookPayload webhookPayload
	if err := json.Unmarshal(body, &webhookPayload); err != nil {
		slog.Error("failed to unmarshal payload JSON", "error", err, "body", body)
		rsp.WriteHeader(http.StatusBadRequest)

		return
	}

	if webhookPayload.Action != "created" {
		slog.Debug("non-created action received from github starred webhook", "action", webhookPayload.Action)
		rsp.WriteHeader(http.StatusOK)

		return
	}

	select {
	case starrerChannel <- webhookPayload.Sender.Login:
	default:
		slog.Warn("github checks channel is full, dropping message")
	}

	rsp.WriteHeader(http.StatusOK)
}

func validateSignature(signature string, secret string, body []byte) int {
	hash := hmac.New(sha256.New, []byte(secret))

	_, err := hash.Write(body)
	if err != nil {
		slog.Error("failed to write github payload to hmac")

		return http.StatusInternalServerError
	}

	expected := hash.Sum(nil)
	expectedSig := "sha256=" + hex.EncodeToString(expected)

	if !hmac.Equal([]byte(expectedSig), []byte(signature)) {
		slog.Warn("webhook received with mismatching signature")

		return http.StatusUnauthorized
	}

	return 0
}

// Reset clears the state of the channel used to communicate between the api handler and the altar fetcher.
func Reset() {
	if channelInitialized {
		for len(starrerChannel) > 0 {
			<-starrerChannel
		}
	}
}
