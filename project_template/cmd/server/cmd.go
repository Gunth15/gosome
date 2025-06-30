package main

import (
	"log/slog"
	"net/http"
)

func main() {
	PORT := ":8080"

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	})

	server := http.Server{
		Addr:    PORT,
		Handler: mux,
	}

	slog.Info("Starting Server", "port", PORT)
	if err := server.ListenAndServe(); err != nil {
		slog.Error(err.Error())
	}
}
