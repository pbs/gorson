package util

import "testing"

type testpair struct {
	input    string
	expected string
}

var testpairs = []testpair{
	{
		input:    "/EXAMPLE/NAMESPACE/",
		expected: "/EXAMPLE/NAMESPACE/",
	},
	{
		input:    "EXAMPLE/NAMESPACE/",
		expected: "/EXAMPLE/NAMESPACE/",
	},
	{
		input:    "EXAMPLE/NAMESPACE",
		expected: "/EXAMPLE/NAMESPACE/",
	},
}

func TestParameterStorePath(t *testing.T) {
	for _, pair := range testpairs {
		p := NewParameterStorePath(pair.input)
		output := p.String()
		if output != pair.expected {
			t.Error(
				"For", pair.input,
				"expected", pair.expected,
				"got", output,
			)
		}
	}
}
