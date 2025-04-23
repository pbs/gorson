package util

import (
	"fmt"
	"golang.org/x/exp/maps"
	"regexp"
	"sort"
	"strings"
)

type ParameterStorePath struct {
	components []string
}

func (p ParameterStorePath) String() string {
	output := "/" + strings.Join(p.components, "/") + "/"
	return output
}

func NewParameterStorePath(input string) *ParameterStorePath {
	split := strings.Split(input, "/")
	var filtered []string
	for _, str := range split {
		if str != "" {
			filtered = append(filtered, str)
		}
	}
	return &ParameterStorePath{filtered}
}

// ParametersToSlice accepts a map of string key/value pairs and returns an array of strings.
// The elements of the returned array are keys and values conjoined by an `=` sign.
// Keys that do not start with a letter or underscore will result in an error.
// All values are enclosed in single quotes.
// Values that contain single quote characters will first be converted to double quotes.
func ParametersToSlice(parameters map[string]string) ([]string, error) {
	lines := make([]string, 0)
	keys := maps.Keys(parameters)
	sort.Strings(keys)
	//envvar names must begin with a letter or underscore
	invalidKey := regexp.MustCompile(`^[^a-zA-Z_]`)
	for _, key := range keys {
		if invalidKey.MatchString(key) {
			return nil, fmt.Errorf("Key %s invalid", key)
		}
		v := parameters[key]
		if strings.Contains(v, "'") {
			v = strings.ReplaceAll(v, "'", "'\\''")
		}
		line := fmt.Sprintf("%s='%s'", key, v)
		lines = append(lines, line)
	}
	return lines, nil
}
