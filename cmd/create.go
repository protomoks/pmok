/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/protomoks/pmok/internal/ux"
	"github.com/spf13/cobra"
)

type noopLogger struct{}

func (w *noopLogger) Write(b []byte) (n int, err error) {
	return 0, nil
}

// initCmd represents the init command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Creates a new protomok project",
	Long:  `Creates a new protomok project in your current working directory`,
	Run: func(cmd *cobra.Command, args []string) {
		var logger io.Writer
		if len(os.Getenv("DEBUG")) > 0 {
			f, err := tea.LogToFile("debug.log", "debug")
			if err != nil {
				fmt.Println("fatal:", err)
				os.Exit(1)
			}
			logger = f
			defer f.Close()
		} else {
			logger = &noopLogger{}
		}

		state := ux.NewCreateProjectModel(logger)
		if _, err := tea.NewProgram(state).Run(); err != nil {
			fmt.Printf("There was an error %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
}
