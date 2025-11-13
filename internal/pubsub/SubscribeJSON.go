package pubsub

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) Acktype,
) error {
	// Call DeclareAndBind to make sure that the given queue exists and is bound to the exchange
	ch, qu, err := DeclareAndBind(
		conn,
		exchange,
		queueName,
		key,
		queueType,
	)
	if err != nil {
		return fmt.Errorf("error declaring and binding queue using provided inputs: %w", err)
	}
	fmt.Println("successfully declared and bound queue:", qu.Name)

	deliveryChan, err := ch.Consume(
		qu.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to consume %v: %w", qu.Name, err)
	}

	// range over the channel of deliveries
	go func() {
		for delivery := range deliveryChan {
			var data T
			err := json.Unmarshal(delivery.Body, &data)
			if err != nil {
				fmt.Println("failed to unmarshal delivery body", err)
				_ = delivery.Nack(false, false)
				continue
			}

			panicked := false
			func() {
				defer func() {
					if r := recover(); r != nil {
						panicked = true
						fmt.Println("handlder panic:", r)
					}
				}()
				switch handler(data) {
				case Ack:
					delivery.Ack(false)
					fmt.Println("Acked message")
				case NackRequeue:
					delivery.Nack(false, true)
					fmt.Println("Nacked message, requeued")
				case NackDiscard:
					delivery.Nack(false, false)
					fmt.Println("Nacked message, discarded:")
				default:
					delivery.Nack(false, false)
					fmt.Println("Nacked message, discarded:")
				}
			}()

			if panicked {
				_ = delivery.Nack(false, false)
				continue
			}

			_ = delivery.Ack(false)

		}
	}()

	return nil
}
