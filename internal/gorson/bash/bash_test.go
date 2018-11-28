package bash

import "testing"

type testpair struct {
	input    map[string]string
	expected string
}

var testpairs = []testpair{
	testpair{
		input:    map[string]string{"alpha": "the_alpha_value"},
		expected: "export alpha=the_alpha_value",
	},
}

func TestParamsToShell(t *testing.T) {
	for _, pair := range testpairs {
		output := ParamsToShell(pair.input)
		if output != pair.expected {
			t.Error(
				"For", pair.input,
				"expected", pair.expected,
				"got", output,
			)
		}
	}
}
