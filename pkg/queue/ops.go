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
	"github.com/golang/protobuf/proto"
	"github.com/nalej/connectivity-manager/pkg/server/connectivity-manager"
	"github.com/nalej/derrors"
	"github.com/nalej/nalej-bus/pkg/queue/infrastructure/ops"
	"github.com/rs/zerolog/log"
)

type InfrastructureOpsHandler struct {
	// reference manager for infrastructure
	manager *connectivity_manager.Manager
	// ops producer
	producer *ops.InfrastructureOpsProducer
}

// Instantiate a new infrastructure ops handler to manipulate messages from the infrastructure ops queue.
func NewInfrastructureOpsHandler(connectivityManagerManager *connectivity_manager.Manager, producer *ops.InfrastructureOpsProducer) InfrastructureOpsHandler {
	ioHandler := InfrastructureOpsHandler{manager: connectivityManagerManager, producer: producer}
	log.Debug().Msg("new infrastructure ops handler created")
	return ioHandler
}

// Send messages to the infrastructure ops queue
func (i InfrastructureOpsHandler) Send(ctx context.Context, msg proto.Message) derrors.Error {
	return i.producer.Send(ctx, msg)
}
