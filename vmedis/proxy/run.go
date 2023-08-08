package proxy

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/turfaa/vmedis-proxy-api/vmedis/client"
)

// Run runs the proxy server.
func Run(vmedisClient *client.Client, sessionRefreshInterval time.Duration) {
	log.Println("Checking if session id is valid")
	if err := vmedisClient.RefreshSessionId(); err != nil {
		log.Fatalf("Session id check failed: %s\n", err)
	}

	log.Printf("Starting proxy server to %s with refresh interval %d\n", vmedisClient.BaseUrl, sessionRefreshInterval)

	apiServer := ApiServer{Client: vmedisClient}
	engine := apiServer.GinEngine()

	httpServer := http.Server{
		Addr:    ":8080",
		Handler: engine,
	}

	serverError := make(chan error, 1)
	go func() {
		serverError <- httpServer.ListenAndServe()
	}()

	log.Println("Proxy server started")

	stop := vmedisClient.AutoRefreshSessionId(sessionRefreshInterval)
	defer stop()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, os.Kill)

	select {
	case err := <-serverError:
		log.Fatalf("Error starting proxy server: %s\n", err)

	case <-done:
		log.Println("Stopping proxy server")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(ctx); err != nil {
			log.Fatalf("Error stopping proxy server: %s\n", err)
		}
	}

	log.Println("Proxy server stopped")
}
