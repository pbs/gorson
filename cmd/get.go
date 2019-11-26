package cmd

import (
	"fmt"

	"github.com/pbs/gorson/internal/gorson/io"
	"github.com/pbs/gorson/internal/gorson/json"
	"github.com/pbs/gorson/internal/gorson/util"
	"github.com/spf13/cobra"
)

func get(path string) {
	p := util.NewParameterStorePath(path)
	pms := io.ReadFromParameterStore(*p)
	marshalled := json.Marshal(pms)
	fmt.Println(marshalled)
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
