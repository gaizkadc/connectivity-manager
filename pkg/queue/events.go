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

const (
	ClusterOffline = "offline"
	ClusterOnline = "online"
	ClusterOfflineCordon = "offlineCordon"
	ClusterOnlineCordon = "offline"
	ClusterUnknown = "unknown"
)

type InfrastructureEventsHandler struct {
	// reference manager for infrastructure
	connectivityManagerManager *connectivity_manager.Manager
	// events consumer
	consumer *events.InfrastructureEventsConsumer
	// Cluster status: online, offline, onlineCordon, offlineCordon, unknown
	clusterStatus string
}

// Instantiate a new infrastructure events handler to manipulate messages from the infrastructure events queue.
// params:
//  cmManager
//  cons
func NewInfrastructureEventsHandler (connectivityManagerManager *connectivity_manager.Manager, consumer *events.InfrastructureEventsConsumer) InfrastructureEventsHandler {
	ieHandler := InfrastructureEventsHandler{connectivityManagerManager: connectivityManagerManager, consumer: consumer, clusterStatus:ClusterUnknown}
	log.Debug().Str("cluster status", ieHandler.clusterStatus).Msg("new infrastructure events handler created")
	return ieHandler
}

func(i InfrastructureEventsHandler) Run(threshold time.Duration) {
	go i.consumeClusterAlive()
	go i.waitRequests(threshold)
}

// Endless loop waiting for requests
func (i InfrastructureEventsHandler) waitRequests (threshold time.Duration) {
	log.Debug().Msg("wait for requests to be received by the infrastructure events queue")
	for {
		somethingReceived := false
		ctx, cancel := context.WithTimeout(context.Background(), threshold)
		currentTime := time.Now()
		err := i.consumer.Consume(ctx)
		somethingReceived = true
		cancel()
		select {
		case <- ctx.Done():
			// the timeout was reached
			if !somethingReceived {
				i.clusterStatus = ClusterOffline
				log.Debug().Str("cluster status", i.clusterStatus).Msgf("no message received since %s",currentTime.Format(time.RFC3339))
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
		i.clusterStatus = ClusterOnline
		log.Debug().Interface("clusterAlive",received).Str("cluster status", i.clusterStatus).Msg("<- incoming cluster alive check")
		err := i.connectivityManagerManager.ClusterAlive(received)
		if err != nil {
			log.Error().Err(err).Msg("failed processing cluster alive check")
		}
	}
}