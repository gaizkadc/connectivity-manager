/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package server

import (
	"github.com/nalej/derrors"
	"github.com/rs/zerolog/log"
	"github.com/nalej/connectivity-manager/version"
)

type Config struct {
	// incoming port
	Port uint32
	// Debugging flag
	Debug bool
	// URL for the message queue
	QueueAddress string
}

func (conf *Config) Validate() derrors.Error {
	if conf.Port == 0 {
		return derrors.NewInvalidArgumentError("port must be set")
	}
	if conf.QueueAddress == "" {
		return derrors.NewInvalidArgumentError("queue address must be set")
	}

	return nil
}

func (conf * Config) Print() {
	log.Info().Str("app", version.AppVersion).Str("commit", version.Commit).Msg("Version")
	log.Info().Uint32("port", conf.Port).Msg("gRPC port")
}