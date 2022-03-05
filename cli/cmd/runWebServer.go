/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/jmoussa/go-sentitweet/api"
	"github.com/spf13/cobra"
)

// runWebServerCmd represents the runWebServer command
var runWebServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Run the sentitweet web server",
	Long:  `Starts API backend which enables queries to the results of the sentiment analysis pipeline.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting Web Server...")
		api.RunServer()
	},
}

func init() {
	rootCmd.AddCommand(runWebServerCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runWebServerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runWebServerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
