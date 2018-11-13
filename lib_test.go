package gowen_test

import (
	"testing"
)

var libTests = []evalTest{
	{"if", `(if true "foo" "bar")`, `"foo"`},
	{"def", `(def foo "foo") foo`, `"foo"`},
	{"quote", `'[foo bar baz]`, `'[foo bar baz]`},
	{"fn", "((fn [x] 42) 0)", "42"},
	{"fn destructure", "((fn [x [y1 y2] z] (+ x y1 y2 z)) 1 [2 3 4] 5)", "11"},
	{"macro", "((fn [x] 42) 0)", "42"},
	{"apply", `(apply (fn [x y] [x y]) [1 2])`, "[1 2]"},
	{"try", `[(try (throw "boo!") (catch e (str "caught: " e)))
              (try :foobar (catch e "caught"))]`, `["caught: boo!: (throw \"boo!\")" :foobar]`},

	{"macroexpand & defn", "(macroexpand '(defn foo [x & xs] x))", "'(def foo (fn foo [x & xs] x))"},
	{"macroexpand & defmacro", "(macroexpand '(defmacro foo [x & xs] x))", "'(def foo (macro foo [x & xs] x))"},

	{"q list", "'(+ 1 2)", "'(+ 1 2)"},
	{"qq unquote", "`(+ 1 ~(+ 1 2))", "'(+ 1 3)"},
	{"qq unquote-splicing", "`(+ 1 ~@(list 2 3))", "'(+ 1 2 3)"},
	{"qq unquote-splicing", "``(+ 1 ~~@(list 2 3))", "'(+ 1 2 3)"},
	{"qq vector", "`[+ 1 2 ~(+ 1 2)]", "'[+ 1 2 3]"},
	{"qq nested", "`[1 ~@[2] `[3 ~~(+ 3 1) ~(+ 4 1)]]", "'[1 2 [3 4 (+ 4 1)]]"},
}

func TestLib(t *testing.T) {
	for _, test := range libTests {
		if err := compare(test.input, test.expected); err != nil {
			t.Errorf("%s: got %s", test.name, err)
		}
	}
}
