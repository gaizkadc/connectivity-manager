/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package commands

import (
	"github.com/nalej/connectivity-manager/pkg/server"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

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
	runCmd.Flags().Uint32("port", 8383,"port where connectivity-manager listens to")
	runCmd.Flags().String("queueAddress", "", "address of the nalej bus")
	rootCmd.AddCommand(runCmd)
}

func RunConnectivityManager() {
	var port uint32
	var debug bool
	var queueAddress string

	port = uint32(viper.GetInt32("port"))
	debug = viper.GetBool("debug")
	queueAddress = viper.GetString("queueAddress")

	config := server.Config{
		Port:  port,
		Debug: debug,
		QueueAddress: queueAddress,
	}

	log.Info().Msg("Launching connectivity-manager!")
	server, err := server.NewService(&config)
	if err != nil {
		log.Fatal().Err(err).Msg("error creating connectivity-manager")
	}
	server.Run()
}