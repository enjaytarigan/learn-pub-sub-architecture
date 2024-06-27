package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

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

	fmt.Println("Peril game client connected to RabbitMQ")

	username, err := gamelogic.ClientWelcome()

	if err != nil {
		log.Fatalf("couldn't get username: %v", err)
	}

	gameState := gamelogic.NewGameState(username)

	// Consume "pause" message
	var pauseQueueName = fmt.Sprintf("%s.%s", routing.PauseKey, username)
	err = pubsub.SubscribeJSON(
		amqpConn,
		routing.ExchangePerilDirect,
		pauseQueueName,
		routing.PauseKey,
		pubsub.Transient,
		handlerPause(gameState),
	)

	if err != nil {
		log.Fatalf("could not subscribe to pause: %v", err)
	}

	// Publish move channel
	pubChannel, err := amqpConn.Channel()

	if err != nil {
		log.Fatalf("could not create publish channel: %v", pubChannel)
	}

	// Listen "army_moves" queue
	var moveQueueName = fmt.Sprintf("army_moves.%s", username)
	var armyMoveRoutingKey = fmt.Sprintf("%s.*", routing.ArmyMovesPrefix)
	pubsub.SubscribeJSON(amqpConn, routing.ExchangePerilTopic, moveQueueName, armyMoveRoutingKey, pubsub.Transient, handlerMove(gameState, pubChannel))

	// Listen "war" message
	warQueueName := routing.WarRecognitionsPrefix
	warRoutingKey := routing.WarRecognitionsPrefix + ".*"
	pubsub.SubscribeJSON(amqpConn, routing.ExchangePerilTopic, warQueueName, warRoutingKey, pubsub.Durable, handlerWar(gameState, pubChannel))

	gamelogic.PrintClientHelp()

	for {

		words := gamelogic.GetInput()

		if len(words) == 0 {
			continue
		}

		cmd := words[0]

		switch cmd {
		case "move":
			move, err := gameState.CommandMove(words)
			if err != nil {
				fmt.Println(err)
				continue
			}

			var userMoveRoutingKey = routing.ArmyMovesPrefix + "." + username
			err = pubsub.PublishJSON(pubChannel, routing.ExchangePerilTopic, userMoveRoutingKey, move)

			if err != nil {
				fmt.Println(err)
			}
		case "spawn":
			err = gameState.CommandSpawn(words)
			if err != nil {
				fmt.Println(err)
				continue
			}
		case "status":
			gameState.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			// TODO: publish n malicious logs
			if len(words) < 2 {
				continue
			}

			totalSpamToBeSent, err := strconv.Atoi(words[1])

			if err != nil {
				continue
			}

			for i := 0; i < totalSpamToBeSent; i++ {
				log := routing.GameLog{
					Message:     gamelogic.GetMaliciousLog(),
					CurrentTime: time.Now(),
					Username:    username,
				}
				err := pubsub.PublishGOB(pubChannel, routing.ExchangePerilTopic, routing.GameLogSlug+"."+username, log)

				if err != nil {
					fmt.Println(err)
				}
			}
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Println("unknown command")
		}
	}
}
