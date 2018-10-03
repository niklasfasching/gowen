package gowen

import (
	"reflect"
	"testing"
)

type parseTest struct {
	name  string
	input string
	nodes []Node
}

var parseTests = []parseTest{
	{"empty", "", []Node{}},

	{"spaces", " \t\n \n\t ", []Node{}},

	{"strings", `"foo" "" "foo\nbar" "foo
      bar
      baz"`, []Node{
		LiteralNode{"foo"},
		LiteralNode{""},
		LiteralNode{"foo\nbar"},
		LiteralNode{"foo\n      bar\n      baz"},
	}},

	{"vectors", `[1 2 "foo"] []`, []Node{
		VectorNode{[]Node{
			LiteralNode{1.0},
			LiteralNode{2.0},
			LiteralNode{"foo"},
		}},
		VectorNode{[]Node{}},
	}},

	{"lists", `() (+ 1 2)`, []Node{
		ListNode{[]Node{}},
		ListNode{[]Node{
			SymbolNode{"+"},
			LiteralNode{1.0},
			LiteralNode{2.0},
		}},
	}},

	{"maps", `{} {:foo [:bar]}`, []Node{
		MapNode{map[Node]Node{}},
		MapNode{map[Node]Node{
			KeywordNode{"foo"}: VectorNode{[]Node{
				KeywordNode{"bar"},
			}},
		}},
	}},

	{"keywords", ":foo/bar :42", []Node{
		KeywordNode{"foo/bar"},
		KeywordNode{"42"},
	}},

	{"quotes", " `foo 'bar ~baz ~@bam ", []Node{
		ListNode{[]Node{SymbolNode{"quasiquote"}, SymbolNode{"foo"}}},
		ListNode{[]Node{SymbolNode{"quote"}, SymbolNode{"bar"}}},
		ListNode{[]Node{SymbolNode{"unquote"}, SymbolNode{"baz"}}},
		ListNode{[]Node{SymbolNode{"unquote-splicing"}, SymbolNode{"bam"}}},
	}},
}

func TestParse(t *testing.T) {
	for _, test := range parseTests {
		nodes := Parse(test.input)
		if !reflect.DeepEqual(nodes, test.nodes) {
			t.Errorf("%s: got\n\t%#v\nexpected\n\t%#v", test.name, nodes, test.nodes)
		}
	}
}
