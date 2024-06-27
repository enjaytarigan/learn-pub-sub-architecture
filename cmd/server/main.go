package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	const amqpConnString = "amqp://guest:guest@localhost:5672/"

	amqpConn, err := amqp.Dial(amqpConnString)

	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %v", err)
	}
	defer amqpConn.Close()
	fmt.Println("Peril game server connected to RabbitMQ!")

	pubChannel, err := amqpConn.Channel()
	if err != nil {
		log.Fatalf("could not create channel: %v", err)
	}

	err = pubsub.SubscribeGOB(amqpConn, routing.ExchangePerilTopic, routing.GameLogSlug, routing.GameLogSlug+".*", pubsub.Durable, handlersLog)

	if err != nil {
		log.Fatalf("could not subscribe to game logs: %v", err)
	}

	gamelogic.PrintServerHelp()
	for {
		words := gamelogic.GetInput()

		if len(words) == 0 {
			continue
		}

		cmd := words[0]

		switch cmd {
		case "pause":
			fmt.Println("Sending a pause message")
			err := pubsub.PublishJSON(pubChannel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: true,
			})

			if err != nil {
				log.Printf("could not publish time: %v", err)
			}
		case "resume":
			fmt.Println("Sending a resume message...")
			err := pubsub.PublishJSON(pubChannel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: false,
			})

			if err != nil {
				log.Printf("could not publish time: %v", err)
			}
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Println("Unknown command")
			continue
		}
	}
}
