package cmd

import (
	"fmt"

	"github.com/pbs/gorson/internal/gorson/bash"
	"github.com/pbs/gorson/internal/gorson/io"
	"github.com/spf13/cobra"
)

// for alternative to global flag vars, see https://github.com/spf13/cobra/issues/1599#issuecomment-1040738878
var loadFormat string

func load(filename string) {
	if loadFormat == "env" {
		pms := io.(filename)
	}
	pms := io.ReadJSONFile(filename)
	output := bash.ParamsToShell(pms)
	fmt.Println(output)
}

func init() {
	cmd := &cobra.Command{
		Use:     "load ./example.json",
		Short:   `reads a file of key/value pairs and outputs a bash script to export them in a shell`,
		Example: `source <(gorson load ./example.json); source <(gorson load /etc/app/.env -f env)`,
		Run: func(cmd *cobra.Command, args []string) {
			load(args[0])
		},
		Args: cobra.ExactArgs(1),
	}
	cmd.Flags().StringVarP(&format, "format", "f", "json", "the format of the file that is read (one of env, yaml, json [default])")
	rootCmd.AddCommand(cmd)
}
