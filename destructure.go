package gowen

import (
	"fmt"
)

func destructure(binding Node, value Node, env *Env) {
	defer func() {
		if err := recover(); err != nil {
			panic(fmt.Sprintf("could not destructure %s to %s: %s", binding, value, err))
		}
	}()
	switch binding := binding.(type) {
	case SymbolNode:
		env.Set(binding.Value, value)
	case VectorNode, ListNode:
		destructureSeq(binding, value, env)
	case MapNode, ArrayMapNode:
		destructureMap(binding, value, env)
	default:
		panic(fmt.Sprintf("bad node for param %s %s", binding, value))
	}
}

func destructureSeq(binding Node, value Node, env *Env) {
	cbs := seq(binding)
	cvs := seq(value)
	for i := 0; i < len(cbs); i++ {
		cb := cbs[i]
		if kn, _ := cb.(KeywordNode); kn.Value == "as" {
			env.Set(cbs[i+1].(SymbolNode).Value, value)
			i++
		} else if sn, _ := cb.(SymbolNode); sn.Value == "&" {
			ln := ListNode{}
			if len(cvs) >= i {
				ln.Nodes = cvs[i:]
			}
			destructure(cbs[i+1], ln, env)
			i++
		} else {
			destructure(cb, get(VectorNode{cvs}, LiteralNode{float64(i)}), env)
		}
	}
}

func destructureMap(binding Node, value Node, env *Env) {
	vm := toMapNode(value)
	for _, vn := range seq(binding) {
		vns := vn.(VectorNode).Nodes
		k, v := vns[0], vns[1]
		if kn, ok := k.(KeywordNode); ok && kn.Value == "as" {
			env.Set(v.(SymbolNode).Value, vm)
		} else if ok && kn.Value == "keys" {
			for _, n := range v.(VectorNode).Nodes {
				symbol := n.(SymbolNode).Value
				env.Set(symbol, get(vm, KeywordNode{symbol}))
			}
		} else {
			destructure(k, get(vm, v), env)
		}
	}
}

func toMapNode(n Node) Node {
	if ln, ok := n.(LiteralNode); ok && ln.Value == nil {
		return ArrayMapNode{}
	}
	switch n := n.(type) {
	case MapNode, ArrayMapNode:
		return n
	case ListNode:
		return ArrayMapNode{n.Nodes}
	case VectorNode:
		return ArrayMapNode{n.Nodes}
	default:
		panic(fmt.Sprintf("cannot use %s as MapNode", n))
	}
}
