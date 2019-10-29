/*
 * Copyright (C)  2019 Nalej - All Rights Reserved
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
func NewInfrastructureOpsHandler (connectivityManagerManager *connectivity_manager.Manager, producer *ops.InfrastructureOpsProducer) InfrastructureOpsHandler {
	ioHandler := InfrastructureOpsHandler{manager: connectivityManagerManager, producer: producer}
	log.Debug().Msg("new infrastructure ops handler created")
	return ioHandler
}

// Send messages to the infrastructure ops queue
func(i InfrastructureOpsHandler) Send(ctx context.Context, msg proto.Message) derrors.Error {
	return i.producer.Send(ctx,msg)
}