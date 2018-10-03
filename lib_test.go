package gowen

import (
	"reflect"
	"testing"
)

var libTests = []evalTest{
	{"if", `(if true "foo" "bar")`, `"foo"`},
	{"def", `(def foo "foo") foo`, `"foo"`},
	{"quote", `'[foo bar baz]`, `[foo bar baz]`},
	{"fn", "((fn [x] 42) 0)", "42"},
	{"fn destructure", "((fn [x [y1 y2] z] (+ x y1 y2 z)) 1 [2 3 4] 5)", "11"},
	{"macro", "((fn [x] 42) 0)", "42"},
	{"apply", `(apply (fn [x y] [x y]) [1 2])`, "[1 2]"},
	{"try", `[(try (throw "boo!") (catch e (str "caught: " e)))
              (try :foobar (catch e "caught"))]`, `["caught: boo!: (throw \"boo!\")" :foobar]`},

	{"macroexpand & defn", "(macroexpand '(defn foo [x & xs] x))", "(def foo (fn [x & xs] x))"},
	{"macroexpand & defmacro", "(macroexpand '(defmacro foo [x & xs] x))", "(def foo (macro [x & xs] x))"},

	{"q list", "'(+ 1 2)", "(+ 1 2)"},
	{"qq unquote", "`(+ 1 ~(+ 1 2))", "(+ 1 3)"},
	{"qq unquote-splicing", "`(+ 1 ~@(list 2 3))", "(+ 1 2 3)"},
	{"qq unquote-splicing", "``(+ 1 ~~@(list 2 3))", "(+ 1 2 3)"},
	{"qq vector", "`[+ 1 2 ~(+ 1 2)]", "[+ 1 2 3]"},
	{"qq nested", "`[1 ~@[2] `[3 ~~(+ 3 1) ~(+ 4 1)]]", "[1 2 [3 4 (+ 4 1)]]"},
}

func TestLib(t *testing.T) {
	for _, test := range libTests {
		env := NewEnv()
		nodes := EvalMultiple(Parse(test.input), env)
		result := nodes[len(nodes)-1]
		expected := Parse(test.output)[0]
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("%s: got\n\t%v\nexpected\n\t%v", test.name, result, expected)
		}
	}
}

type patternMatchTest struct {
	name   string
	params string
	values string
	output string
}

var patternMatchTests = []patternMatchTest{
	{"simple", "[x y z]", "[1 2 [3 4]]", `{"x" 1 "y" 2 "z" [3 4]}`},
	{"variadic", "[x & xs]", "[1 2 [3 4]]", `{"x" 1 "xs" '(2 [3 4])}`},
	{"nested vector", "[x y [z1 z2]]", "[1 2 [3 4]]", `{"x" 1 "y" 2 "z1" 3 "z2" 4}`},
	{"nested hash", "[x y {z :z}]", "[1 2 {:z 3}]", `{"x" 1 "y" 2 "z" 3}`},
	{"hash", "[{x :x y :y}]", "[{:x 1}]", `{"x" 1 "y" nil}`},
	{"hash keys", "[{:keys [x y]}]", "[{:x 1}]", `{"x" 1 "y" nil}`},
}

func TestPatternMatch(t *testing.T) {
	for _, test := range patternMatchTests {
		parentEnv := NewEnv()
		env := &Env{parentEnv, nil}
		value := Eval(Parse(test.values)[0], env)
		match(Parse(test.params)[0], value, env)
		expected := Eval(Parse(test.output)[0], parentEnv).ToGo()
		result := map[interface{}]interface{}{}
		for k, v := range env.values {
			if v == nil {
				result[k] = nil
			} else {
				result[k] = v.(Node).ToGo()
			}
		}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("%s: got\n\t%#v\nexpected\n\t%#v", test.name, result, expected)
		}
	}
}
