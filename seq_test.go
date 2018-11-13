package gowen

import (
	"reflect"
	"testing"
)

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
	{"nil", "nil", 0, nil},
	{"List", "'(1 2 3)", 1, 2.0},
	{"Vector", "[1 2 3]", 1, 2.0},
	{"Map", "{1 2 3 4}", 3, 4.0},
	{"ArrayMap", "'{1 2 3 4}", 3, 4.0},
}

func TestGet(t *testing.T) {
	for _, test := range getTests {
		env := NewEnv(false)
		result := eval(wrapInCall("get", append(parse(test.input), FromGo(test.key))), env)
		expected := FromGo(test.output)
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
