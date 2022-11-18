package util

import "testing"

var testcases = []struct {
	input    map[string]string
	expected []string
}{
	{
		map[string]string{"lucy": "football", "linus": "blanket", "schroeder": "piano"},
		[]string{"lucy='football'", "linus='blanket'", "schroeder='piano'"},
	},
}

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

func TestParmsToArray(t *testing.T) {
	for _, m := range testcases {
		arr := ParmsToArray(m)

	}
}
