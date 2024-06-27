package pubsub

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func subscribe[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType QueueType,
	handler func(T) AcknowledgeType,
	unmarshaller func([]byte) (T, error),
) error {
	ch, q, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)

	if err != nil {
		return fmt.Errorf("pubsub: %w", err)
	}

	err = ch.Qos(10, 0, true)

	if err != nil {
		return fmt.Errorf("pubsub: %w", err)
	}
	delivery, err := ch.Consume(q.Name, "", false, false, false, false, nil)

	if err != nil {
		return fmt.Errorf("pubsub: %w", err)
	}

	go func() {
		defer ch.Close()
		for d := range delivery {
			body, err := unmarshaller(d.Body)

			if err != nil {
				fmt.Printf("pubsub: could not unmarshal message: %v\n", err)
				continue
			}

			ack := handler(body)

			switch ack {
			case Ack:
				d.Ack(false)
			case NackRequeue:
				d.Nack(false, true)
			case NackDiscard:
				d.Nack(false, false)
			}
		}
	}()

	return nil
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType QueueType,
	handler func(T) AcknowledgeType,
) error {
	jsonUnmarshaller := func(buf []byte) (T, error) {
		var body T
		err := json.Unmarshal(buf, &body)
		if err != nil {
			fmt.Printf("could not unmarshal message: %v\n", err)
		}

		return body, nil
	}

	return subscribe(conn, exchange, queueName, key, simpleQueueType, handler, jsonUnmarshaller)
}

func SubscribeGOB[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType QueueType,
	handler func(T) AcknowledgeType,
) error {
	var gobUnmarshaller = func(buf []byte) (T, error) {
		msg := bytes.NewBuffer(buf)

		var messageData T
		err := gob.NewDecoder(msg).Decode(&messageData)

		return messageData, err
	}

	return subscribe(conn, exchange, queueName, key, simpleQueueType, handler, gobUnmarshaller)
}
