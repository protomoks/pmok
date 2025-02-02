/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

	"github.com/protomoks/pmok/internal/config"
	"github.com/protomoks/pmok/internal/functions/serve"
	"github.com/protomoks/pmok/internal/functions/serve/docker"
	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start your mock server locally",
	Run: func(cmd *cobra.Command, args []string) {
		conf := config.GetConfig()
		if conf == nil {
			log.Fatal("Unable to read your project. Did you initialize a project already in this directory?")
		}

		cm, err := docker.NewContainerManager()
		if err != nil {
			log.Fatalf("Error %s\n", err)
		}
		if err := serve.Run(cmd.Context(), cm); err != nil {
			log.Fatalf("Error %s\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
