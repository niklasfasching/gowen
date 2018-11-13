package gowen

import (
	"reflect"
	"testing"
)

type destructureTest struct {
	name   string
	params string
	values string
	output string
}

var destructureTests = []destructureTest{
	{"sequential vector", "[x y z]", "[1 2 3]", `{"x" 1 "y" 2 "z" 3}`},
	{"sequential list", "[x y z]", "(1 2 3)", `{"x" 1 "y" 2 "z" 3}`},
	{"sequential string", "[x y z]", `"abc"`, `{"x" "a" "y" "b" "z" "c"}`},
	{"sequential shorter", "[x y z]", "[1]", `{"x" 1 "y" nil "z" nil}`},
	{"sequential longer", "[x y z]", "[1 2 3 4]", `{"x" 1 "y" 2 "z" 3}`},
	{"sequential & rest", "[x & xs]", "[1 2 3]", `{"x" 1 "xs" '(2 3)}`},
	{"sequential ignore", "[x _ z]", "[1 2 3]", `{"x" 1 "z" 3}`},
	{"sequential vector :all", "[x :as xs]", "[1 2 3]", `{"x" 1 "xs" [1 2 3]}`},
	{"sequential list :all", "[x :as xs]", "(1 2 3)", `{"x" 1 "xs" '(1 2 3)}`},
	{"sequential string :all", "[x :as xs]", `"abc"`, `{"x" "a" "xs" "abc"}`},
	{"sequential ignore & :all", "[x _ & xs :as all-xs]", "[1 2 3]", `{"x" 1 "xs" '(3) "all-xs" [1 2 3]}`},
	{"sequential nested", "[[x] [y z]]", "[[1] [2 3]]", `{"x" 1 "y" 2 "z" 3}`},

	{"associative", "{x :x y :y z :z}", "{:x 1 :y 2 :z 3}", `{"x" 1 "y" 2 "z" 3}`},
	{"associative missing", "{x :x y :y z :z}", "{:z 3}", `{"x" nil "y" nil "z" 3}`},
	{"associative :as", "{x :x :as m}", "{:x 1 :y 2}", `{"x" 1 "m" {:x 1 :y 2}}`},
	{"associative :keys", "{:keys [x y]}", "{:x 1 :y 2}", `{"x" 1 "y" 2}`},
	{"associative :keys & :as", "{:keys [x y] :as m}", "{:x 1 :y 2}", `{"x" 1 "y" 2 "m" {:x 1 :y 2}}`},
	{"associative nested", "{{x :x1} :x0 [_ _ y] :y}", "{:x0 {:x1 1} :y [1 2 3]}", `{"x" 1 "y" 3}`},
	{"associative from sequence", "{x :x :as m}", "[:x 1 :y 2]", `{"x" 1 "m" {:x 1 :y 2}}`},
}

func TestDestructure(t *testing.T) {
	for _, test := range destructureTests {
		env := NewEnv(false)
		value := parse(test.values)[0]
		destructure(parse(test.params)[0], value, env)
		expected := eval(parse(test.output)[0], env).ToGo()
		result := Map{}
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
