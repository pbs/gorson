package bash

import "testing"

type testpair struct {
	input    map[string]string
	expected string
}

var testpairs = []testpair{
	{
		input:    map[string]string{"alpha": "the_alpha_value"},
		expected: "export alpha='the_alpha_value'",
	},
	{
		input: map[string]string{
			"alpha": "the_alpha_value",
			"beta":  "the_beta_value",
			"delta": "the_delta_value",
		},
		expected: `export alpha='the_alpha_value'
export beta='the_beta_value'
export delta='the_delta_value'`,
	},
	{
		input:    map[string]string{"alpha": "with&ampersand"},
		expected: "export alpha='with&ampersand'",
	},
	{
		input:    map[string]string{"colon": "with colon:"},
		expected: `export colon='with colon:'`,
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
