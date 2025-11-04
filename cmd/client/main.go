package main

import (
	"fmt"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"

	gamelogic "github.com/rickNoise/learn-pub-sub-starter/internal/gamelogic"
	pubsub "github.com/rickNoise/learn-pub-sub-starter/internal/pubsub"
	routing "github.com/rickNoise/learn-pub-sub-starter/internal/routing"
)

func main() {
	fmt.Println("Starting Peril client...")

	// connect to the RabbitMQ server
	ampqURL := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(ampqURL)
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

	ch, qu, err := pubsub.DeclareAndBind(
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

	// remove once values are used
	fmt.Println(ch, qu)

	// keep client running
	select {}
}
