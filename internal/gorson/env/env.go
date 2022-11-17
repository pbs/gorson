package env

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pbs/gorson/internal/gorson/util"
)

// Marshal env-formats parameters
func Marshal(parameters map[string]string) string {
	lines := make([]string, 0)
	keys := util.GetKeys(parameters)
	sort.Strings(keys)
	for _, key := range keys {
		v := parameters[key]
		line := fmt.Sprintf("%s='%s'", key, v)
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}
