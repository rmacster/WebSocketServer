package main

import (
	"log"
	"time"
)

func main() {
	log.Println("Starting WebServer...")

	wsStart()

	for {
		time.Sleep(time.Second)
	}
}
