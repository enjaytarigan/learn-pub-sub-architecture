package main

import (
	"fmt"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AcknowledgeType {
	return func(ps routing.PlayingState) pubsub.AcknowledgeType {
		defer fmt.Printf("> ")
		gs.HandlePause(ps)
		return pubsub.Ack
	}
}

func handlerMove(gs *gamelogic.GameState, publishChannel *amqp.Channel) func(gamelogic.ArmyMove) pubsub.AcknowledgeType {
	return func(move gamelogic.ArmyMove) pubsub.AcknowledgeType {
		defer fmt.Printf("> ")
		outcome := gs.HandleMove(move)
		switch outcome {
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			newWar := gamelogic.RecognitionOfWar{
				Attacker: move.Player,
				Defender: gs.GetPlayerSnap(),
			}

			warExchange := routing.WarRecognitionsPrefix + "." + gs.GetUsername()
			err := pubsub.PublishJSON(publishChannel, routing.ExchangePerilTopic, warExchange, newWar)

			if err != nil {
				fmt.Println(err)
			}

			return pubsub.Ack
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.NackDiscard
		default:
			return pubsub.NackDiscard
		}
	}
}

func handlerWar(gs *gamelogic.GameState, publishChannel *amqp.Channel) func(gamelogic.RecognitionOfWar) pubsub.AcknowledgeType {
	return func(war gamelogic.RecognitionOfWar) pubsub.AcknowledgeType {
		defer fmt.Printf("> ")

		outcome, winner, loser := gs.HandleWar(war)

		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeOpponentWon, gamelogic.WarOutcomeYouWon, gamelogic.WarOutcomeDraw:
			logWarOutCome(outcome, war, winner, loser, publishChannel)
			return pubsub.Ack
		default:
			fmt.Printf("Invalid war outcome: %v\n", outcome)
			return pubsub.NackDiscard
		}
	}
}

func logWarOutCome(outcome gamelogic.WarOutcome, war gamelogic.RecognitionOfWar, winner string, loser string, publishChannel *amqp.Channel) {
	var message string

	switch outcome {
	case gamelogic.WarOutcomeOpponentWon, gamelogic.WarOutcomeYouWon:
		message = fmt.Sprintf("%s won a war against %s", winner, loser)
	case gamelogic.WarOutcomeDraw:
		message = fmt.Sprintf("A war between %s and %s resulted in a draw", winner, loser)
	}

	if message == "" {
		return
	}

	var warIniater = war.Attacker
	log := routing.GameLog{
		CurrentTime: time.Now(),
		Username:    warIniater.Username,
		Message:     message,
	}

	err := pubsub.PublishGOB(
		publishChannel,
		routing.ExchangePerilTopic,
		routing.GameLogSlug+"."+warIniater.Username,
		log,
	)

	fmt.Println("sucessfully log the war")

	if err != nil {
		fmt.Printf("logWarOutCome: %v\n", err)
	}

	return
}
