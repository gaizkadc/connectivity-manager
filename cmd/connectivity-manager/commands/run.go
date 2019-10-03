/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package commands

import (
	"github.com/nalej/connectivity-manager/pkg/server"
	config2 "github.com/nalej/connectivity-manager/pkg/server/config"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"time"
)

var config = config2.Config{}

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
	runCmd.Flags().StringVar(&config.QueueAddress, "queueAddress", "","address of the nalej bus")
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