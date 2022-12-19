package env

import (
	"log"
	"strings"

	"github.com/pbs/gorson/internal/gorson/util"
)

// Marshal env-formats parameters
func Marshal(parameters map[string]string) string {
	lines, err := util.ParametersToSlice(parameters)
	if err != nil {
		log.Printf("ParametersToSlice returned %s. Check validity of output for shell.", err.Error())
	}
	return strings.Join(lines, "\n")
}

func Deserialize()
