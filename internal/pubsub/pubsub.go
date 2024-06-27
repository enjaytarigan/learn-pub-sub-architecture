package pubsub

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

type QueueType int

type AcknowledgeType int

const (
	_ AcknowledgeType = iota
	Ack
	NackRequeue
	NackDiscard
)

const (
	_ QueueType = iota
	Durable
	Transient
)

func DeclareAndBind(conn *amqp.Connection, exchange, queueName, key string, simpleQueueType QueueType) (*amqp.Channel, amqp.Queue, error) {
	ch, err := conn.Channel()

	if err != nil {
		return ch, amqp.Queue{}, fmt.Errorf("pubsub: %w", err)
	}

	var isDurable = simpleQueueType == Durable
	var autoDelete = simpleQueueType == Transient
	var isTransient = simpleQueueType == Transient

	queue, err := ch.QueueDeclare(queueName, isDurable, autoDelete, isTransient, false, amqp.Table{
		"x-dead-letter-exchange": routing.ExchangePerilDeadLetter,
	})

	if err != nil {
		return ch, amqp.Queue{}, fmt.Errorf("pubsub: %w", err)
	}

	err = ch.QueueBind(queue.Name, key, exchange, false, nil)

	if err != nil {
		return ch, amqp.Queue{}, fmt.Errorf("pubsub: %w", err)
	}

	return ch, queue, nil
}
