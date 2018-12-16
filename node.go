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
	Get(Node) Node
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
func (n SymbolNode) Get(_ Node) Node  { panic(errorf("get on SymbolNode %v", n)) }
func (n SymbolNode) String() string   { return n.Value }
func (n SymbolNode) ToGo() Any        { return n }

func (n KeywordNode) Seq() []Node      { panic(errorf("seq on KeywordNode %v", n)) }
func (n KeywordNode) Conj(_ Node) Node { panic(errorf("conj on KeywordNode %v", n)) }
func (n KeywordNode) Get(_ Node) Node  { panic(errorf("get on KeywordNode %v", n)) }
func (n KeywordNode) String() string   { return ":" + n.Value }
func (n KeywordNode) ToGo() Any        { return n }

func (n ListNode) Seq() []Node      { return n.Nodes }
func (n ListNode) Conj(x Node) Node { return ListNode{copyAppendNodes([]Node{x}, n.Nodes...)} }

func (n ListNode) Get(x Node) Node {
	i := reflect.ValueOf(x.(LiteralNode).Value).Convert(reflect.TypeOf(0)).Int()
	if len(n.Nodes) <= int(i) {
		return LiteralNode{nil}
	}
	return n.Nodes[i]
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

func (n ListNode) ToGo() Any {
	values := make([]Any, len(n.Nodes))
	for i, cn := range n.Nodes {
		values[i] = cn.ToGo()
	}
	return values
}

func (n VectorNode) Seq() []Node      { return n.Nodes }
func (n VectorNode) Conj(x Node) Node { return VectorNode{copyAppendNodes(n.Nodes, x)} }

func (n VectorNode) Get(x Node) Node {
	i := reflect.ValueOf(x.(LiteralNode).Value).Convert(reflect.TypeOf(0)).Int()
	if len(n.Nodes) <= int(i) {
		return LiteralNode{nil}
	}
	return n.Nodes[i]
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
	ns := x.Seq()
	m[ns[0]] = ns[1]
	return MapNode{m}
}

func (n MapNode) Get(x Node) Node {
	v, ok := n.Nodes[x]
	if !ok {
		return LiteralNode{nil}
	}
	return v
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

func (n ArrayMapNode) Get(x Node) Node {
	for i := 0; i < len(n.Nodes); i += 2 {
		if reflect.DeepEqual(n.Nodes[i], x) {
			return n.Nodes[i+1]
		}
	}
	return LiteralNode{nil}
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
		panic(errorf("seq on LiteralNode %#v", n))
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
		return ArrayMapNode{append(ns, x.Seq()...)}
	default:
		panic(errorf("conj on LiteralNode %v", n))
	}
}

func (n LiteralNode) Get(x Node) Node {
	switch v := reflect.ValueOf(n.Value); {
	case n.Value == nil:
		return n
	case v.Kind() == reflect.Slice:
		i := reflect.ValueOf(x.(LiteralNode).Value).Convert(reflect.TypeOf(0)).Int()
		if int(i) >= v.Len() {
			return LiteralNode{nil}
		}
		return LiteralNode{v.Index(int(i)).Interface()}
	case v.Kind() == reflect.Map:
		result := v.MapIndex(reflect.ValueOf(x.ToGo()))
		if result.IsValid() {
			return LiteralNode{result.Interface()}
		}
		return LiteralNode{nil}
	default:
		panic(errorf("conj on LiteralNode %v", n))
	}
}

func (n LiteralNode) String() string { return fmt.Sprintf("%#v", n.Value) }
func (n LiteralNode) ToGo() Any      { return n.Value }
