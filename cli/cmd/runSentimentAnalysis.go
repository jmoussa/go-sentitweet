/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	data_pipelines "github.com/jmoussa/go-sentitweet/data-pipelines"
	"github.com/spf13/cobra"
)

// runSentimentAnalysisCmd represents the runSentimentAnalysis command
var runSentimentAnalysisCmd = &cobra.Command{
	Use:   "pipeline",
	Short: "Run the Sentiment Analysis Pipeline with default search phrase: #nft",
	Long: `Will run the sentiment analysis pipeline, that saves the results to the database.
	Runs in foreground.`,
	Run: func(cmd *cobra.Command, args []string) {
		searchTerm, _ := cmd.Flags().GetString("term")
		fmt.Println("Sentiment Analysis Pipeline Starting for: ", searchTerm)
		data_pipelines.RunTwitterPipeline(searchTerm)
	},
}

func init() {
	rootCmd.AddCommand(runSentimentAnalysisCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	runSentimentAnalysisCmd.PersistentFlags().String("term", "", "Search term to filter tweets for the pipeline (default:'#nft'")
	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runSentimentAnalysisCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
