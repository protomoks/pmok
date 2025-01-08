/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	gopath "path"
	"strings"

	_ "embed"

	"github.com/protomoks/pmok/internal/config"
	"github.com/protomoks/pmok/internal/functions"
	"github.com/protomoks/pmok/internal/ux"
	"github.com/spf13/cobra"
)

var (
	path string
)

//go:embed handlertemplate.ts
var template []byte

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a function handler",
	Long:  `Add a function handler`,
	Run: func(cmd *cobra.Command, args []string) {
		dir, err := config.HasProject()
		if err != nil || dir == "" {
			log.Fatal("Unable to add a mock handler. Did you create your project with 'pmok create' yet?")
		}
		fname, err := functions.PathPatternToFileName(path)
		if err != nil {
			log.Fatal(err)
		}
		f, err := os.Create(gopath.Join(dir, config.FunctionsDir, fname))
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		r := strings.NewReader(string(template))
		if _, err := io.Copy(f, r); err != nil {
			log.Fatal(err)
		}

		s := ux.DefaultStyleRenderer()
		fmt.Printf("Created %s\n", s.SuccessText.Render(fname))

	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().StringVarP(&path, "path", "p", "", "the url path pattern")
	addCmd.MarkFlagRequired("path")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// addCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
