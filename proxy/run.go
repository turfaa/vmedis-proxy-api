package proxy

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/vmedis"
)

// Config is the proxy server configuration.
type Config struct {
	VmedisClient           *vmedis.Client
	DB                     *gorm.DB
	RedisClient            *redis.Client
	SessionRefreshInterval time.Duration
}

// Run runs the proxy server.
func Run(config Config) {
	log.Println("Checking if session ids are valid")

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	if err := config.VmedisClient.RefreshSessionIds(ctx); err != nil {
		log.Fatalf("Session id check failed: %s\n", err)
	}

	log.Printf("Starting proxy server to %s with refresh interval %d\n", config.VmedisClient.BaseUrl, config.SessionRefreshInterval)

	apiServer := NewApiServer(config.VmedisClient, config.DB, config.RedisClient)
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

	stop := config.VmedisClient.AutoRefreshSessionIds(config.SessionRefreshInterval)
	defer stop()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

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
