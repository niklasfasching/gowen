package core_test

import (
	"reflect"
	"testing"

	"github.com/niklasfasching/gowen"
)

type coreTest struct {
	name   string
	input  string
	output string
}

var coreTests = []coreTest{
	{"+", "(+ 10 (- 20 10) (* 5 2) (/ 20 2))", "40"},
	{"count", "[(count [1 2 3]) (count '(1 2))]", "[3 2]"},
	{"let", "(let [x 1 y 2] (+ x y))", "3"},
	{"do", "(do 1 2 3)", "3"},
	{"and", "(and 1 2 false 3)", "false"},
	{"or", "(or false nil 3 false)", "3"},
	{"reduce", "(reduce (fn [x y] (+ x y)) 0 [1 2 3 4])", "10"},
	{"map", "(map (fn [x] (+ x 1)) [1 2 3])", "'(2 3 4)"},
	{"filter", "(filter (fn [x] (> x 1)) [0 1 2 3])", "'(2 3)"},
	{"type", "(type :foo)", `"keyword"`},
	{"hashmap", `{"a" (+ 1 2) 2 "b"}`, `{2 "b" "a" 3}`},
	{"cond", `(cond false 1 nil 2 true 3)`, "3"},
	{"spit & slurp", `(spit "/tmp/spat" "yo") (slurp "/tmp/spat")`, `"yo"`},
}

func TestCore(t *testing.T) {
	for _, test := range coreTests {
		env := gowen.NewEnv(false)
		nodes := gowen.EvalMultiple(gowen.Expand(gowen.Parse(test.input), env), env)
		result := nodes[len(nodes)-1]
		expected := gowen.EvalMultiple(gowen.Parse(test.output), env)[0]
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("%s: got\n\t%v\nexpected\n\t%v", test.name, result, expected)
		}
	}
}
