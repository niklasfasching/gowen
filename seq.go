package gowen

import (
	"fmt"
)

func seq(n Node) []Node {
	switch n := n.(type) {
	case VectorNode:
		return n.Nodes
	case ListNode:
		return n.Nodes
	case ArrayMapNode:
		ns := []Node{}
		for i := 0; i < len(n.Nodes); i += 2 {
			ns = append(ns, VectorNode{n.Nodes[i : i+2]})
		}
		return ns
	case MapNode:
		ns := []Node{}
		for k, v := range n.Nodes {
			ns = append(ns, VectorNode{[]Node{k, v}})
		}
		return ns
	case LiteralNode:
		if n.Value == nil {
			return []Node{}
		}
		if s, ok := n.Value.(string); ok {
			ns := make([]Node, len(s))
			for i, c := range s {
				ns[i] = LiteralNode{string(c)}
			}
			return ns
		}
		panic(fmt.Sprintf("don't know how to create seq from %#v", n))
	default:
		panic(fmt.Sprintf("don't know how to create seq from %#v", n))
	}
}

func get(n Node, x Node) Node {
	switch n := n.(type) {
	case ListNode, VectorNode:
		ns := seq(n)
		i := int(x.(LiteralNode).Value.(float64))
		if len(ns) <= i {
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
		panic(fmt.Sprintf("could not get %s from %s", x, n))
	}
}

func conj(xs Node, x Node) Node {
	if ln, ok := xs.(LiteralNode); ok && ln.Value == nil {
		return ListNode{[]Node{x}}
	}
	switch xs := xs.(type) {
	case ListNode:
		return ListNode{copyAppendNodes([]Node{x}, xs.Nodes...)}
	case VectorNode:
		return VectorNode{copyAppendNodes(xs.Nodes, x)}
	case ArrayMapNode:
		return ArrayMapNode{copyAppendNodes(xs.Nodes, x.(VectorNode).Nodes...)}
	case MapNode:
		m := map[Node]Node{}
		for k, v := range xs.Nodes {
			m[k] = v
		}
		kvs := x.(VectorNode).Nodes
		m[kvs[0]] = kvs[1]
		return MapNode{m}
	default:
		panic("bad conj")
	}
}
