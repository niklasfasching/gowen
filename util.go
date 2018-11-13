package gowen

import (
	"fmt"
	"strconv"
)

type Any = interface{}
type Vector []Any
type List []Any
type Keyword string
type Symbol string
type Error struct {
	Context Any
	Message string
}

func (e Error) Error() string { return fmt.Sprintf("%s: %s", e.Message, e.Context) }

func CallTo(n Node) string {
	ln, _ := n.(ListNode)
	if len(ln.Nodes) == 0 {
		return ""
	}
	sn, _ := ln.Nodes[0].(SymbolNode)
	return sn.Value
}

func assert(assertion bool, format string, vs ...Any) {
	if !assertion {
		panic(fmt.Sprintf(format, vs...))
	}
}

func wrapInCall(symbol string, ns []Node) ListNode {
	return ListNode{append([]Node{SymbolNode{symbol}}, ns...)}
}

func (n SymbolNode) String() string { return n.Value }

func (v Keyword) String() string     { return ":" + string(v) }
func (n KeywordNode) String() string { return ":" + n.Value }

func (n LiteralNode) String() string {
	switch n.Value.(type) {
	case string:
		return fmt.Sprintf(`"%s"`, n.Value)
	case float64:
		return strconv.FormatFloat(n.Value.(float64), 'f', -1, 64)
	default:
		return fmt.Sprintf("›%#v‹", n.Value)
	}
}

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

func (n ListNode) ToGo() Any {
	values := make(List, len(n.Nodes))
	for i, cn := range n.Nodes {
		values[i] = cn.ToGo()
	}
	return values
}

func (n ArrayMapNode) ToGo() Any {
	m := make(map[Any]Any, len(n.Nodes))
	for i := 0; i < len(n.Nodes); i += 2 {
		m[n.Nodes[i].ToGo()] = n.Nodes[i+1].ToGo()
	}
	return m
}

func (n MapNode) ToGo() Any {
	m := make(map[Any]Any, len(n.Nodes))
	for k, v := range n.Nodes {
		m[k.ToGo()] = v.ToGo()
	}
	return m
}

func (n VectorNode) ToGo() Any {
	values := make(Vector, len(n.Nodes))
	for i, cn := range n.Nodes {
		values[i] = cn.ToGo()
	}
	return values
}

func (n LiteralNode) ToGo() Any { return n.Value }

func (n SymbolNode) ToGo() Any { return Symbol(n.Value) }

func (n KeywordNode) ToGo() Any { return Keyword(n.Value) }

// TODO: handle non interface slices and maps
func FromGo(value Any) Node {
	switch x := value.(type) {
	case Node:
		return x
	case []Any:
		nodes := make([]Node, len(x))
		for i, v := range x {
			nodes[i] = FromGo(v)
		}
		return ListNode{nodes}
	case List:
		nodes := make([]Node, len(x))
		for i, v := range x {
			nodes[i] = FromGo(v)
		}
		return ListNode{nodes}
	case Vector:
		nodes := make([]Node, len(x))
		for i, v := range x {
			nodes[i] = FromGo(v)
		}
		return VectorNode{nodes}
	case map[Any]Any:
		m := map[Node]Node{}
		for k, v := range x {
			m[FromGo(k)] = FromGo(v)
		}
		return MapNode{m}
	case Keyword:
		return KeywordNode{string(x)}
	case Symbol:
		return SymbolNode{string(x)}
	case int:
		return LiteralNode{float64(x)}
	default:
		return LiteralNode{x}
	}
}
