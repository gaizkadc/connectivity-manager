/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package server

import (
	"github.com/nalej/derrors"
	"github.com/rs/zerolog/log"
)

type Config struct {
	// incoming port
	Port uint32
	// Debugging flag
	Debug bool

}

func (conf *Config) Validate() derrors.Error {

	return nil
}

func (conf * Config) Print() {
	log.Info().Uint32("port", conf.Port).Msg("gRPC port")
}