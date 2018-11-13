package gowen

import (
	"log"
)

func EvalTopological(nodes []Node, env *Env) (_ Node, err error) {
	defer handleError(&err)
	return evalTopological(nodes, env), nil
}

func evalTopological(nodes []Node, env *Env) Node {
	if len(nodes) == 0 {
		return LiteralNode{nil}
	}
	type defNode struct {
		node Node
		name string
		deps map[string]bool
	}

	nodes = expand(nodes, env)
	bodyNodes := []Node{}
	defNodes := []defNode{}
	for _, n := range nodes {
		if callTo(n) == "def" {
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
		return evalMultiple(bodyNodes, env)[len(bodyNodes)-1]
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
		panic(errorf("cyclic dependency detected"))

	}
	for _, def := range defNodes {
		bodyNodes = append(bodyNodes, def.node)
	}
	return evalTopological(bodyNodes, env)
}

func getDependencies(nodes []Node) []string {
	deps := []string{}
	for _, n := range nodes {
		switch n := n.(type) {
		case ListNode:
			switch callTo(n) {
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
			panic(errorf("bad node (get deps): %s", n))
		}
	}
	return deps
}

func Expand(nodes []Node, env *Env) (_ []Node, err error) {
	defer handleError(&err)
	return expand(nodes, env), nil
}

func expand(nodes []Node, env *Env) []Node {
	for i := 0; i < len(nodes); i++ {
		switch n := nodes[i].(type) {
		case VectorNode:
			n.Nodes = expand(n.Nodes, env)
		case ArrayMapNode:
			for i := range n.Nodes {
				n.Nodes[i] = expand([]Node{n.Nodes[i]}, env)[0]
			}
		case MapNode:
			en := MapNode{map[Node]Node{}}
			nodes[i] = en
			for k, v := range n.Nodes {
				k = expand([]Node{k}, env)[0]
				v = expand([]Node{v}, env)[0]
				en.Nodes[k] = v
			}
		case ListNode:
			vn, _ := env.Get(callTo(n))
			ln, _ := vn.(LiteralNode)
			f, isMacro := ln.Value.(MacroFn)
			switch {
			case isMacro:
				nodes[i] = f(n.Nodes[1:], env)
				i--
			case callTo(n) == "fn" || callTo(n) == "macro":
				fnEnv := ChildEnv(env)
				destructure(n.Nodes[1], VectorNode{}, fnEnv)
				n.Nodes = expand(n.Nodes, fnEnv)
			case callTo(n) == "quote":
				continue
			default:
				n.Nodes = expand(n.Nodes, env)
			}
		case SymbolNode, LiteralNode, KeywordNode:
			continue
		default:
			panic(errorf("bad node (expand): %s", n))
		}
	}
	return nodes
}
