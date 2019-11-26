package cmd

import (
	"fmt"
	"log"

	"github.com/pbs/gorson/internal/gorson/io"
	"github.com/pbs/gorson/internal/gorson/json"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func get(path string, y bool) {
	pms := io.ReadFromParameterStore(path)
	if y {
		serialized, err := yaml.Marshal(pms)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(serialized))

	} else {
		marshalled := json.Marshal(pms)
		fmt.Println(marshalled)
	}
}

func init() {
	cmd := &cobra.Command{
		Use:   "get /a/parameter/store/path",
		Short: "Get parameters from a parameter store path",
		Run: func(cmd *cobra.Command, args []string) {
			path := args[0]
			y, _ := cmd.Flags().GetBool("yaml")
			get(path, y)
		},
		Args: cobra.ExactArgs(1),
	}
	cmd.Flags().Bool("yaml", false, "get command outputs as yaml")
	rootCmd.AddCommand(cmd)
}
