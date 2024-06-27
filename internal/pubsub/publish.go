package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	v, err := json.Marshal(val)

	if err != nil {
		return fmt.Errorf("pubsub: %w", err)
	}

	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        v,
	})

	if err != nil {
		return fmt.Errorf("pubsub: %w", err)
	}

	return nil
}

func PublishGOB[T any](ch *amqp.Channel, exchange, key string, val T) error {
	var reader bytes.Buffer
	err := gob.NewEncoder(&reader).Encode(val)

	if err != nil {
		return fmt.Errorf("pubsub: %w", err)
	}

	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/gob",
		Body:        reader.Bytes(),
	})

	if err != nil {
		return fmt.Errorf("pubsub: %w", err)
	}

	return nil
}
