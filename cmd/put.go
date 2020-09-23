package cmd

import (
	"log"
	"strconv"
	"time"

	"github.com/pbs/gorson/internal/gorson/io"
	"github.com/pbs/gorson/internal/gorson/util"
	"github.com/spf13/cobra"
)

var filename string
var timeout string

func put(path string, parameters map[string]string, timeout string) {
	p := util.NewParameterStorePath(path)
	timeoutInt, err := strconv.ParseInt(timeout, 0, 64)
	timeoutDuration := time.Duration(timeoutInt) * time.Minute
	if err != nil {
		log.Fatal(err)
	}
	err = io.WriteToParameterStore(parameters, *p, timeoutDuration, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	cmd := &cobra.Command{
		Use:   "put /a/parameter/store/path --file /path/to/a/file",
		Short: "write parameters to a parameter store path",
		Run: func(cmd *cobra.Command, args []string) {
			path := args[0]
			parameters := io.ReadJSONFile(filename)
			put(path, parameters, timeout)
		},
		Args: cobra.ExactArgs(1),
	}
	cmd.Flags().StringVarP(&filename, "file", "f", "", "json file to read key/value pairs from")
	cmd.Flags().StringVarP(&timeout, "timeout", "t", "1", "timeout in minutes for put")
	cmd.MarkFlagRequired("file")
	rootCmd.AddCommand(cmd)
}
