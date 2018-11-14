package gowen

import (
	"reflect"
	"testing"
)

type fromGoToGoTest struct {
	name   string
	input  Any
	inGo   Node
	output Any
}

var exampleStruct = fromGoToGoTest{name: "foo"}
var fromGoToGoTests = []fromGoToGoTest{
	{"nil", nil, LiteralNode{nil}, nil},
	{"string", "foo", LiteralNode{"foo"}, "foo"},
	{"struct", exampleStruct, LiteralNode{exampleStruct}, exampleStruct},
	{"int (-> float64)", 1, LiteralNode{1.0}, 1.0},
	{"float64", 1.0, LiteralNode{1.0}, 1.0},

	{"KeywordNode", KeywordNode{"foo"}, KeywordNode{"foo"}, KeywordNode{"foo"}},
	{"SymbolNode", SymbolNode{"foo"}, SymbolNode{"foo"}, SymbolNode{"foo"}},

	{"[]Any (-> List)", []Any{1, "foo"}, ListNode{[]Node{LiteralNode{1.0}, LiteralNode{"foo"}}}, List{1.0, "foo"}},
	{"non-any [] (-> List)", []int{1, 2}, ListNode{[]Node{LiteralNode{1.0}, LiteralNode{2.0}}}, List{1.0, 2.0}},
	{"List", List{1, "foo"}, ListNode{[]Node{LiteralNode{1.0}, LiteralNode{"foo"}}}, List{1.0, "foo"}},

	{"Vector", Vector{1, "foo"}, VectorNode{[]Node{LiteralNode{1.0}, LiteralNode{"foo"}}}, Vector{1.0, "foo"}},

	{"Map", Map{1: "foo"}, MapNode{map[Node]Node{LiteralNode{1.0}: LiteralNode{"foo"}}}, Map{1.0: "foo"}},
	{"non-any map (-> Map)", map[int]string{1: "foo"}, MapNode{map[Node]Node{LiteralNode{1.0}: LiteralNode{"foo"}}}, Map{1.0: "foo"}},
	{"map[Any]Any (-> Map)", map[Any]Any{1: "foo"}, MapNode{map[Node]Node{LiteralNode{1.0}: LiteralNode{"foo"}}}, Map{1.0: "foo"}},
}

func TestFromGoToGo(t *testing.T) {
	for _, test := range fromGoToGoTests {
		inGo := FromGo(test.input)
		if !reflect.DeepEqual(inGo, test.inGo) {
			t.Errorf("%s: (inGo) got\n\t%v\nexpected\n\t%v", test.name, inGo, test.inGo)
		}
		output := ToGo(inGo)
		if !reflect.DeepEqual(output, test.output) {
			t.Errorf("%s: (toGo) got\n\t%#v\nexpected\n\t%#v", test.name, output, test.output)
		}
	}
}

type applyInteropTest struct {
	name   string
	fn     Any
	args   []Any
	output Any
}

var applyInteropTests = []applyInteropTest{
	{"basic",
		func(x, y float64) Any { return x + y },
		[]Any{1, 2},
		3.0,
	},

	{"basic variadic",
		func(x float64, xs ...float64) Any { return x + xs[0] },
		[]Any{1, 2, 3, 4, 5},
		3.0,
	},

	{"convert (float64 -> int)",
		func(x int) Any { return x },
		[]Any{1},
		1.0,
	},

	{"convert (float64 -> *int)",
		func(x *int) Any { return *x },
		[]Any{1},
		1.0,
	},

	{"convert (*float64 -> int)",
		func(x int) Any { return x },
		[]Any{new(float64)},
		0.0,
	},

	{"convert (*[]Any -> []int)",
		func(x []int) Any { return x },
		[]Any{&[]Any{1, 2}},
		[]Any{1, 2},
	},

	{"nil as []Any",
		func(x []Any) Any { return x },
		[]Any{nil},
		List{},
	},
	{"nil as non-any []",
		func(x []int) Any { return x },
		[]Any{nil},
		List{},
	},
	{"nil as map[Any]Any",
		func(x map[Any]Any) Any { return x },
		[]Any{nil},
		Map{},
	},
	{"nil as non-any map",
		func(x map[int]int) Any { return x },
		[]Any{nil},
		Map{},
	},

	{"convert ([]Any -> []int)",
		func(xs []int) Any { return xs[0] + xs[1] },
		[]Any{List{1, 2, 3}},
		3.0,
	},
	{"convert ([]Any -> []string)",
		func(xs []string) Any { return xs[0] + xs[1] },
		[]Any{List{"foo", "bar"}},
		"foobar",
	},

	{"convert (map[Any]Any -> map[int]string)",
		func(xs map[int]string) Any { return xs },
		[]Any{Map{1: "bar"}},
		Map{1: "bar"},
	},
}

func TestApplyInterop(t *testing.T) {
	for _, test := range applyInteropTests {
		argns := make([]Node, len(test.args))
		for i, arg := range test.args {
			argns[i] = FromGo(arg)
		}
		output := applyInterop(LiteralNode{test.fn}, argns)
		expected := FromGo(test.output)
		if !reflect.DeepEqual(output, expected) {
			t.Errorf("%s: got\n\t%#v\nexpected\n\t%#v", test.name, output, expected)
		}
	}
}

type applyMemberInteropTest struct {
	name     string
	it       LiteralNode
	input    string
	expected []Any
}

type applyMemberExample struct {
	Value   string
	Pointer *string
}

func (_ applyMemberExample) ValueMethod(x int) int    { return x }
func (_ *applyMemberExample) PointerMethod(x int) int { return x }

var applyMemberInteropTests = []applyMemberInteropTest{
	{"zero value",
		LiteralNode{applyMemberExample{}},
		"[(.valueMethod it 1) (.pointerMethod it 1) (.value it) (.pointer it)]",
		[]Any{1.0, 1.0, "", (*string)(nil)},
	},
	{"value",
		LiteralNode{applyMemberExample{"foo", new(string)}},
		"[(.valueMethod it 1) (.pointerMethod it 1) (.value it) (.pointer it)]",
		[]Any{1.0, 1.0, "foo", new(string)},
	},
}

func TestApplyMemberInterop(t *testing.T) {
	for _, test := range applyMemberInteropTests {
		env := NewEnv(false)
		env.Set("it", test.it)
		result := eval(parse(test.input)[0], env).ToGo()
		expected := Vector(test.expected)
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("%s: got\n\t%#v\nexpected\n\t%#v", test.name, result, expected)
		}
	}
}
