package env

import (
	"strings"

	"github.com/pbs/gorson/internal/gorson/util"
)

// Marshal env-formats parameters
func Marshal(parameters map[string]string) string {
	lines := util.ParmsToArray(parameters)
	return strings.Join(lines, "\n")
}
