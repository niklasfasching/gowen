package gowen

import (
	"reflect"
	"testing"
)

type evalTest struct {
	name   string
	input  string
	output string
}

var evalTests = []evalTest{
	{"number", `42`, `42`},
	{"string", `"foo"`, `"foo"`},
	{"vector", "[1 2 (+ 1 2)]", "[1 2 3]"},
	{"map", "{(+ 1 2) 3}", "{3 3}"},
	{"list (call)", "(+ 1 2)", "3"},
	{"nested call", "(+ 1 (+ -1 (+ 1 -1)))", "0"},
	{"fn & nested call", "(def foo (fn [x y] [(+ x 1) (+ y 1)])) (foo 1 2)", "[2 3]"},
	{"env shadowing", "((fn [x y] ((fn [y z] (+ x (+ y z))) 2 2)) 1 1)", "5"},
}

func TestEval(t *testing.T) {
	for _, test := range evalTests {
		env := NewEnv(false)
		nodes := evalMultiple(parse(test.input), env)
		result := nodes[len(nodes)-1]
		expected := eval(parse(test.output)[0], env)
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("%s: got\n\t%#v\nexpected\n\t%#v", test.name, result, expected)
		}
	}
}
