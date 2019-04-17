package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		env := make(map[string]string)
		for _, v := range os.Environ() {
			parts := strings.SplitN(v, "=", 2)
			env[parts[0]] = parts[1]
		}

		b, err := json.Marshal(env)
		if err != nil {
			fmt.Fprintln(w, "error: %s", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, string(b))
	})

	server := &http.Server{
		Addr:    "0.0.0.0:" + port,
		Handler: mux,
	}
	serverCh := make(chan struct{})
	go func() {
		log.Printf("server is listening on %s\n", port)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("server exited with: %s", err)
		}
		close(serverCh)
	}()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	<-signalCh

	log.Printf("received interrupt, shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("failed to shutdown server: %s", err)
	}
}
