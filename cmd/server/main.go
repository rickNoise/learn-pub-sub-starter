package main

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/rickNoise/learn-pub-sub-starter/internal/gamelogic"
	"github.com/rickNoise/learn-pub-sub-starter/internal/pubsub"
	"github.com/rickNoise/learn-pub-sub-starter/internal/routing"
)

func main() {

	// ***************************** //
	// **** CONNECT TO RABBITMQ **** //
	// ***************************** //

	amqpURL := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		log.Fatalf("failed to Dial amqp: %v\n", err)
	}
	defer conn.Close()
	fmt.Println("Peril game server connected to RabbitMQ!")

	// Create a new publishing channel on the Rabbit MQ connection
	chPub, err := conn.Channel()
	if err != nil {
		log.Fatalf("failed to create publishing channel on the AMQP connection: %v\n", err)
	}
	defer chPub.Close()
	fmt.Println("Publishing channel created on the connection successfully!")

	// Declare and bind a queue for game logs, creating a logging channel
	chLog, quLog, err := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogSlug+".*",
		pubsub.QueueDurable,
	)
	if err != nil {
		log.Fatalf("failed to declare and bind game logs queue!: %v", err)
	}
	fmt.Println("Successfully declared and bound game logs queue:", quLog.Name)
	defer chLog.Close()

	// Process game logs by printing to the console
	msgs, err := chLog.Consume(quLog.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to consume game logs channel: %v", err)
	}
	go func() {
		for d := range msgs {
			fmt.Println("chLog message:", d.Body)
			// handle d.Body
			d.Ack(false)
		}
	}()

	// ******************* //
	// **** USER REPL **** //
	// ******************* //

	// print useful server commands to REPL for user
	fmt.Println()
	gamelogic.PrintServerHelp()

	quitRequest := false
	for !quitRequest {
		// get user input
		userInput := []string{}
		for len(userInput) == 0 {
			userInput = gamelogic.GetInput()
		}
		switch userInput[0] {
		case "pause":
			fmt.Println("sending a pause message...")
			// use PublishJSON function to publish a message to the exchange
			err = pubsub.PublishJSON(
				chPub,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{IsPaused: true},
			)
			if err != nil {
				log.Fatalf("failed to publish message to the exchange: %v", err)
			}
		case "resume":
			fmt.Println("sending a resume message...")
			// use PublishJSON function to publish a message to the exchange
			err = pubsub.PublishJSON(
				chPub,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{IsPaused: false},
			)
			if err != nil {
				log.Fatalf("failed to publish message to the exchange: %v", err)
			}
		case "quit":
			fmt.Println("quitting...")
			quitRequest = true
		default:
			fmt.Println("invalid command")
		}
	}
}
