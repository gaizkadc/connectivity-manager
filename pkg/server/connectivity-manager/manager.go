/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package connectivity_manager

import (
	"context"
	"github.com/nalej/connectivity-manager/pkg/server/config"
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-common-go"
	"github.com/nalej/grpc-connectivity-manager-go"
	"github.com/nalej/grpc-infrastructure-go"
	"github.com/nalej/grpc-organization-go"
	"github.com/nalej/grpc-utils/pkg/conversions"
	"github.com/nalej/nalej-bus/pkg/queue/infrastructure/events"
	"github.com/rs/zerolog/log"
	"time"
)


const (
	DefaultTimeout =  time.Minute
)

// Manager structure with the remote clients required
type Manager struct {
	OrganizationsClient          grpc_organization_go.OrganizationsClient
	ClustersClient               grpc_infrastructure_go.ClustersClient
	InfrastructureEventsConsumer *events.InfrastructureEventsConsumer
	config                       config.Config
}



// NewManager creates a new manager.
func NewManager(clustersClient *grpc_infrastructure_go.ClustersClient,organizationsClient *grpc_organization_go.OrganizationsClient,infrastructureEventsConsumer *events.InfrastructureEventsConsumer, config config.Config) (*Manager, error) {
	return &Manager{
		ClustersClient: *clustersClient,
		InfrastructureEventsConsumer: infrastructureEventsConsumer,
		OrganizationsClient: *organizationsClient,
		config: config,
	}, nil
}

func (m * Manager) ClusterAlive (alive *grpc_connectivity_manager_go.ClusterAlive) derrors.Error {
	log.Debug().Interface("clusterAlive", alive).Msg("<- incoming cluster alive check")

	clusterID := &grpc_infrastructure_go.ClusterId{
		OrganizationId:       alive.OrganizationId,
		ClusterId:            alive.ClusterId,
	}
	getCtx, getCancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer getCancel()
	previous, err := m.ClustersClient.GetCluster(getCtx, clusterID)
	if err != nil{
		log.Error().Str("trace", conversions.ToDerror(err).DebugReport()).Msg("unable to get cluster")
		return conversions.ToDerror(err)
	}


	updateClusterRequest := &grpc_infrastructure_go.UpdateClusterRequest{
		OrganizationId:       alive.OrganizationId,
		ClusterId:            alive.ClusterId,
		UpdateLastClusterTimestamp : true,
		LastClusterTimestamp: alive.Timestamp,
	}

	if previous.ClusterStatus == grpc_connectivity_manager_go.ClusterStatus_OFFLINE || previous.ClusterStatus == grpc_connectivity_manager_go.ClusterStatus_UNKNOWN{
		updateClusterRequest.UpdateStatus = true
		updateClusterRequest.Status = grpc_connectivity_manager_go.ClusterStatus_ONLINE
	}

	updateCtx, updateCancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer updateCancel()
	_, err = m.ClustersClient.UpdateCluster(updateCtx, updateClusterRequest)
	if err != nil{
		log.Error().Str("trace", conversions.ToDerror(err).DebugReport()).Msg("unable to update cluster")
		return conversions.ToDerror(err)
	}

	return nil
}

func (m *Manager) TransitionClustersToOffline() {
	// TODO Get only clusters that are online or online_cordon using a specific endpoint
	orgCtx, orgCancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer orgCancel()
	organizations, err := m.OrganizationsClient.ListOrganizations(orgCtx, &grpc_common_go.Empty{})
	if err != nil{
		log.Error().Str("trace", conversions.ToDerror(err).DebugReport()).Msg("unable to get the list of organization, skipping transitioning clusters to offline")
		return
	}
	for _, org := range organizations.Organizations{
		log.Debug().Str("organizationID", org.OrganizationId).Msg("checking organization clusters")
		organizationID := &grpc_organization_go.OrganizationId{
			OrganizationId:       org.OrganizationId,
		}
		clusterCtx, clusterCancel := context.WithTimeout(context.Background(), DefaultTimeout)
		defer clusterCancel()
		clusters, err := m.ClustersClient.ListClusters(clusterCtx, organizationID)
		if err != nil{
			log.Error().Str("organizationID", org.OrganizationId).Str("trace", conversions.ToDerror(err).DebugReport()).Msg("unable to get the list of organization clusters, skipping transitioning clusters to offline in that organization")
		}else{
			for _, cluster := range clusters.Clusters{
				m.checkClusterTransitionToOffline(cluster)
			}
		}
	}

}

func (m *Manager) checkClusterTransitionToOffline(cluster *grpc_infrastructure_go.Cluster) {
	// check if the cluster is online or online_cordon
	if cluster.LastAliveTimestamp < time.Now().Unix() - m.config.Threshold.Nanoseconds() {
		var nextStatus grpc_connectivity_manager_go.ClusterStatus
		send := false
		if cluster.ClusterStatus == grpc_connectivity_manager_go.ClusterStatus_ONLINE{
			nextStatus = grpc_connectivity_manager_go.ClusterStatus_OFFLINE
			send = true
		}else if cluster.ClusterStatus == grpc_connectivity_manager_go.ClusterStatus_ONLINE_CORDON{
			nextStatus = grpc_connectivity_manager_go.ClusterStatus_OFFLINE_CORDON
			send = true
		}

		if send {
			updateClusterRequest := &grpc_infrastructure_go.UpdateClusterRequest {
				OrganizationId:       cluster.OrganizationId,
				ClusterId:            cluster.ClusterId,
				UpdateStatus: true,
				Status: nextStatus,
			}
			updateCtx, updateCancel := context.WithTimeout(context.Background(), DefaultTimeout)
			defer updateCancel()
			_, err := m.ClustersClient.UpdateCluster(updateCtx, updateClusterRequest)
			if err != nil{
				log.Error().Interface("update", updateClusterRequest).Str("trace", conversions.ToDerror(err).DebugReport()).Msg("unable to transition cluster to OFFLINE*")
			}
		}
	}else if cluster.ClusterStatus == grpc_connectivity_manager_go.ClusterStatus_OFFLINE {
		if cluster.LastAliveTimestamp < time.Now().Unix() - cluster.GracePeriod{
			// TODO Trigger policy
			updateClusterRequest := &grpc_infrastructure_go.UpdateClusterRequest {
				OrganizationId:       cluster.OrganizationId,
				ClusterId:            cluster.ClusterId,
				UpdateStatus: true,
				Status: grpc_connectivity_manager_go.ClusterStatus_OFFLINE_CORDON,
			}
			updateCtx, updateCancel := context.WithTimeout(context.Background(), DefaultTimeout)
			defer updateCancel()
			_, err := m.ClustersClient.UpdateCluster(updateCtx, updateClusterRequest)
			if err != nil{
				log.Error().Interface("update", updateClusterRequest).Str("trace", conversions.ToDerror(err).DebugReport()).Msg("unable to transition cluster to OFFLINE_CORDON")
			}
		}
	}
}
