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

package server

import (
	"fmt"
	"github.com/nalej/connectivity-manager/pkg/queue"
	"github.com/nalej/connectivity-manager/pkg/server/config"
	connectivity_manager "github.com/nalej/connectivity-manager/pkg/server/connectivity-manager"
	"github.com/nalej/derrors"
	grpc_infrastructure_go "github.com/nalej/grpc-infrastructure-go"
	grpc_organization_go "github.com/nalej/grpc-organization-go"
	pulsar_comcast "github.com/nalej/nalej-bus/pkg/bus/pulsar-comcast"
	"github.com/nalej/nalej-bus/pkg/queue/infrastructure/events"
	infra_ops "github.com/nalej/nalej-bus/pkg/queue/infrastructure/ops"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

const (
	InfrastructureEventsConsumerName = "ConnectivityManager-infra_events"
	InfrastructureOpsProducerName    = "ConnectivityManager-infra_ops"
)

type Service struct {
	// Server for incoming requests
	server *grpc.Server
	// Configuration object
	configuration *config.Config
}

func NewService(config *config.Config) (*Service, error) {
	server := grpc.NewServer()
	instance := Service{
		server:        server,
		configuration: config,
	}

	return &instance, nil
}

type Clients struct {
	ClusterClient grpc_infrastructure_go.ClustersClient
	OrgClient     grpc_organization_go.OrganizationsClient
}

type BusClients struct {
	InfrastructureEventsConsumer *events.InfrastructureEventsConsumer
	InfrastructureOpsProducer    *infra_ops.InfrastructureOpsProducer
}

// GetClients creates the required connections with the remote clients.
func (s *Service) GetClients() (*Clients, derrors.Error) {
	smConn, err := grpc.Dial(s.configuration.SystemModelAddress, grpc.WithInsecure())
	if err != nil {
		return nil, derrors.AsError(err, "cannot create connection with the system model component")
	}

	clClient := grpc_infrastructure_go.NewClustersClient(smConn)
	orgClient := grpc_organization_go.NewOrganizationsClient(smConn)

	return &Clients{ClusterClient: clClient, OrgClient: orgClient}, nil
}

// GetBusClients creates the required connections with the bus
func (s *Service) GetBusClients() (*BusClients, derrors.Error) {
	queueClient := pulsar_comcast.NewClient(s.configuration.QueueAddress, nil)

	// Infrastructure Events Consumer
	InfrastructureEventsConsumerStruct := events.ConsumableStructsInfrastructureEventsConsumer{
		UpdateClusterRequest:    false,
		SetClusterStatusRequest: false,
		ClusterAliveRequest:     true,
	}
	infrastructureEventConsumerConfig := events.NewConfigInfrastructureEventsConsumer(5, InfrastructureEventsConsumerStruct)
	infraEventsConsumer, err := events.NewInfrastructureEventsConsumer(queueClient, InfrastructureEventsConsumerName, true, infrastructureEventConsumerConfig)
	if err != nil {
		return nil, err
	}

	infraOpsProducer, err := infra_ops.NewInfrastructureOpsProducer(queueClient, InfrastructureOpsProducerName)
	if err != nil {
		return nil, err
	}

	return &BusClients{
		InfrastructureEventsConsumer: infraEventsConsumer,
		InfrastructureOpsProducer:    infraOpsProducer,
	}, nil
}

func (s *Service) Run() {
	vErr := s.configuration.Validate()
	if vErr != nil {
		log.Fatal().Str("err", vErr.DebugReport()).Msg("invalid configuration")
	}
	s.configuration.Print()

	lis, lErr := net.Listen("tcp", fmt.Sprintf(":%d", s.configuration.Port))
	if lErr != nil {
		log.Fatal().Errs("failed to listen: %v", []error{lErr})
	}

	clients, cErr := s.GetClients()
	if cErr != nil {
		log.Fatal().Str("err", cErr.DebugReport()).Msg("Cannot create clients")
	}

	busClients, bErr := s.GetBusClients()
	if bErr != nil {
		log.Fatal().Str("err", bErr.DebugReport()).Msg("Cannot create bus clients")
	}

	connectivityManagerManager, nmErr := connectivity_manager.NewManager(
		&clients.ClusterClient,
		&clients.OrgClient,
		busClients.InfrastructureEventsConsumer,
		busClients.InfrastructureOpsProducer,
		*s.configuration)
	if nmErr != nil {
		log.Fatal().Str("err", nmErr.Error()).Msg("Cannot create connectivity-manager manager")
	}

	infraEventsHandler := queue.NewInfrastructureEventsHandler(connectivityManagerManager, busClients.InfrastructureEventsConsumer)
	infraEventsHandler.Run(s.configuration.Threshold)

	// Register reflection service on gRPC server
	if s.configuration.Debug {
		reflection.Register(s.server)
	}

	// Run
	log.Info().Uint32("port", s.configuration.Port).Msg("Launching gRPC server")
	if err := s.server.Serve(lis); err != nil {
		log.Fatal().Errs("failed to serve: %v", []error{err})
	}

}
