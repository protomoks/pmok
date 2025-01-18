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
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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
