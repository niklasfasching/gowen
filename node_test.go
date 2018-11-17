package gowen

import (
	"reflect"
	"testing"
)

type readPrintReadTest struct {
	name     string
	input    string
	expected string
}

var readPrintReadPrintTests = []readPrintReadTest{
	{"numbers", "0 1 2.0 -3 +4.20 1e6 1000000", "0 1 2 -3 4.2 1000000 1e6"},
	{"strings", `"foo" "foo\nbar"`, `"foo" "foo\nbar"`},
	{"collections", "{1 2} [3 4] (5 6)", "{1 2} [3 4] (5 6)"},
}

func TestReadPrintReadPrint(t *testing.T) {
	for _, test := range readPrintReadPrintTests {
		results := []Node{}
		for _, node := range parse(test.input) {
			results = append(results, parse(node.String())[0])
		}
		expected := parse(test.expected)
		if !reflect.DeepEqual(results, expected) {
			t.Errorf("%s: got\n\t%v\nexpected\n\t%v", test.name, results, expected)
		}
	}
}
