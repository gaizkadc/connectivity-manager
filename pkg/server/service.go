/*
 * Copyright (C) 2018 Nalej Group - All Rights Reserved
 *
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
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

const (
	InfrastructureEventsConsumerName = "ApplicationManager-app_ops"
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
		server:             server,
		configuration:      config,
	}

	return &instance, nil
}

type Clients struct {
	ClusterClient grpc_infrastructure_go.ClustersClient
	OrgClient grpc_organization_go.OrganizationsClient
}

type BusClients struct {
	InfrastructureEventsConsumer *events.InfrastructureEventsConsumer
}

// GetClients creates the required connections with the remote clients.
func (s*Service) GetClients() (*Clients, derrors.Error) {
	smConn, err := grpc.Dial(s.configuration.SystemModelAddress, grpc.WithInsecure())
	if err != nil{
		return nil, derrors.AsError(err, "cannot create connection with the system model component")
	}

	clClient := grpc_infrastructure_go.NewClustersClient(smConn)
	orgClient := grpc_organization_go.NewOrganizationsClient(smConn)

	return &Clients{ClusterClient:clClient, OrgClient:orgClient}, nil
}

// GetBusClients creates the required connections with the bus
func (s*Service) GetBusClients() (*BusClients, derrors.Error) {
	queueClient := pulsar_comcast.NewClient(s.configuration.QueueAddress, nil)

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

	return &BusClients{
		InfrastructureEventsConsumer:infraEventsConsumer,
	}, nil
}

func(s *Service) Run() {
	s.configuration.Validate()
	s.configuration.Print()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.configuration.Port))
	if err != nil {
		log.Fatal().Errs("failed to listen: %v", []error{err})
	}

	clients, cErr := s.GetClients()
	if cErr != nil{
		log.Fatal().Str("err", cErr.DebugReport()).Msg("Cannot create clients")
	}

	busClients, bErr := s.GetBusClients()
	if bErr != nil{
		log.Fatal().Str("err", bErr.DebugReport()).Msg("Cannot create bus clients")
	}

	connectivityManagerManager, err := connectivity_manager.NewManager(&clients.ClusterClient, &clients.OrgClient, busClients.InfrastructureEventsConsumer, *s.configuration)
	if err != nil{
		log.Fatal().Str("err", err.Error()).Msg("Cannot create connectivity-manager manager")
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
