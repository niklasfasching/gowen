package gowen_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/niklasfasching/gowen"
	_ "github.com/niklasfasching/gowen/lib/core"
)

type evalTest struct {
	name     string
	input    string
	expected string
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
		if err := compare(test.input, test.expected); err != nil {
			t.Errorf("%s: got %s", test.name, err)
		}
	}
}

func compare(input string, expected string) error {
	env := gowen.NewEnv(false)
	inputNodes, err := parseAndEval(input, env)
	if err != nil {
		return err
	}
	expectedNodes, err := parseAndEval(expected, env)
	if err != nil {
		return err
	}
	inputNode := inputNodes[len(inputNodes)-1]
	expectedNode := expectedNodes[len(expectedNodes)-1]
	if !reflect.DeepEqual(inputNode, expectedNode) {
		return fmt.Errorf("\n\t%+v\nexpected\n\t%+v", inputNode, expectedNode)
	}
	return nil
}

func parseAndEval(input string, env *gowen.Env) (nodes []gowen.Node, err error) {
	nodes, err = gowen.Parse(input)
	if err != nil {
		return nil, err
	}
	return gowen.EvalMultiple(nodes, env)
}
