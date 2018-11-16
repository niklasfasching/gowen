package gowen

import "fmt"

func (n SymbolNode) String() string { return n.Value }

func (n KeywordNode) String() string { return ":" + n.Value }

func (n LiteralNode) String() string { return fmt.Sprintf("%#v", n.Value) }

func (n ListNode) String() string {
	s := "("
	for _, n := range n.Nodes {
		s += n.String() + " "
	}
	if len(s) > 1 {
		s = s[:len(s)-1]
	}
	return s + ")"
}

func (n VectorNode) String() string {
	s := "["
	for _, n := range n.Nodes {
		s += n.String() + " "
	}
	if len(s) > 1 {
		s = s[:len(s)-1]
	}
	return s + "]"
}

func (n MapNode) String() string {
	s := "{"
	for k, v := range n.Nodes {
		s += k.String() + " " + v.String() + ", "
	}
	if len(s) > 1 {
		s = s[:len(s)-2]
	}
	return s + "}"
}

func (n ArrayMapNode) String() string {
	s := "{"
	for i := 0; i < len(n.Nodes); i += 2 {
		s += n.Nodes[i].String() + " " + n.Nodes[i+1].String() + ", "
	}
	if len(s) > 1 {
		s = s[:len(s)-2]
	}
	return s + "}"
}
