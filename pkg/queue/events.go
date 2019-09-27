/*
 * Copyright (C)  2019 Nalej - All Rights Reserved
 */

package queue

import (
	"context"
	"github.com/nalej/connectivity-manager/pkg/server/connectivity-manager"
	"github.com/nalej/nalej-bus/pkg/queue/infrastructure/events"
	"github.com/rs/zerolog/log"
	"time"
)

// Timeout between incoming messages
const InfrastructureEventsTimeout = time.Minute * 60

type InfrastructureEventsHandler struct {
	// reference manager for infrastructure
	connectivityManagerManager *connectivity_manager.Manager
	// events consumer
	consumer *events.InfrastructureEventsConsumer
}

type ClusterStatus struct {
	Online bool
	Offline bool
	OnlineCordon bool
	OfflineCordon bool
}

// Instantiate a new network ops handler to manipulate messages from the network ops queue.
// params:
//  netManager
//  cons
func NewInfrastructureEventsHandler (connectivityManagerManager *connectivity_manager.Manager, consumer *events.InfrastructureEventsConsumer) InfrastructureEventsHandler {
	return InfrastructureEventsHandler{connectivityManagerManager: connectivityManagerManager, consumer: consumer}
}

func(i InfrastructureEventsHandler) Run() {
	go i.consumeClusterAlive()
	go i.waitRequests()
}

// Endless loop waiting for requests
func (i InfrastructureEventsHandler) waitRequests () {
	log.Debug().Msg("wait for requests to be received by the infrastructure events queue")
	for {
		somethingReceived := false
		ctx, cancel := context.WithTimeout(context.Background(), InfrastructureEventsTimeout)
		currentTime := time.Now()
		err := i.consumer.Consume(ctx)
		somethingReceived = true
		cancel()
		select {
		case <- ctx.Done():
			// the timeout was reached
			if !somethingReceived {
				log.Debug().Msgf("no message received since %s",currentTime.Format(time.RFC3339))
			}
		default:
			if err != nil {
				log.Error().Err(err).Msg("error consuming data from infrastructure events")
			}
		}
	}
}

func (i InfrastructureEventsHandler) consumeClusterAlive () {
	log.Debug().Msg("waiting for cluster alive checks...")
	for {
		received := <- i.consumer.Config.ChClusterAlive
		log.Debug().Interface("clusterAlive",received).Msg("<- incoming cluster alive check")
		err := i.connectivityManagerManager.ClusterAlive(received)
		if err != nil {
			log.Error().Err(err).Msg("failed processing cluster alive check")
		}
	}
}