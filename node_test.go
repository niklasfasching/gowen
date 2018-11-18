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

type seqTest struct {
	name   string
	input  string
	output string
}

var seqTests = []seqTest{
	{"ListNode", `'(1 2 3)`, `(1 2 3)`},
	{"VectorNode", `[1 2 3]`, `(1 2 3)`},
	{"MapNode", `{1 2}`, `([1 2])`},
	{"ArrayMapNode", `'{1 2}`, `([1 2])`},
	{"string", `"foo"`, `("f" "o" "o")`},
	{"nil", `nil`, `()`},
}

func TestSeq(t *testing.T) {
	for _, test := range seqTests {
		env := NewEnv(false)
		result := eval(wrapInCall("seq", parse(test.input)), env)
		expected := parse(test.output)[0]
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("%s: got\n\t%v\nexpected\n\t%v", test.name, result, expected)
		}
	}
}

type getTest struct {
	name   string
	input  string
	key    Any
	output Any
}

var getTests = []getTest{
	{"nil", "nil", 0.0, nil},
	{"List", "'(1 2 3)", 1.0, 2.0},
	{"Vector", "[1 2 3]", 1.0, 2.0},
	{"Map", "{1 2 3 4}", 3.0, 4.0},
	{"ArrayMap", "'{1 2 3 4}", 3.0, 4.0},
}

func TestGet(t *testing.T) {
	for _, test := range getTests {
		env := NewEnv(false)
		result := eval(wrapInCall("get", append(parse(test.input), ToNode(test.key))), env)
		expected := ToNode(test.output)
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("%s: got\n\t%v\nexpected\n\t%v", test.name, result, expected)
		}
	}
}

type conjTest struct {
	name   string
	xs     string
	x      string
	output string
}

var conjTests = []conjTest{
	{"nil", "nil", "4", "'(4)"},
	{"List", "'(1 2 3)", "4", "'(4 1 2 3)"},
	{"Vector", "[1 2 3]", "4", "[1 2 3 4]"},
	{"ArrayMap", "'{}", "[1 2]", "'{1 2}"},
	{"Map", "{}", "[1 2]", "{1 2}"},
}

func TestConj(t *testing.T) {
	for _, test := range conjTests {
		env := NewEnv(false)
		result := eval(wrapInCall("conj", []Node{
			parse(test.xs)[0],
			parse(test.x)[0],
		}), env)
		expected := eval(parse(test.output)[0], env)
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("%s: got\n\t%v\nexpected\n\t%v", test.name, result, expected)
		}
	}
}

type concatTest struct {
	name   string
	xs     string
	output string
}

var concatTests = []concatTest{
	{"nil", "nil nil nil", "()"},
	{"Vector", "[1] [2] nil [3 4] [5]", "'(1 2 3 4 5)"},
	{"List", "'(1 2) nil [3 4] [5]", "'(1 2 3 4 5)"},
	{"Map", "'(1 2) nil {3 4}", "'(1 2 [3 4])"},
	{"ArrayMap", "'(1 2) nil '{3 4}", "'(1 2 [3 4])"},
	{"string", `'(1 2) nil "foo"`, `'(1 2 "f" "o" "o")`},
}

func TestConcat(t *testing.T) {
	for _, test := range concatTests {
		env := NewEnv(false)
		result := eval(wrapInCall("concat", parse(test.xs)), env)
		expected := eval(parse(test.output)[0], env)
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("%s: got\n\t%v\nexpected\n\t%v", test.name, result, expected)
		}
	}
}
