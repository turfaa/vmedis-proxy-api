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

	"github.com/turfaa/vmedis-proxy-api/auth"
	"github.com/turfaa/vmedis-proxy-api/drug"
	"github.com/turfaa/vmedis-proxy-api/procurement"
	"github.com/turfaa/vmedis-proxy-api/sale"
	"github.com/turfaa/vmedis-proxy-api/stockopname"
)

// Config is the proxy server configuration.
type Config struct {
	DB          *gorm.DB
	RedisClient *redis.Client
	AuthService *auth.Service

	AuthHandler        *auth.ApiHandler
	DrugHandler        *drug.ApiHandler
	SaleHandler        *sale.ApiHandler
	ProcurementHandler *procurement.ApiHandler
	StockOpnameHandler *stockopname.ApiHandler
}

// Run runs the proxy server.
func Run(config Config) {
	log.Println("Starting proxy server")

	apiServer := NewApiServer(
		config.DB,
		config.RedisClient,
		config.AuthService,
		config.AuthHandler,
		config.DrugHandler,
		config.SaleHandler,
		config.ProcurementHandler,
		config.StockOpnameHandler,
	)

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
