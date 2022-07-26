package turnpike

import (
	"testing"
)

func TestExpandPath(t *testing.T) {
	type testCase struct {
		name     string
		input    string
		expected []string
	}

	tests := []testCase{
		{name: "BasicPath", input: "test", expected: []string{"test"}},
		{name: "NestedPath", input: "test/path", expected: []string{"test", "path"}},
		{name: "NestedPathTrailingSlash", input: "/some/test/path/", expected: []string{"some", "test", "path"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ret := expandPath(test.input)
			if !areSlicesEqByValue(ret, test.expected) {
				t.Errorf("expected input %s to expand to %v but got %v", test.input, test.expected, ret)
			}
		})
	}
}

func TestDeriveLabelPattern(t *testing.T) {
	type testCase struct {
		name     string
		input    string
		expected string
	}

	tests := []testCase{
		{name: "BasicRegex", input: ":id[^\\d+$]", expected: "^\\d+$"},
		{name: "EmptyRegex", input: ":id[]", expected: ""},
		{name: "NoRegex", input: ":id", expected: "(.+)"},
		{name: "LiteralRegex", input: ":id[xxx]", expected: "xxx"},
		{name: "WildcardRegex", input: ":id[*]", expected: "*"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ret := deriveLabelPattern(test.input)
			if ret != test.expected {
				t.Errorf("expected %s but got %s\n", test.expected, ret)
			}
		})
	}
}

func TestDeriveParameterKey(t *testing.T) {
	type testCase struct {
		name     string
		input    string
		expected string
	}

	tests := []testCase{
		{name: "BasicKey", input: ":id[^\\d+$]", expected: "id"},
		{name: "BasicKeyEmptyRegex", input: ":val[]", expected: "val"},
		{name: "BasicKeyWildcardRegex", input: ":ex[(.*)]", expected: "ex"},
		{name: "BasicKeyNoRegex", input: ":id", expected: "id"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if deriveParameterKey(test.input) != test.expected {
				t.Errorf("expected %s but got %s\n", test.expected, test.input)
			}
		})
	}
}

func areSlicesEqByValue(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
