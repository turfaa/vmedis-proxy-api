package proxy

import (
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/turfaa/vmedis-proxy-api/vmedis/client"
)

// Run runs the proxy server.
func Run(vmedisClient *client.Client, sessionRefreshInterval time.Duration) {
	log.Printf("Starting proxy server to %s with refresh interval %d\n", vmedisClient.BaseUrl, sessionRefreshInterval)
	log.Println("Proxy server started")

	stop := vmedisClient.AutoRefreshSessionId(sessionRefreshInterval)
	defer stop()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, os.Kill)

	<-done

	log.Println("Proxy server stopped")
}
