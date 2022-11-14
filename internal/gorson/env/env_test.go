package env

import "testing"

type testpair struct {
	input    map[string]string
	expected string
}

var testpairs = []testpair{
	{
		input:    map[string]string{"WITH_AMPERSAND": "with&ampersand"},
		expected: `WITH_AMPERSAND='with&ampersand'`,
	},
	{
		input:    map[string]string{"WITH_SPACES": "with spaces"},
		expected: `WITH_SPACES='with spaces'`,
	},
	{
		input:    map[string]string{"WITH_COLON": "with colon:"},
		expected: `WITH_COLON='with colon:'`,
	},
}

func TestMarshal(t *testing.T) {
	for _, pair := range testpairs {
		output := Marshal(pair.input)
		if output != pair.expected {
			t.Error(
				"For", pair.input,
				"expected", pair.expected,
				"got", output,
			)
		}
	}
}
