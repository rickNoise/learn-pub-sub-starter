package pubsub

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// enum
type SimpleQueueType string

const (
	QueueDurable   SimpleQueueType = "durable"
	QueueTransient SimpleQueueType = "transient"
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {
	// create a new channel on the connection
	ch, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("error creating a channel: %v", err)
	}

	// declare a new queue
	qu, err := ch.QueueDeclare(
		queueName,
		queueType == QueueDurable,
		queueType == QueueTransient,
		queueType == QueueTransient,
		false,
		nil,
	)
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("error declaring a new queue: %v", err)
	}

	// Bind the queue to the exchange
	err = ch.QueueBind(
		qu.Name,
		key,
		exchange,
		false,
		nil,
	)
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("error binding the queue to the exchange: %v", err)
	}

	return ch, qu, nil
}
