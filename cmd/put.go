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
var delete bool

func put(path string, parameters map[string]string, timeout string, delete bool, autoApprove bool) {
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
	if delete {
		_, err = io.DeleteDeltaFromParameterStore(parameters, *p, autoApprove, nil)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func init() {
	cmd := &cobra.Command{
		Use:   "put /a/parameter/store/path --file /path/to/a/file",
		Short: "write parameters to a parameter store path",
		Run: func(cmd *cobra.Command, args []string) {
			path := args[0]
			parameters := io.ReadJSONFile(filename)
			put(path, parameters, timeout, delete, autoApprove)
		},
		Args: cobra.ExactArgs(1),
	}
	cmd.Flags().StringVarP(&filename, "file", "f", "", "json file to read key/value pairs from")
	cmd.Flags().StringVarP(&timeout, "timeout", "t", "1", "timeout in minutes for put")
	cmd.Flags().BoolVarP(&delete, "delete", "d", false, "deletes parameters that are not present in the json file")
	err := cmd.MarkFlagRequired("file")
	if err != nil {
		log.Fatal(err)
	}
	rootCmd.AddCommand(cmd)
}
