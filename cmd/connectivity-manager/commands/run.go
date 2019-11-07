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

package commands

import (
	"github.com/nalej/connectivity-manager/pkg/server"
	cmConfig "github.com/nalej/connectivity-manager/pkg/server/config"
	grpc_connectivity_manager_go "github.com/nalej/grpc-connectivity-manager-go"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"strings"
	"time"
)

var config = cmConfig.Config{}

var policyName string

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run connectivity-manager",
	Long:  `Run connectivity-manager`,
	Run: func(cmd *cobra.Command, args []string) {
		SetupLogging()
		RunConnectivityManager()
	},
}

func init() {
	runCmd.Flags().Uint32Var(&config.Port, "port", 8383, "port where connectivity-manager listens to")
	runCmd.PersistentFlags().StringVar(&config.SystemModelAddress, "systemModelAddress", "localhost:8800",
		"System Model address (host:port)")
	runCmd.Flags().StringVar(&config.QueueAddress, "queueAddress", "", "address of the nalej bus")
	runCmd.Flags().DurationVar(&config.Threshold, "threshold", time.Minute, "threshold for a cluster to be considered Offline or Online")
	runCmd.Flags().StringVar(&policyName, "offlinePolicy", "none", "Offline policy to trigger when cordoning an offline cluster: none or drain")

	rootCmd.AddCommand(runCmd)
}

func RunConnectivityManager() {

	policy, exists := grpc_connectivity_manager_go.OfflinePolicy_value[strings.ToUpper(policyName)]
	if !exists {
		log.Fatal().Msg("invalid offline policy set")
	}
	config.OfflinePolicy = grpc_connectivity_manager_go.OfflinePolicy(policy)

	log.Info().Msg("Launching connectivity-manager!")
	server, err := server.NewService(&config)
	if err != nil {
		log.Fatal().Err(err).Msg("error creating connectivity-manager")
	}
	server.Run()
}
