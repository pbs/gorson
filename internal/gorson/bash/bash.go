package bash

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pbs/gorson/internal/gorson/util"
)

// ParamsToShell generates a shell script to export environment variables
func ParamsToShell(parameters map[string]string) string {
	lines := make([]string, 0)
	keys := util.GetKeys(parameters)
	sort.Strings(keys)
	for _, key := range keys {
		v := parameters[key]
		line := fmt.Sprintf("export %s='%s'", key, v)
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}
