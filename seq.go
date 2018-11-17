package gowen

import "reflect"

func (n SymbolNode) Seq() []Node      { panic(errorf("seq on SymbolNode %v", n)) }
func (n SymbolNode) Conj(_ Node) Node { panic(errorf("conj on SymbolNode %v", n)) }

func (n KeywordNode) Seq() []Node      { panic(errorf("seq on KeywordNode %v", n)) }
func (n KeywordNode) Conj(_ Node) Node { panic(errorf("conj on KeywordNode %v", n)) }

func (n ListNode) Seq() []Node      { return n.Nodes }
func (n ListNode) Conj(x Node) Node { return ListNode{copyAppendNodes([]Node{x}, n.Nodes...)} }

func (n VectorNode) Seq() []Node      { return n.Nodes }
func (n VectorNode) Conj(x Node) Node { return VectorNode{copyAppendNodes(n.Nodes, x)} }

func (n ArrayMapNode) Conj(x Node) Node {
	return ArrayMapNode{copyAppendNodes(n.Nodes, x.(VectorNode).Nodes...)}
}
func (n ArrayMapNode) Seq() []Node {
	ns := []Node{}
	for i := 0; i < len(n.Nodes); i += 2 {
		ns = append(ns, VectorNode{n.Nodes[i : i+2]})
	}
	return ns
}

func (n MapNode) Conj(x Node) Node {
	m := map[Node]Node{}
	for k, v := range n.Nodes {
		m[k] = v
	}
	kvs := seq(x)
	m[kvs[0]] = kvs[1]
	return MapNode{m}
}
func (n MapNode) Seq() []Node {
	ns := []Node{}
	for k, v := range n.Nodes {
		ns = append(ns, VectorNode{[]Node{k, v}})
	}
	return ns
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
		ns := make([]Node, v.Len()*2)
		for _, k := range v.MapKeys() {
			kv := VectorNode{[]Node{ToNode(k.Interface()), ToNode(v.MapIndex(k).Interface())}}
			ns = append(ns, kv)
		}
		return ns
	default:
		panic(errorf("seq on LiteralNode %s", n))
	}
}

func seq(n Node) []Node {
	return n.Seq()
}

func get(n Node, x Node) Node {
	switch n := n.(type) {
	case ListNode, VectorNode:
		i := reflect.ValueOf(x.(LiteralNode).Value).Convert(reflect.TypeOf(0)).Int()
		ns := seq(n)
		if len(ns) <= int(i) {
			return LiteralNode{nil}
		}
		return ns[i]
	case ArrayMapNode:
		for i := 0; i < len(n.Nodes); i += 2 {
			if n.Nodes[i] == x {
				return n.Nodes[i+1]
			}
		}
		return LiteralNode{nil}
	case MapNode:
		v, ok := n.Nodes[x]
		if !ok {
			return LiteralNode{nil}
		}
		return v
	default:
		if ln, ok := n.(LiteralNode); ok && ln.Value == nil {
			return LiteralNode{nil}
		}
		panic(errorf("could not get %s from %s", x, n))
	}
}

func conj(xs Node, x Node) Node {
	return xs.Conj(x)
}

func concat(ns ...Node) Node {
	out := []Node{}
	for _, n := range ns {
		out = append(out, seq(n)...)
	}
	return ListNode{out}
}

func cons(x Node, xs Node) Node {
	return ListNode{append([]Node{x}, seq(xs)...)}
}
