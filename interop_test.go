package gowen

import (
	"reflect"
	"testing"
)

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
		[]Any{1.0},
		1,
	},

	{"convert (float64 -> *int)",
		func(x *int) Any { return *x },
		[]Any{1},
		1,
	},

	{"convert (*float64 -> int)",
		func(x int) Any { return x },
		[]Any{new(float64)},
		0,
	},

	{"convert (*[]Any -> []int)",
		func(x []int) Any { return x },
		[]Any{&[]Any{1.0, 2}},
		[]int{1, 2},
	},

	{"nil as []Any",
		func(x []Any) Any { return x },
		[]Any{nil},
		([]Any)(nil),
	},
	{"nil as non-any []",
		func(x []int) Any { return x },
		[]Any{nil},
		([]int)(nil),
	},
	{"nil as map[Any]Any",
		func(x map[Any]Any) Any { return x },
		[]Any{nil},
		(map[Any]Any)(nil),
	},
	{"nil as non-any map",
		func(x map[int]int) Any { return x },
		[]Any{nil},
		(map[int]int)(nil),
	},

	{"convert ([]Any -> []int)",
		func(xs []int) Any { return xs[0] + xs[1] },
		[]Any{[]Any{1.0, 2.0, 3}},
		3,
	},
	{"convert ([]Any -> []string)",
		func(xs []string) Any { return xs[0] + xs[1] },
		[]Any{[]Any{"foo", "bar"}},
		"foobar",
	},

	{"convert (map[Any]Any -> map[int]string)",
		func(xs map[int]string) Any { return xs },
		[]Any{map[Any]Any{1: "bar"}},
		map[int]string{1: "bar"},
	},
}

func TestApplyInterop(t *testing.T) {
	for _, test := range applyInteropTests {
		argns := make([]Node, len(test.args))
		for i, arg := range test.args {
			argns[i] = ToNode(arg)
		}
		output := applyInterop(LiteralNode{test.fn}, argns)
		expected := ToNode(test.output)
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
		[]Any{1, 1, "", (*string)(nil)},
	},
	{"value",
		LiteralNode{applyMemberExample{"foo", new(string)}},
		"[(.valueMethod it 1) (.pointerMethod it 1) (.value it) (.pointer it)]",
		[]Any{1, 1, "foo", new(string)},
	},
}

func TestApplyMemberInterop(t *testing.T) {
	for _, test := range applyMemberInteropTests {
		env := NewEnv(false)
		env.Set("it", test.it)
		result := eval(parse(test.input)[0], env).ToGo()
		expected := []Any(test.expected)
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("%s: got\n\t%#v\nexpected\n\t%#v", test.name, result, expected)
		}
	}
}
