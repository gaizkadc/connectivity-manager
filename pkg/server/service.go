/*
 * Copyright (C) 2018 Nalej Group - All Rights Reserved
 *
 */


package server

import (
	"google.golang.org/grpc"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/reflection"
	"net"
	"fmt"
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


func(c *Service) Run() {

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", c.configuration.Port))
	if err != nil {
		log.Fatal().Errs("failed to listen: %v", []error{err})
	}

	// Register reflection service on gRPC server.
	if c.configuration.Debug {
		reflection.Register(c.server)
	}

	// Run
	log.Info().Uint32("port", c.configuration.Port).Msg("Launching gRPC server")
	if err := c.server.Serve(lis); err != nil {
		log.Fatal().Errs("failed to serve: %v", []error{err})
	}

}
