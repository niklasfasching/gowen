package gowen

import (
	"fmt"
)

type Error struct {
	context Any
	error   error
}

func (e Error) Error() string {
	return fmt.Sprintf("%s: %s", e.error, e.context)
}

func errorf(format string, vs ...Any) error { return fmt.Errorf(format, vs...) }

func handleError(err *error) {
	if e := recover(); e != nil {
		*err = Error{e, errorf("gowen")}
	}
}

func callTo(n Node) string {
	ln, _ := n.(ListNode)
	if len(ln.Nodes) == 0 {
		return ""
	}
	sn, _ := ln.Nodes[0].(SymbolNode)
	return sn.Value
}

func assert(assertion bool, format string, vs ...Any) {
	if !assertion {
		panic(errorf(format, vs...))
	}
}

func wrapInCall(symbol string, ns []Node) ListNode {
	return ListNode{append([]Node{SymbolNode{symbol}}, ns...)}
}

func copyAppendNodes(ns1 []Node, ns2 ...Node) []Node {
	return append(append([]Node{}, ns1...), ns2...)
}

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
