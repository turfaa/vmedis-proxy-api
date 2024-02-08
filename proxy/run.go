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
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/vmedis"
)

// Config is the proxy server configuration.
type Config struct {
	VmedisClient *vmedis.Client
	DB           *gorm.DB
	RedisClient  *redis.Client
	KafkaWriter  *kafka.Writer
}

// Run runs the proxy server.
func Run(config Config) {
	log.Printf("Starting proxy server to %s", config.VmedisClient.BaseUrl)

	apiServer := NewApiServer(config.VmedisClient, config.DB, config.RedisClient, config.KafkaWriter)
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

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGINT)

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
