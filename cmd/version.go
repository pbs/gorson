package cmd

import (
	"fmt"

	"github.com/pbs/gorson/internal/gorson/version"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Get gorson version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version.Version)
		},
	}
	rootCmd.AddCommand(cmd)
}
