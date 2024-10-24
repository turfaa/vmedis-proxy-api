package main

import (
	"log"
	"time"

	"github.com/klauspost/lctime"
	"github.com/turfaa/vmedis-proxy-api/cmd"
)

func main() {
	setupTime()

	cmd.Execute()
}

func setupTime() {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		log.Fatalf("LoadLocation: %s", err)
	}
	time.Local = loc

	lctime.SetLocale("id_ID")
}
