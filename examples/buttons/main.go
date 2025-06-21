// Package buttons provides a handler and configuration for listening to awtrix button presses.
package buttons

import (
	"io"
	"log/slog"
	"net/http"
)

// Handler listens for Awtrix's button callback requests.
func Handler(rsp http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		slog.Error("github checks handler failed to read body", "error", err)
		rsp.WriteHeader(http.StatusBadRequest)

		return
	}

	slog.Info("awtrix button callback received", "body", string(body))
	rsp.WriteHeader(http.StatusOK)
}
