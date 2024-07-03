package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Moonlight-Companies/goresp/resp"
)

func main() {
	redisAddr := flag.String("redis", "bus:6379", "Redis server address")
	channelsFlag := flag.String("channels", "*", "Comma-separated list of channels to subscribe to")
	flag.Parse()

	channels := strings.Split(*channelsFlag, ",")

	reconn := resp.NewReconnecting(*redisAddr)

	for _, channel := range channels {
		channel = strings.TrimSpace(channel)
		if strings.Contains(channel, "*") {
			reconn.PSubscribe(channel)
			fmt.Printf("PSubscribed to channel: %s\n", channel)
		} else {
			reconn.Subscribe(channel)
			fmt.Printf("Subscribed to channel: %s\n", channel)
		}
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	fmt.Printf("Connected to Redis at %s\n", *redisAddr)
	fmt.Println("Waiting for messages. Press Ctrl+C to exit.")

	go func() {
		for msg := range reconn.Producer {
			temp, err := msg.IntoMap()
			if err != nil {
				log.Printf("Error parsing message: %v", err)
				continue
			}
			fmt.Printf("Channel: %s, Pattern: %s, Message: %v\n", msg.Channel, msg.Pattern, temp)
		}
	}()

	// Wait for shutdown signal
	<-shutdown

	// Perform cleanup
	fmt.Println("\nShutting down...")
	reconn.Close()
	fmt.Println("Goodbye!")
}
