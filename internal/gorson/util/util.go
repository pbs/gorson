package util

import "strings"

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

func GetKeys(parameters map[string]string) []string {
	i := 0
	keys := make([]string, len(parameters))
	for k := range parameters {
		keys[i] = k
		i++
func ParmsToArray(parameters map[string]string) []string {
	lines := make([]string, 0)
	keys := maps.Keys(parameters)
	sort.Strings(keys)
	for _, key := range keys {
		v := parameters[key]
		line := fmt.Sprintf("%s='%s'", key, v)
		lines = append(lines, line)
	}
	return lines
}
