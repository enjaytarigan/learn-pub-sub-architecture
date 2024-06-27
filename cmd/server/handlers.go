package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func handlersLog(log routing.GameLog) pubsub.AcknowledgeType {
	err := gamelogic.WriteLog(log)

	if err != nil {
		fmt.Printf("could not write game log: %v", err)
		return pubsub.NackDiscard
	}

	return pubsub.Ack
}
