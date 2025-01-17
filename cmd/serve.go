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
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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
