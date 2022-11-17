package bash

import (
	"fmt"
	"github.com/pbs/gorson/internal/gorson/util"
	"strings"
)

// ParamsToShell generates a shell script to export environment variables
func ParamsToShell(parameters map[string]string) string {
	lines := util.ParmsToArray(parameters)
	expLines := make([]string, len(lines))
	for i, line := range lines {
		expLine := fmt.Sprintf("export %s", line)
		expLines[i] = expLine
	}
	return strings.Join(expLines, "\n")
}
