package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var region string

var rootCmd = &cobra.Command{
	Use:   "gorson",
	Short: "get/put parameters to/from AWS parameter store, load them as environment variables",
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&region, "region", "r", "us-east-1", "AWS region")
	viper.BindPFlag("region", rootCmd.PersistentFlags().Lookup("region"))
	viper.SetDefault("region", "us-east-1")
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
