/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package connectivity_manager

import (
	"github.com/nalej/grpc-connectivity-manager-go"
	"github.com/nalej/nalej-bus/pkg/queue/infrastructure/events"
)

// Manager structure with the remote clients required
type Manager struct {
	InfrastructureEventsConsumer *events.InfrastructureEventsConsumer
}

// NewManager creates a new manager.
func NewManager(infrastructureEventsConsumer *events.InfrastructureEventsConsumer) (*Manager, error) {
	return &Manager{
		InfrastructureEventsConsumer: infrastructureEventsConsumer,
	}, nil
}

func (m * Manager) ClusterAlive (alive *grpc_connectivity_manager_go.ClusterAlive) error {
	return nil
}