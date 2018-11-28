package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/pbs/gorson/internal/gorson/bash"
	"github.com/pbs/gorson/internal/gorson/io"
)

func init() {
	cmd := &cobra.Command{
		Use: "load ./example.json",
		Short: `reads a json file of key/value pairs, outputs a bash script to export them
			to set in shell, $(gorson load ./example.json)`,
		Run: func(cmd *cobra.Command, args []string) {
			filename := args[0]
			pms := io.ReadJSONFile(filename)
			output := bash.ParamsToShell(pms)
			fmt.Println(output)
		},
		Args: cobra.ExactArgs(1),
	}
	rootCmd.AddCommand(cmd)
}
