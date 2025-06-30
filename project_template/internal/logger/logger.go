// Package logger configures the logger
package logger

import (
	"log/slog"
	"net/http"
	"os"
)

func Init() {
	logger := slog.NewJSONHandler(os.Stdout, nil)
	slog.SetDefault(slog.New(logger))
}

func Request(req http.Request) {
	g := slog.Group("request", "method", req.Method, "url", req.URL.String(), "body", req.Body)
	slog.Info("Recieved new request", g)
}

func Response(msg string, status int, body string) {
	g := slog.Group("response", slog.Int("status", status), "body", body)
	if status >= 400 && status <= 599 {
		slog.Error(msg, g)
	} else {
		slog.Info(msg, g)
	}
}
