package main

import (
	"fmt"

	"github.com/rickNoise/learn-pub-sub-starter/internal/gamelogic"
	"github.com/rickNoise/learn-pub-sub-starter/internal/pubsub"
	"github.com/rickNoise/learn-pub-sub-starter/internal/routing"
)

// Function called handlerPause in the cmd/client application package. It accepts a game state struct and returns a new handler function that accepts a routing.PlayingState struct. This will be the handler we pass into SubscribeJSON that will be called each time a new message is consumed.
func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.Acktype {
	return func(ps routing.PlayingState) pubsub.Acktype {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return pubsub.Ack
	}
}
