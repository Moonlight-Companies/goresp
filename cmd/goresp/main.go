package main

import (
	// "flag"
	// "fmt"
	// "log"
	// "os"
	// "os/signal"
	// "strings"
	// "syscall"

	"flag"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Moonlight-Companies/goresp/connection"
	"github.com/Moonlight-Companies/goresp/logging"
)

func main() {
	log := logging.NewLogger(logging.LogLevelInfo)

	redisAddr := flag.String("redis", "bus:6379", "Redis server address")
	channelsFlag := flag.String("channels", "*", "Comma-separated list of channels to subscribe to")
	flag.Parse()

	channels := strings.Split(*channelsFlag, ",")

	reconn := connection.NewReconnecting(*redisAddr)

	for _, channel := range channels {
		channel = strings.TrimSpace(channel)
		if strings.Contains(channel, "*") {
			reconn.PSubscribe(channel)
			log.Infoln("PSubscribed to channel", channel)
		} else {
			reconn.Subscribe(channel)
			log.Infoln("Subscribed to channel", channel)
		}
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	log.Infoln("Connected to Redis", *redisAddr)
	log.Infoln("Waiting for messages. Press Ctrl+C to exit.")

	go func() {
		for msg := range reconn.Messages {
			temp, err := msg.IntoMap()
			if err != nil {
				log.Errorln("Error parsing message", err)
				continue
			}
			log.Info("Channel: %s, Pattern: %s, Message: %v\n", msg.Channel, msg.Pattern, temp)
		}
	}()

	// Wait for shutdown signal
	<-shutdown

	// Perform cleanup
	log.Infoln("Calling .Close...")
	reconn.Close()
	log.Infoln("Done!")
}
