package bash

import (
	"fmt"
	"github.com/pbs/gorson/internal/gorson/util"
	"log"
	"strings"
)

// ParamsToShell generates a shell script to export environment variables
func ParamsToShell(parameters map[string]string) string {
	lines, err := util.ParametersToSlice(parameters)
	if err != nil {
		log.Printf("ParametersToSlice returned %s. Check validity of output for shell.", err.Error())
	}
	expLines := make([]string, len(lines))
	for i, line := range lines {
		expLine := fmt.Sprintf("export %s", line)
		expLines[i] = expLine
	}
	return strings.Join(expLines, "\n")
}
