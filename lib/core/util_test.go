package core

import (
	"testing"
)

type lispCaseTest struct {
	input    string
	expected string
}

var lispCaseTests = []lispCaseTest{
	{"FooBar", "foo-bar"},
	{"FooBarBAZ", "foo-bar-baz"},
	{"FOOBarBAZ", "foo-bar-baz"},
	{"Foo123Bar", "foo-123-bar"},
	{"Foo_bar_Baz", "foo-bar-baz"},
}

func TestToLispCase(t *testing.T) {
	for _, test := range lispCaseTests {
		if result := toLispCase(test.input); result != test.expected {
			t.Errorf("got:\n\t%v\nexpected\n\t%v", result, test.expected)
		}
	}
}
