package gowen

import "reflect"

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
