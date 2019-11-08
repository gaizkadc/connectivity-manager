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

package config

import (
	"github.com/nalej/connectivity-manager/version"
	"github.com/nalej/derrors"
	grpc_connectivity_manager_go "github.com/nalej/grpc-connectivity-manager-go"
	"github.com/rs/zerolog/log"
	"time"
)

type Config struct {
	// incoming port
	Port uint32
	// Debugging flag
	Debug bool
	// SystemModelAddress with the host:port to connect to System Model
	SystemModelAddress string
	// URL for the message queue
	QueueAddress string
	// Threshold
	Threshold time.Duration
	// Offline Policy must be set to true when a cluster is offline thus an offline policy should be triggered
	OfflinePolicy grpc_connectivity_manager_go.OfflinePolicy
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

func (conf *Config) Print() {
	log.Info().Str("app", version.AppVersion).Str("commit", version.Commit).Msg("Version")
	log.Info().Uint32("port", conf.Port).Msg("gRPC port")
	log.Info().Str("URL", conf.SystemModelAddress).Msg("System Model")
	log.Info().Dur("threshold", conf.Threshold).Msg("Threshold")
	log.Info().Str("offline policy", conf.OfflinePolicy.String()).Msg("Offline policy")
}
