package main

import (
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
)

func requestLoggerHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("error reading body", "error", err)
		http.Error(w, "error reading request body", http.StatusInternalServerError)

		return
	}
	slog.Info("request received", "method", r.Method, "path", r.URL.Path, "body", body, "queryParamaters", r.URL.RawQuery)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("request logged successfully"))
}

func main() {
	port := "8080"

	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	http.HandleFunc("/", requestLoggerHandler)

	addr := ":" + port
	slog.Info("starting request logger server on " + addr)

	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
