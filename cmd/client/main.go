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
	fmt.Println("Starting Peril client...")

	// ***************************** //
	// ** Connect to RabbitMQ ****** //
	// ***************************** //

	amqpURL := "amqp://guest:guest@localhost:5672/"

	// Create connection
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		log.Fatalf("failed to Dial amqp: %v\n", err)
	}
	defer conn.Close()
	fmt.Println("Peril client successfully connected to RabbitMQ!")

	// Create a channel for publishing messages on the connection
	publishCh, err := conn.Channel()
	if err != nil {
		log.Fatalf("could not create publish channel: %v", err)
	}

	// ***************************** //
	// ** Initialise game state **** //
	// ***************************** //

	// Prompt user for a username
	username, err := gamelogic.ClientWelcome()
	for err != nil {
		fmt.Println(err, "\npick a username: ")
		username, err = gamelogic.ClientWelcome()
	}

	// Create a new game state
	gs := gamelogic.NewGameState(username)

	// ******************************** //
	// ** Subscribe to player queues ** //
	// ******************************** //

	// Subscribe to the pause queue
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+gs.GetUsername(),
		routing.PauseKey,
		pubsub.QueueTransient,
		handlerPause(gs),
	)
	if err != nil {
		log.Fatalf("could not subscribe to pause queue: %v", err)
	}
	fmt.Println("Client subscribed to the pause queue!")

	// Subscribe to the move queue
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+gs.GetUsername(),
		routing.ArmyMovesPrefix+".*",
		pubsub.QueueTransient,
		handlerMove(gs, publishCh),
	)
	if err != nil {
		log.Fatalf("could not subscribe to move queue: %v", err)
	}
	fmt.Println("Client subscribed to the move queue!")

	// Subscribe the war queue
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		"war",
		routing.WarRecognitionsPrefix+".*",
		pubsub.QueueDurable,
		handlerWar(gs, publishCh),
	)
	if err != nil {
		log.Fatalf("could bot subscribe to war queue: %v", err)
	}
	fmt.Println("Client subscribed to the war queue!")

	// ******************************** //
	// ** User Input REPL ************* //
	// ******************************** //

	for {
		userInput := gamelogic.GetInput()
		if len(userInput) == 0 {
			continue
		}

		switch userInput[0] {
		case "spawn":
			// expects e.g. spawn europe infantry
			err = gs.CommandSpawn(userInput)
			if err != nil {
				fmt.Println(err)
				continue
			}
		case "move":
			// expects e.g. move europe 1
			mv, err := gs.CommandMove(userInput)
			if err != nil {
				fmt.Println(err)
				continue
			}

			err = pubsub.PublishJSON(
				publishCh,
				routing.ExchangePerilTopic,
				routing.ArmyMovesPrefix+"."+mv.Player.Username,
				mv,
			)
			if err != nil {
				fmt.Printf("error: %s\n", err)
				continue
			}
			fmt.Printf("Moved %v units to %s\n", len(mv.Units), mv.ToLocation)
		case "status":
			gs.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Println("Invalid command!")
		}
	}
}
