package cmd

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/pbs/gorson/internal/gorson/io"
	"github.com/spf13/cobra"
)

func get(path string) {
	pms := io.ReadFromParameterStore(path)
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
			get(path)
		},
		Args: cobra.ExactArgs(1),
	}
	rootCmd.AddCommand(cmd)
}
