package cmd

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"projects.pbs.org/bitbucket/users/cmacdonald/repos/gorson/internal/gorson/io"
)

func get(path string, region string) {
	pms := io.ReadFromParameterStore(path, region)
	serialized, err := json.MarshalIndent(pms, "", "    ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(serialized))
}

func init() {
	cmd := &cobra.Command{
		Use:   "get /a/parameter/store/path",
		Short: "Get parameters from a parameter store path",
		Run: func(cmd *cobra.Command, args []string) {
			path := args[0]
			get(path, region)
		},
		Args: cobra.ExactArgs(1),
	}
	rootCmd.AddCommand(cmd)
}
