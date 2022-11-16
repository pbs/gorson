package json

import "testing"

type testpair struct {
	input    map[string]string
	expected string
}

/* whitespace is important in this array definition */
var testpairs = []testpair{
	{
		input: map[string]string{"WITH_AMPERSAND": "with&ampersand"},
		expected: `{
    "WITH_AMPERSAND": "with&ampersand"
}
`},
	{
		input: map[string]string{"WITH SPACE": "with space"},
		expected: `{
    "WITH SPACE": "with space"
}
`},
	{
		input: map[string]string{"WITH_COLON": "with_colon:"},
		expected: `{
    "WITH_COLON": "with_colon:"
}
`},
}

func TestParamsToJson(t *testing.T) {
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
