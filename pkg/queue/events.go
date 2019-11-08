/*
 * Copyright 2019 Nalej
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package queue

import (
	"context"
	"fmt"
	"github.com/hashicorp/golang-lru"
	"github.com/nalej/connectivity-manager/pkg/server/connectivity-manager"
	"github.com/nalej/nalej-bus/pkg/queue/infrastructure/events"
	"github.com/rs/zerolog/log"
	"time"
)

const (
	DefaultTimeout   = 2 * time.Minute
	MaxCachedEntries = 50
)

type InfrastructureEventsHandler struct {
	// reference manager for infrastructure
	manager *connectivity_manager.Manager
	// events consumer
	consumer *events.InfrastructureEventsConsumer
	// cache
	clusterCache *lru.Cache
}

// Instantiate a new infrastructure events handler to manipulate messages from the infrastructure events queue.
// params:
//  cmManager
//  cons
func NewInfrastructureEventsHandler(connectivityManagerManager *connectivity_manager.Manager, consumer *events.InfrastructureEventsConsumer) InfrastructureEventsHandler {
	ieHandler := InfrastructureEventsHandler{manager: connectivityManagerManager, consumer: consumer}
	log.Debug().Msg("new infrastructure events handler created")
	return ieHandler
}

type clusterCacheEntry struct {
	timestamp     time.Time
	clusterStatus string
}

func newClusterCacheEntry(clusterStatus string) *clusterCacheEntry {
	return &clusterCacheEntry{
		timestamp:     time.Now(),
		clusterStatus: clusterStatus,
	}
}

func (i InfrastructureEventsHandler) Run(threshold time.Duration) {
	clusterCache, err := lru.New(MaxCachedEntries)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create cache")
	}
	i.clusterCache = clusterCache
	go i.consumeClusterAlive()
	go i.waitRequests()
	go i.checkClusterStatusExpiration(threshold)
}

// Endless loop waiting for requests
func (i InfrastructureEventsHandler) waitRequests() {
	log.Debug().Msg("wait for requests to be received by the infrastructure events queue")
	for {
		somethingReceived := false
		rCtx, rCancel := context.WithTimeout(context.Background(), DefaultTimeout)
		defer rCancel()
		currentTime := time.Now()
		err := i.consumer.Consume(rCtx)
		somethingReceived = true
		select {
		case <-rCtx.Done():
			// the timeout was reached
			if !somethingReceived {
				log.Debug().Msgf("no message received since %s", currentTime.Format(time.RFC3339))
			}
		default:
			if err != nil {
				log.Error().Err(err).Msg("error consuming data from infrastructure events")
			}
		}
	}
}

func (i InfrastructureEventsHandler) getClusterKey(organizationID string, clusterID string) string {
	return fmt.Sprintf("%s#%s", organizationID, clusterID)
}

func (i InfrastructureEventsHandler) consumeClusterAlive() {
	log.Debug().Msg("waiting for cluster alive checks...")
	for {
		received := <-i.consumer.Config.ChClusterAlive
		i.manager.ClusterAlive(received)
	}
}

func (i InfrastructureEventsHandler) checkClusterStatusExpiration(threshold time.Duration) {
	for {
		ticker := time.NewTicker(threshold)
		select {
		case <-ticker.C:
			i.manager.TransitionClustersToOffline()
		}
	}
}
