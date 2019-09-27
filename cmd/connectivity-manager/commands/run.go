/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package commands

import (
	"github.com/nalej/connectivity-manager/pkg/server"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"time"
)

var config = server.Config{}

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
	runCmd.Flags().StringVar(&config.QueueAddress, "queueAddress", "","address of the nalej bus")
	runCmd.Flags().DurationVar(&config.GracePeriod, "gracePeriod", 5*time.Minute, "grace period for a cluster to be considered OfflineCordon")
	runCmd.Flags().DurationVar(&config.Threshold, "threshold", time.Minute, "threshold for a cluster to be considered Offline or Online")

	rootCmd.AddCommand(runCmd)
}

func RunConnectivityManager() {
	log.Info().Msg("Launching connectivity-manager!")
	server, err := server.NewService(&config)
	if err != nil {
		log.Fatal().Err(err).Msg("error creating connectivity-manager")
	}
	config.Print()
	server.Run()
}