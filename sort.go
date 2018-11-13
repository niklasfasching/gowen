package gowen

import (
	"fmt"
	"log"
)

func EvalTopological(nodes []Node, env *Env) Node {
	if len(nodes) == 0 {
		return LiteralNode{nil}
	}
	type defNode struct {
		node Node
		name string
		deps map[string]bool
	}

	nodes = Expand(nodes, env)
	bodyNodes := []Node{}
	defNodes := []defNode{}
	for _, n := range nodes {
		if CallTo(n) == "def" {
			deps := map[string]bool{}
			symbol := n.(ListNode).Nodes[1].(SymbolNode).Value
			for _, d := range getDependencies([]Node{n}) {
				if _, ok := env.Get(d); !ok {
					deps[d] = true
				}
			}
			defNodes = append(defNodes, defNode{n, symbol, deps})
		} else {
			bodyNodes = append(bodyNodes, n)
		}
	}
	if len(defNodes) == 0 {
		return EvalMultiple(bodyNodes, env)[len(bodyNodes)-1]
	}
	lenBefore := len(defNodes)
	for i := 0; i < len(defNodes); i++ {
		def := defNodes[i]
		if len(def.deps) == 0 {
			Eval(def.node, env)
			defNodes[i] = defNodes[len(defNodes)-1]
			defNodes = defNodes[:len(defNodes)-1]
			i = 0
			for _, def2 := range defNodes {
				delete(def2.deps, def.name)
			}
		}
	}
	if len(defNodes) == lenBefore {
		for _, d := range defNodes {
			log.Printf("%s depends on %v", d.name, d.deps)
		}
		panic(fmt.Sprintf("cyclic dependency detected"))

	}
	for _, def := range defNodes {
		bodyNodes = append(bodyNodes, def.node)
	}
	return EvalTopological(bodyNodes, env)
}

func getDependencies(nodes []Node) []string {
	deps := []string{}
	for _, n := range nodes {
		switch n := n.(type) {
		case ListNode:
			switch CallTo(n) {
			case "quote":
				continue
			case "fn", "macro":
				env := NewEnv(false)
				destructure(n.Nodes[1], VectorNode{}, env)
				for _, dep := range getDependencies(n.Nodes[2:]) {
					if _, ok := env.values[dep]; !ok {
						deps = append(deps, dep)
					}
				}
			case "def":
				deps = append(deps, getDependencies(n.Nodes[2:])...)
			default:
				deps = append(deps, getDependencies(n.Nodes)...)
			}
		case VectorNode:
			deps = append(deps, getDependencies(n.Nodes)...)
		case SymbolNode:
			deps = append(deps, n.Value)
		case LiteralNode, KeywordNode: // ignore
		default:
			panic("bad node (get deps)")
		}
	}
	return deps
}

func Expand(nodes []Node, env *Env) []Node {
	for i := 0; i < len(nodes); i++ {
		switch n := nodes[i].(type) {
		case VectorNode:
			n.Nodes = Expand(n.Nodes, env)
		case ArrayMapNode:
			for i := range n.Nodes {
				n.Nodes[i] = Expand([]Node{n.Nodes[i]}, env)[0]
			}
		case MapNode:
			en := MapNode{map[Node]Node{}}
			nodes[i] = en
			for k, v := range n.Nodes {
				k = Expand([]Node{k}, env)[0]
				v = Expand([]Node{v}, env)[0]
				en.Nodes[k] = v
			}
		case ListNode:
			vn, _ := env.Get(CallTo(n))
			ln, _ := vn.(LiteralNode)
			f, isMacro := ln.Value.(Macro)
			switch {
			case isMacro:
				nodes[i] = f(n.Nodes[1:], env)
				i--
			case CallTo(n) == "fn" || CallTo(n) == "macro":
				fnEnv := ChildEnv(env)
				destructure(n.Nodes[1], VectorNode{}, fnEnv)
				n.Nodes = Expand(n.Nodes, fnEnv)
			case CallTo(n) == "quote":
				continue
			default:
				n.Nodes = Expand(n.Nodes, env)
			}
		case SymbolNode, LiteralNode, KeywordNode:
			continue
		default:
			panic("bad node (expand)")
		}
	}
	return nodes
}
