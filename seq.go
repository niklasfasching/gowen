package gowen

import (
	"fmt"
	"reflect"
)

func nodeSeq(n Node) []Node {
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
		xs := anySeq(n.Value)
		ns := make([]Node, len(xs))
		for i, x := range xs {
			ns[i] = FromGo(x)
		}
		return ns
	default:
		panic(fmt.Sprintf("don't know how to create seq from %#v", n))
	}
}

func anySeq(x Any) (xs []Any) {
	switch vx := reflect.ValueOf(x); vx.Kind() {
	case reflect.Slice:
		for i := 0; i < vx.Len(); i++ {
			xs = append(xs, vx.Index(i).Interface())
		}
	case reflect.Map:
		for _, vk := range vx.MapKeys() {
			k := vk.Interface()
			v := vx.MapIndex(vk).Interface()
			xs = append(xs, []Any{k, v})
		}
	case reflect.String:
		for _, c := range x.(string) {
			xs = append(xs, string(c))
		}
	default:
		panic(fmt.Sprintf("don't know how to create seq from %#v", x))
	}
	return xs
}

func get(n Node, x Node) Node {
	switch n := n.(type) {
	case ListNode, VectorNode:
		ns := nodeSeq(n)
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
