package main

import (
	"fmt"
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/rickNoise/learn-pub-sub-starter/internal/gamelogic"
	"github.com/rickNoise/learn-pub-sub-starter/internal/pubsub"
	"github.com/rickNoise/learn-pub-sub-starter/internal/routing"
)

func main() {
	// **** CONNECT TO RABBITMQ **** //
	ampqURL := "amqp://guest:guest@localhost:5672/"

	// connect to the RabbitMQ server
	conn, err := amqp.Dial(ampqURL)
	if err != nil {
		fmt.Printf("failed to Dial amqp: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()
	fmt.Println("Peril game server connected to RabbitMQ!")

	// create a new channel on the Rabbit MQ connection
	ch, err := conn.Channel()
	if err != nil {
		fmt.Printf("failed to create a channel on the connection: %v\n", err)
		os.Exit(1)
	}
	defer ch.Close()
	fmt.Println("channel succsesfully created on the connection")

	// declare and bind a queue for game logs
	_, qu, err := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogSlug+".*",
		pubsub.QueueDurable,
	)
	if err != nil {
		log.Fatalf("failed to declare and bind game logs queue!")
	}
	fmt.Println("successfully declared and bound game logs queue:", qu.Name)

	// print useful server commands to REPL for user
	gamelogic.PrintServerHelp()

	// **** USER REPL **** //

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
				ch,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{IsPaused: true},
			)
			if err != nil {
				fmt.Println("failed to publish message to the exchange:", err)
				os.Exit(1)
			}
		case "resume":
			fmt.Println("sending a resume message...")
			// use PublishJSON function to publish a message to the exchange
			err = pubsub.PublishJSON(
				ch,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{IsPaused: false},
			)
			if err != nil {
				fmt.Println("failed to publish message to the exchange:", err)
				os.Exit(1)
			}
		case "quit":
			fmt.Println("quitting...")
			quitRequest = true
		default:
			fmt.Println("invalid command")
		}
	}
}
