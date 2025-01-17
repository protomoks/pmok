/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"

	_ "embed"

	"github.com/protomoks/pmok/internal/config"
	"github.com/protomoks/pmok/internal/functions/add"
	"github.com/protomoks/pmok/internal/ux"
	"github.com/spf13/cobra"
)

var (
	path    string
	name    string
	methods []string
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a function handler",
	Long:  `Add a function handler`,
	Run: func(cmd *cobra.Command, args []string) {
		conf := config.GetConfig()
		if conf == nil {
			log.Fatal("could not find a protomok project")
		}
		fmt.Println(conf.Manifest)
		if err := add.AddFunction(add.AddFunctionCommand{
			Name:           name,
			HttpPath:       path,
			AllowedMethods: methods,
		}); err != nil {
			log.Fatal(err)
		}

		s := ux.DefaultStyleRenderer()
		fmt.Printf("Created %s\n", s.SuccessText.Render(name))

	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().StringVarP(&path, "path", "p", "", "the url path pattern")
	addCmd.Flags().StringVarP(&name, "name", "n", "", "the name of the function")
	addCmd.Flags().StringSliceVar(&methods, "m", []string{"GET"}, "the http methods this function responds to")
	addCmd.MarkFlagRequired("path")
	addCmd.MarkFlagRequired("name")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// addCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
