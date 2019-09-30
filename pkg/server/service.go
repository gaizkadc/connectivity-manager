/*
 * Copyright (C) 2018 Nalej Group - All Rights Reserved
 *
 */


package server

import (
	connectivity_manager "github.com/nalej/connectivity-manager/pkg/server/connectivity-manager"
	"github.com/nalej/derrors"
	pulsar_comcast "github.com/nalej/nalej-bus/pkg/bus/pulsar-comcast"
	"github.com/nalej/nalej-bus/pkg/queue/infrastructure/events"
	"github.com/nalej/connectivity-manager/pkg/queue"
	"google.golang.org/grpc"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/reflection"
	"net"
	"fmt"
)

const (
	InfrastructureEventsConsumerName = "ApplicationManager-app_ops"
)

type Service struct {
	// Server for incoming requests
	server *grpc.Server
	// Configuration object
	configuration *Config
}

func NewService(config *Config) (*Service, error) {
	server := grpc.NewServer()
	instance := Service{
		server:             server,
		configuration:      config,
	}

	return &instance, nil
}

type BusClients struct {
	InfrastructureEventsConsumer *events.InfrastructureEventsConsumer
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
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.configuration.Port))
	if err != nil {
		log.Fatal().Errs("failed to listen: %v", []error{err})
	}

	busClients, bErr := s.GetBusClients()
	if err != nil{
		log.Fatal().Str("err", bErr.DebugReport()).Msg("Cannot create bus clients")
	}

	connectivityManagerManager, err := connectivity_manager.NewManager(busClients.InfrastructureEventsConsumer)
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
