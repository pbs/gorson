package cmd

import (
	"log"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	noColor     bool
	autoApprove bool
	rootCmd     = &cobra.Command{
		Use:   "gorson",
		Short: "get/put parameters to/from AWS parameter store, load them as environment variables",
	}
)

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "deactivate color usage")
	rootCmd.PersistentFlags().BoolVar(&autoApprove, "auto-approve", false, "automatically approve any prompt")
}

func initConfig() {
	color.NoColor = noColor // disables colorized output
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
