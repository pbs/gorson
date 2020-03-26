package env

import (
	"log"
	"regexp"

	"gopkg.in/yaml.v2"
)

// Format env-formats parameters
func Format(parameters map[string]string) string {
	serialized, err := yaml.Marshal(parameters)
	if err != nil {
		log.Fatal(err)
	}
	var re = regexp.MustCompile(`(\w*):\s"?([^\n"]*)"?`)
	return re.ReplaceAllString(string(serialized), "${1}='${2}'")
}
