/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package connectivity_manager

import (
	"github.com/nalej/grpc-connectivity-manager-go"
)

// Manager structure with the remote clients required
type Manager struct {
}

// NewManager creates a new manager.
func NewManager() (*Manager, error) {
	return &Manager{}, nil
}

func (m * Manager) ClusterAlive (alive *grpc_connectivity_manager_go.ClusterAlive) error {
	return nil
}