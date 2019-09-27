/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package server

import (
	"github.com/nalej/derrors"
	"github.com/rs/zerolog/log"
	"github.com/nalej/connectivity-manager/version"
	"time"
)

type Config struct {
	// incoming port
	Port uint32
	// Debugging flag
	Debug bool
	// URL for the message queue
	QueueAddress string
	// Grace Period
	GracePeriod time.Duration
	// Threshold
	Threshold time.Duration
}

func (conf *Config) Validate() derrors.Error {
	if conf.Port == 0 {
		return derrors.NewInvalidArgumentError("port must be set")
	}
	if conf.QueueAddress == "" {
		return derrors.NewInvalidArgumentError("queue address must be set")
	}
	if conf.GracePeriod < conf.Threshold {
		return derrors.NewInvalidArgumentError("threshold can't be longer than gracePeriod")
	}

	return nil
}

func (conf * Config) Print() {
	log.Info().Str("app", version.AppVersion).Str("commit", version.Commit).Msg("Version")
	log.Info().Uint32("port", conf.Port).Msg("gRPC port")
	log.Info().Dur("gracePeriod", conf.GracePeriod).Msg("Grace Period")
	log.Info().Dur("threshold", conf.Threshold).Msg("Threshold")
}