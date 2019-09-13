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
	},
}

func init() {
	runCmd.Flags().Uint32P("port", "p", 8383,"port where connectivity-manager listens to")
	rootCmd.AddCommand(runCmd)
}

func RunConnectivityManager() {
	// Incoming requests port
	var port uint32
	// Debug flag
	var debug bool

	port = uint32(viper.GetInt32("port"))
	debug = viper.GetBool("debug")

	config := server.Config{
		Port:  port,
		Debug: debug,
	}

	log.Info().Msg("Launching connectivity-manager!")
	server, _ := server.NewService(&config)
	server.Run()
}