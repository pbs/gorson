package util

import (
	"golang.org/x/exp/slices"
	"testing"
)

var testcases = []struct {
	input    map[string]string
	expected []string
}{
	{
		map[string]string{"lucy": "football", "linus": "blanket", "schroeder": "piano"},
		[]string{"lucy='football'", "linus='blanket'", "schroeder='piano'"},
	},
	{
		map[string]string{"0ucy": "football", "linus": "blanket", "schroeder": "piano"},
		nil,
	},
	{
		map[string]string{"lucy": "footb'l'", "linus": "b:anket", "schroeder": "p i	ano"},
		[]string{`lucy='footb'\''l'\'''`, "linus='b:anket'", "schroeder='p i	ano'"},
	},
	{
		map[string]string{"complex": "a$b'c-d_e;f&g\"h|i@j"},
		[]string{`complex='a$b'\''c-d_e;f&g"h|i@j'`},
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

func TestParametersToSlice(t *testing.T) {
	for i, tc := range testcases {
		arr, err := ParametersToSlice(tc.input)
		if err != nil {
			if tc.expected != nil {
				t.Errorf("Function returned an error, but this was not expected")
			}
		}
		slices.Sort(arr)
		slices.Sort(tc.expected)
		if slices.CompareFunc(tc.expected, arr, func(str1 string, str2 string) int {
			retval := 0
			if str1 != str2 {
				t.Errorf("This is arr: %s; %s does not equal %s", arr, str1, str2)
				retval = -1
			}
			return retval
		}) != 0 {
			t.Errorf("Test case %d failed", i)
		}
	}
}
