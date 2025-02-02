/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

	"github.com/protomoks/pmok/internal/recorder"
	"github.com/spf13/cobra"
)

var (
	target   string
	mockpath string
)

// recordCmd represents the record command
var recordCmd = &cobra.Command{
	Use:   "record",
	Short: "Start a proxy to record requests and responses",
	Run: func(cmd *cobra.Command, args []string) {
		if err := recorder.Run(cmd.Context(), recorder.RecordCommand{
			Target:        target,
			ResponsesPath: mockpath,
		}); err != nil {
			log.Fatalln(err)
		}

	},
}

func init() {
	rootCmd.AddCommand(recordCmd)
	recordCmd.Flags().StringVarP(&target, "target", "t", "", "The target url you want to record responses for")
	recordCmd.Flags().StringVarP(&mockpath, "path", "p", "", "If provided, mocks are stored in this path")

	recordCmd.MarkFlagRequired("target")

}
