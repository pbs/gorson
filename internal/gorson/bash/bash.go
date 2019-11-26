package bash

import (
	"fmt"
	"sort"
	"strings"
)

func getKeys(parameters map[string]string) []string {
	i := 0
	keys := make([]string, len(parameters))
	for k := range parameters {
		keys[i] = k
		i++
	}
	return keys
}

// ParamsToShell generates a shell script to export environment variables
func ParamsToShell(parameters map[string]string) string {
	lines := make([]string, 0)
	keys := getKeys(parameters)
	sort.Strings(keys)
	for _, key := range keys {
		v := parameters[key]
		line := fmt.Sprintf("export %s='%s'", key, v)
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}
