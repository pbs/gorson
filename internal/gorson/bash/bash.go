package bash

import "strings"

// ParamsToShell generates a shell script to export environment variables
func ParamsToShell(parameters map[string]string) string {
	lines := make([]string, 0)
	for key, value := range parameters {
		// TODO do we need to do extra escaping here?
		lines = append(lines, "export "+key+"="+value)
	}
	return strings.Join(lines, "\n")
}
