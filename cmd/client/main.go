package main

import (
	"fmt"
	"log"
	"os"
	"slices"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/rickNoise/learn-pub-sub-starter/internal/gamelogic"
	"github.com/rickNoise/learn-pub-sub-starter/internal/pubsub"
	"github.com/rickNoise/learn-pub-sub-starter/internal/routing"
)

func main() {
	// CONSTANTS
	unitValues := []string{"infantry", "cavalry", "artillery"}
	locationValues := []string{"americas", "europe", "africa", "asia", "antarctica", "australia"}

	fmt.Println("Starting Peril client...")

	// ***************************** //
	// **** CONNECT TO RABBITMQ **** //
	// ***************************** //

	amqpURL := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		log.Fatalf("failed to Dial amqp: %v\n", err)
	}
	defer conn.Close()
	fmt.Println("Peril client successfully created connection to RabbitMQ!")

	// Prompt user for a username
	username, err := gamelogic.ClientWelcome()
	for err != nil {
		fmt.Println(err, "\npick a username:")
		username, err = gamelogic.ClientWelcome()
	}

	// Declare and bind the pause queue
	chPause, quPause, err := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+username,
		routing.PauseKey,
		pubsub.QueueTransient,
	)
	if err != nil {
		log.Fatalf("error with DeclareAndBind: %v", err)
	}
	fmt.Printf("Queue %v declared and bound!\n", quPause.Name)
	defer chPause.Close()

	// Declare and bind the queue receiving moves by other players
	moveCh, moveQu, err := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilTopic,
		"army_moves."+username,
		"army_moves.*",
		pubsub.QueueTransient,
	)
	if err != nil {
		fmt.Printf("error with DeclareAndBind: %v", err)
		os.Exit(1)
	}
	fmt.Printf("Queue %v declared and bound!\n", moveQu.Name)
	defer moveCh.Close()

	// Create a new game state
	gs := gamelogic.NewGameState(username)

	// Subscribe to the pause queue
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		quPause.Name,
		routing.PauseKey,
		pubsub.QueueTransient,
		handlerPause(gs),
	)
	if err != nil {
		log.Fatalf("could not subscribe to pause queue")
	}

	// Subscribe to the move queue
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		moveQu.Name,
		"army_moves.*",
		pubsub.QueueTransient,
		handlerMove(gs),
	)
	if err != nil {
		log.Fatalf("could not subscribe to move queue")
	}

	// ***************************** //
	// **** USER INPUT REPL ******** //
	// ***************************** //

	quitRequest := false
	for !quitRequest {
		// get user input
		userInput := []string{}
		for len(userInput) == 0 {
			userInput = gamelogic.GetInput()
		}

		// logic based on input
		switch userInput[0] {
		case "spawn":
			// expects e.g. spawn europe infantry
			if len(userInput) != 3 {
				fmt.Println("example usage: spawn europe infantry")
				break
			}
			location := userInput[1]
			unit := userInput[2]
			if !slices.Contains(locationValues, location) || !slices.Contains(unitValues, unit) {
				fmt.Println("example usage: spawn europe infantry")
				fmt.Println("available locations:", locationValues)
				fmt.Println("available unit types:", unitValues)
				break
			}
			err := gs.CommandSpawn(userInput)
			if err != nil {
				fmt.Println("error spawning units:", err)
			}
		case "move":
			// expects e.g. move europe 1
			if len(userInput) != 3 {
				fmt.Println("example usage: move europe 1")
				break
			}
			location := userInput[1]
			if !slices.Contains(locationValues, location) {
				fmt.Println("example usage: move europe 1")
				fmt.Println("available locations:", locationValues)
				break
			}
			move, err := gs.CommandMove(userInput)
			if err != nil {
				fmt.Println(err)
			} else {
				// publish the move
				fmt.Println("sending a move message...")
				// use PublishJSON function to publish a message to the exchange
				err = pubsub.PublishJSON(
					moveCh,
					routing.ExchangePerilTopic,
					"army_moves."+username,
					move,
				)
				if err != nil {
					fmt.Println("failed to publish message to the exchange:", err)
					os.Exit(1)
				}
				fmt.Println("successfully published move!")
			}
		case "status":
			gs.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			quitRequest = true
		default:
			fmt.Println("Invalid command!")
		}
	}
}
