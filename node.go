package gowen

import (
	"fmt"
	"reflect"
)

type Node interface {
	ToGo() Any
	String() string
	Seq() []Node
	Conj(Node) Node
}

type SymbolNode struct{ Value string }
type KeywordNode struct{ Value string }
type LiteralNode struct{ Value Any }

type ListNode struct{ Nodes []Node }
type VectorNode struct{ Nodes []Node }
type MapNode struct{ Nodes map[Node]Node }
type ArrayMapNode struct{ Nodes []Node }

func (n SymbolNode) Seq() []Node      { panic(errorf("seq on SymbolNode %v", n)) }
func (n SymbolNode) Conj(_ Node) Node { panic(errorf("conj on SymbolNode %v", n)) }
func (n SymbolNode) String() string   { return n.Value }
func (n SymbolNode) ToGo() Any        { return n }

func (n KeywordNode) Seq() []Node      { panic(errorf("seq on KeywordNode %v", n)) }
func (n KeywordNode) Conj(_ Node) Node { panic(errorf("conj on KeywordNode %v", n)) }
func (n KeywordNode) String() string   { return ":" + n.Value }
func (n KeywordNode) ToGo() Any        { return n }

func (n ListNode) Seq() []Node      { return n.Nodes }
func (n ListNode) Conj(x Node) Node { return ListNode{copyAppendNodes([]Node{x}, n.Nodes...)} }

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

func (n ListNode) ToGo() Any {
	values := make([]Any, len(n.Nodes))
	for i, cn := range n.Nodes {
		values[i] = cn.ToGo()
	}
	return values
}

func (n VectorNode) Seq() []Node      { return n.Nodes }
func (n VectorNode) Conj(x Node) Node { return VectorNode{copyAppendNodes(n.Nodes, x)} }
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
func (n VectorNode) ToGo() Any {
	values := make([]Any, len(n.Nodes))
	for i, cn := range n.Nodes {
		values[i] = cn.ToGo()
	}
	return values
}

func (n MapNode) Seq() []Node {
	ns := []Node{}
	for k, v := range n.Nodes {
		ns = append(ns, VectorNode{[]Node{k, v}})
	}
	return ns
}

func (n MapNode) Conj(x Node) Node {
	m := map[Node]Node{}
	for k, v := range n.Nodes {
		m[k] = v
	}
	ns := seq(x)
	m[ns[0]] = ns[1]
	return MapNode{m}
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

func (n MapNode) ToGo() Any {
	m := map[Any]Any{}
	for k, v := range n.Nodes {
		m[k.ToGo()] = v.ToGo()
	}
	return m
}

func (n ArrayMapNode) Seq() []Node {
	ns := []Node{}
	for i := 0; i < len(n.Nodes); i += 2 {
		ns = append(ns, VectorNode{n.Nodes[i : i+2]})
	}
	return ns
}

func (n ArrayMapNode) Conj(x Node) Node {
	return ArrayMapNode{copyAppendNodes(n.Nodes, x.(VectorNode).Nodes...)}
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

func (n ArrayMapNode) ToGo() Any {
	m := map[Any]Any{}
	for i := 0; i < len(n.Nodes); i += 2 {
		m[n.Nodes[i].ToGo()] = n.Nodes[i+1].ToGo()
	}
	return m
}

func (n LiteralNode) Seq() []Node {
	switch v := reflect.ValueOf(n.Value); {
	case n.Value == nil:
		return []Node{}
	case v.Kind() == reflect.String:
		s := v.String()
		ns := make([]Node, len(s))
		for i, c := range s {
			ns[i] = LiteralNode{string(c)}
		}
		return ns
	case v.Kind() == reflect.Slice:
		ns := make([]Node, v.Len())
		for i := 0; i < v.Len(); i++ {
			ns[i] = ToNode(v.Index(i).Interface())
		}
		return ns
	case v.Kind() == reflect.Map:
		ns := []Node{}
		for _, k := range v.MapKeys() {
			kv := VectorNode{[]Node{ToNode(k.Interface()), ToNode(v.MapIndex(k).Interface())}}
			ns = append(ns, kv)
		}
		return ns
	default:
		panic(errorf("seq on LiteralNode %s", n))
	}
}

func (n LiteralNode) Conj(x Node) Node {
	switch v := reflect.ValueOf(n.Value); {
	case n.Value == nil:
		return ListNode{[]Node{x}}
	case v.Kind() == reflect.Slice:
		ns := make([]Node, v.Len())
		for i := 0; i < v.Len(); i++ {
			ns[i] = ToNode(v.Index(i).Interface())
		}
		return ListNode{append(ns, x)}
	case v.Kind() == reflect.Map:
		ns := make([]Node, v.Len()*2)
		for _, k := range v.MapKeys() {
			ns = append(ns, ToNode(k.Interface()), ToNode(v.MapIndex(k).Interface()))
		}
		return ArrayMapNode{append(ns, seq(x)...)}
	default:
		panic(errorf("conj on LiteralNode %v", n))
	}
}

func (n LiteralNode) String() string { return fmt.Sprintf("%#v", n.Value) }
func (n LiteralNode) ToGo() Any      { return n.Value }
