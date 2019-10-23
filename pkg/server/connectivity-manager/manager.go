/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package connectivity_manager

import (
	"context"
	"github.com/nalej/connectivity-manager/pkg/server/config"
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-common-go"
	grpc_conductor_go "github.com/nalej/grpc-conductor-go"
	"github.com/nalej/grpc-connectivity-manager-go"
	"github.com/nalej/grpc-infrastructure-go"
	"github.com/nalej/grpc-organization-go"
	"github.com/nalej/grpc-utils/pkg/conversions"
	"github.com/nalej/nalej-bus/pkg/queue/infrastructure/events"
	"github.com/nalej/nalej-bus/pkg/queue/infrastructure/ops"
	"github.com/rs/zerolog/log"
	"time"
)


const (
	DefaultTimeout =  2*time.Minute
)

// Manager structure with the remote clients required
type Manager struct {
	OrganizationsClient          grpc_organization_go.OrganizationsClient
	ClustersClient               grpc_infrastructure_go.ClustersClient
	InfrastructureOpsProducer	*ops.InfrastructureOpsProducer
	InfrastructureEventsConsumer *events.InfrastructureEventsConsumer
	config                       config.Config
}



// NewManager creates a new manager.
func NewManager(clustersClient *grpc_infrastructure_go.ClustersClient,
	organizationsClient *grpc_organization_go.OrganizationsClient,
	infrastructureEventsConsumer *events.InfrastructureEventsConsumer,
	infrastructureOpsProducer *ops.InfrastructureOpsProducer,
	config config.Config) (*Manager, error) {
	return &Manager{
		ClustersClient: *clustersClient,
		OrganizationsClient: *organizationsClient,
		InfrastructureEventsConsumer: infrastructureEventsConsumer,
		InfrastructureOpsProducer:infrastructureOpsProducer,
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

	var nextStatus grpc_connectivity_manager_go.ClusterStatus
	send := false
	if previous.ClusterStatus == grpc_connectivity_manager_go.ClusterStatus_OFFLINE || previous.ClusterStatus == grpc_connectivity_manager_go.ClusterStatus_UNKNOWN {
		nextStatus = grpc_connectivity_manager_go.ClusterStatus_ONLINE
		send = true
	} else if previous.ClusterStatus == grpc_connectivity_manager_go.ClusterStatus_OFFLINE_CORDON {
		nextStatus = grpc_connectivity_manager_go.ClusterStatus_ONLINE_CORDON
		send = true
	}

	if send {
		updateClusterRequest.UpdateStatus = true
		updateClusterRequest.Status = nextStatus
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
				m.checkTransitionClusterToOffline(cluster)
			}
		}
	}

}

func (m *Manager) checkTransitionClusterToOffline(cluster *grpc_infrastructure_go.Cluster) {
	if time.Now().Unix() - cluster.LastAliveTimestamp > int64(m.config.Threshold.Seconds()) {
		var nextStatus grpc_connectivity_manager_go.ClusterStatus
		send := false
		if cluster.ClusterStatus == grpc_connectivity_manager_go.ClusterStatus_ONLINE {
			nextStatus = grpc_connectivity_manager_go.ClusterStatus_OFFLINE
			send = true
		}else if cluster.ClusterStatus == grpc_connectivity_manager_go.ClusterStatus_ONLINE_CORDON {
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
	}
	if cluster.ClusterStatus == grpc_connectivity_manager_go.ClusterStatus_OFFLINE {
		if time.Now().Unix() - cluster.LastAliveTimestamp > cluster.GracePeriod {
			log.Debug().Msg("transitioning cluster from offline to offline cordon")
			m.config.OfflinePolicy = grpc_connectivity_manager_go.OfflinePolicy_DRAIN
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

			// Trigger Offline policy
			switch m.config.OfflinePolicy {
			case grpc_connectivity_manager_go.OfflinePolicy_NONE:
				log.Debug().Str("offline policy", m.config.OfflinePolicy.String()).Msg("offline policy set to none, no additional steps required")
			case grpc_connectivity_manager_go.OfflinePolicy_DRAIN:
				triggerOfflinePolicy(cluster, m)
			default:
				log.Debug().Msg("offline policy not set, doing nothing")
			}

		}
	}
}

func triggerOfflinePolicy (cluster *grpc_infrastructure_go.Cluster, m *Manager) {
	drainClusterRequest := &grpc_conductor_go.DrainClusterRequest{
		ClusterId: &grpc_infrastructure_go.ClusterId{
			OrganizationId: cluster.OrganizationId,
			ClusterId:      cluster.ClusterId,
		},
		ClusterOffline:       true,
	}
	drainCtx, drainCancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer drainCancel()
	drainErr := m.InfrastructureOpsProducer.Send(drainCtx, drainClusterRequest)
	if drainErr != nil {
		log.Error().Interface("send drain cluster request", drainClusterRequest).Str("trace", drainErr.Error()).Msg("unable to send drain cluster request")
	}
	log.Debug().Str("cluster id", cluster.ClusterId).Str("organization id", cluster.OrganizationId).Msg("drain cluster request sent to the bus")
}
