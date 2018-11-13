package gowen

import (
	"reflect"
	"testing"
)

// var sortTests = []builtinTest{
// 	{"defs", `
// 	   (def foo 1)
// 	   (def bar 1)
// 	   (+ foo bar)
// 	   `, LiteralNode{2.0, 0}},
// 	{"defs and defns", `
// 	   (def foobarbaz (foo bar baz))
// 	   (defn foo [bar baz] (+ bar baz))
// 	   (def bar (+ baz 1))
// 	   (def baz 1)
// 	   (+ bar baz foobarbaz 1)
// 	   `, LiteralNode{7.0, 0}},
// 	{"misc", `
//        (def foobar [bar baz])
//        (def bar (foo x y z))
//        (def baz [bar foo])
//        (defmacro foo [& body] "foo")
//        (first foobar)
// 	   `, LiteralNode{"foo", 0}},
// }

// func TestSort(t *testing.T) {
// 	for _, test := range sortTests {
// 		env := &Env{GlobalEnv, nil}
// 		nodes := parse(test.input)
// 		result := unindexNode(EvalTopological(nodes, env))
// 		if !reflect.DeepEqual(result, test.result) {
// 			t.Errorf("%s: got\n\t%+v\nexpected\n\t%+v", test.name, result, test.result)
// 		}
// 	}
// }

type expandTest struct {
	name   string
	eval   string
	input  string
	output string
}

var expandTests = []expandTest{
	{"simple macro expansion",
		`(def foo (macro [] 'bar))`, `((fn [x] (foo)) 1)`, `((fn [x] bar) 1)`},
	{"fn shadowing",
		`(def foo (macro [] 'bar))`, `((fn [foo] (foo)) 1)`, `((fn [foo] (foo)) 1)`},
	{"quote shadowing",
		`(def foo (macro [] 'bar))`, `((fn [] '(foo) (foo)) 1)`, `((fn [] '(foo) bar) 1)`},
}

func TestExpand(t *testing.T) {
	for _, test := range expandTests {
		env := NewEnv(false)
		EvalMultiple(parse(test.eval), env)
		nodes := expand(parse(test.input), env)
		expanded := nodes[len(nodes)-1]
		expected := parse(test.output)[0]
		if !reflect.DeepEqual(expanded, expected) {
			t.Errorf("%s: got\n\t%s\nexpected\n\t%s", test.name, expanded, expected)
		}
	}
}
