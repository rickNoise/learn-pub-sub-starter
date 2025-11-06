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
	amqpURL := "amqp://guest:guest@localhost:5672/"

	// connect to the RabbitMQ server
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		fmt.Printf("failed to Dial amqp: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()
	fmt.Println("RabbitMQ connection successful (client)")

	// prompt user for a username
	username, err := gamelogic.ClientWelcome()
	for err != nil {
		fmt.Println(err, "\npick a username:")
		username, err = gamelogic.ClientWelcome()
	}

	// declare and bind the pause queue
	pauseCh, pauseQu, err := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+username, // pause.username where username is the user's input. The pause section of the name is the routing key constant in the internal/routing package and is joined by a .
		routing.PauseKey,
		pubsub.QueueTransient,
	)
	if err != nil {
		fmt.Printf("error with DeclareAndBind: %v", err)
		os.Exit(1)
	}
	fmt.Printf("Queue %v declared and bound!\n", pauseQu.Name)

	// remove once values are used
	fmt.Println(pauseCh, pauseQu)

	// create a new game state
	gs := gamelogic.NewGameState(username)

	// call pubsub.SubscribeJSON
	err = pubsub.SubscribeJSON[routing.PlayingState](
		conn,
		routing.ExchangePerilDirect,
		"pause."+username,
		routing.PauseKey,
		pubsub.QueueTransient,
		handlerPause(gs),
	)
	if err != nil {
		log.Fatalf("could not subscribe to pause queue")
	}

	// REPL user input loop
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
			_, err := gs.CommandMove(userInput)
			if err != nil {
				fmt.Println(err)
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

	// keep client running
	select {}
}
