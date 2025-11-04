package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"

	pubsub "github.com/rickNoise/learn-pub-sub-starter/internal/pubsub"
	routing "github.com/rickNoise/learn-pub-sub-starter/internal/routing"
)

func main() {
	// where to connect to the RabbitMQ server
	ampqURL := "amqp://guest:guest@localhost:5672/"

	conn, err := amqp.Dial(ampqURL)
	if err != nil {
		fmt.Printf("failed to Dial amqp: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()
	fmt.Println("RabbitMQ connection successful")

	// create a new channel on the Rabbit MQ connection
	ch, err := conn.Channel()
	if err != nil {
		fmt.Printf("failed to create a channel on the connection: %v\n", err)
		os.Exit(1)
	}

	// use PublishJSON function to publish a message to the exchange
	err = pubsub.PublishJSON(
		ch,
		routing.ExchangePerilDirect,
		routing.PauseKey,
		routing.PauseKey,
	)
	if err != nil {
		fmt.Printf("failed to publish message to the exchange: %v\n", err)
		os.Exit(1)
	}

	// wait for ctrl+c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	sig := <-signalChan

	fmt.Printf("\nReceived signal: %v\n", sig)
	fmt.Println("Shutting down gracefully...")
}
