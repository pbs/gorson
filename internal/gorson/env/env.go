package env

import (
	"log"
	"regexp"

	"gopkg.in/yaml.v2"
)

// Marshal env-formats parameters
func Marshal(parameters map[string]string) string {
	serialized, err := yaml.Marshal(parameters)
	if err != nil {
		log.Fatal(err)
	}
	var re = regexp.MustCompile(`(\w*):\s"?([^\n"]*)"?`)
	return re.ReplaceAllString(string(serialized), "${1}='${2}'")
}
