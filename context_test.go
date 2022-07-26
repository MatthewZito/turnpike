package turnpike

import (
	"context"
	"testing"
)

func TestContext(t *testing.T) {
	type testCase struct {
		name     string
		actual   string
		expected string
	}

	params := &[]*parameter{
		{
			key:   "id",
			value: "12",
		},
		{
			key:   "user",
			value: "uxc",
		},
	}

	ctx := context.WithValue(context.Background(), parameterKey, *params)

	tests := []testCase{
		{
			name:     "BasicKey",
			expected: "12",
			actual:   GetParam(ctx, "id"),
		},
		{
			name:     "KeyNotExtant",
			expected: "",
			actual:   GetParam(ctx, "test"),
		},
		{
			name:     "BasicKeySecond",
			expected: "uxc",
			actual:   GetParam(ctx, "user"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.actual != test.expected {
				t.Errorf("expected %s but got %s", test.expected, test.actual)
			}
		})
	}
}
