package gowen

import (
	"github.com/niklasfasching/gowen/lib/core"
)

func init() {
	Register(Values, "(def version \"not even 0\")")
	Register(core.Values, core.Input)
}

var Values = map[string]Any{
	"if":         SpecialFn(fi),
	"def":        SpecialFn(def),
	"fn":         SpecialFn(newFn),
	"macro":      SpecialFn(newMacro),
	"try":        SpecialFn(try),
	"quote":      SpecialFn(quote),
	"quasiquote": Macro(quasiquote),

	"list":   func(xs ...Any) []Any { return xs },
	"vector": func(xs []Any) Any { return Vector(xs) },
	"get": func(ns []Node, env *Env) (Node, *Env, bool) {
		v := get(ns[0], ns[1])
		if ln, ok := v.(LiteralNode); ok && ln.Value == nil && len(ns) == 3 {
			v = ns[2]
		}
		return v, env, true
	},
	"seq": func(ns []Node, env *Env) (Node, *Env, bool) { return ListNode{nodeSeq(ns[0])}, env, true },
	"count": func(ns []Node, env *Env) (Node, *Env, bool) {
		return LiteralNode{float64(len(nodeSeq(ns[0])))}, env, true
	},

	"macroexpand": func(ns []Node, env *Env) (Node, *Env, bool) { return Expand(ns, env)[0], env, true },
	"parse":       func(in string) []Node { return Parse(in) },
	"eval":        func(ns []Node, env *Env) (Node, *Env, bool) { return Eval(ns[0], env), env, true },
	"apply":       func(ns []Node, env *Env) (Node, *Env, bool) { return Apply(ns[0], nodeSeq(ns[1]), env) },

	"defn": Macro(func(ns []Node, _ *Env) Node {
		return wrapInCall("def", append([]Node{ns[0]}, wrapInCall("fn", ns[1:])))
	}),
	"defmacro": Macro(func(ns []Node, _ *Env) Node {
		return wrapInCall("def", append([]Node{ns[0]}, wrapInCall("macro", ns[1:])))
	}),
}

func quasiquote(nodes []Node, env *Env) Node {
	assert(len(nodes) == 1, "wrong number of arguments for quasiquote")
	var qq func(n Node, lvl int) (Node, bool)
	qq = func(n Node, lvl int) (Node, bool) {
		switch n := n.(type) {
		case LiteralNode:
			return n, false
		case SymbolNode:
			if lvl == 0 {
				return n, false
			}
			return wrapInCall("quote", []Node{n}), false
		case VectorNode:
			out := wrapInCall("concat", []Node{})
			for _, cn := range n.Nodes {
				qn, splicing := qq(cn, lvl)
				if splicing {
					out.Nodes = append(out.Nodes, qn)
				} else {
					out.Nodes = append(out.Nodes, wrapInCall("list", []Node{qn}))
				}
			}
			return wrapInCall("vector", []Node{out}), false
		case ListNode:
			if lvl == 0 {
				return n, false
			}
			switch CallTo(n) {
			case "quasiquote":
				return qq(n.Nodes[1], lvl+1)
			case "unquote", "unquote-splicing":
				lvl -= 1
				assert(lvl >= 0, "call to unquote outside of quasiquote")
				assert(len(n.Nodes) == 2, "wrong number of arguments for unquote/unquote-splicing")
				qn, splicing := qq(n.Nodes[1], lvl)
				return qn, CallTo(n) == "unquote-splicing" || splicing
			default:
				out := wrapInCall("concat", []Node{})
				for _, cn := range n.Nodes {
					qn, splicing := qq(cn, lvl)
					if splicing {
						out.Nodes = append(out.Nodes, qn)
					} else {
						out.Nodes = append(out.Nodes, wrapInCall("list", []Node{qn}))
					}
				}
				return out, false
			}
		default:
			panic("bad node in quasiquote")
		}
	}
	node, splicing := qq(nodes[0], 1)
	assert(!splicing, "cannot unquote-splice outside of a sequence")
	return node
}

func fi(nodes []Node, parentEnv *Env) (Node, *Env, bool) {
	env := ChildEnv(parentEnv)
	assert(len(nodes) >= 2, "wrong number of arguments for if")
	ln, isLn := Eval(nodes[0], env).(LiteralNode)
	if !isLn || (ln.Value != false && ln.Value != nil) {
		return nodes[1], env, false
	} else if len(nodes) == 3 {
		return nodes[2], env, false
	} else {
		return LiteralNode{nil}, env, true
	}
}

func def(nodes []Node, env *Env) (Node, *Env, bool) {
	assert(env.IsTopLevel(), "def must only be called from top level")
	assert(len(nodes) == 2, "wrong number of arguments for def")
	sn, ok := nodes[0].(SymbolNode)
	assert(ok, "def must be called with a symbol as the first argument")
	env.Set(sn.Value, Eval(nodes[1], env))
	return LiteralNode{nil}, env, true
}

func newFn(nodes []Node, defsideEnv *Env) (Node, *Env, bool) {
	assert(len(nodes) >= 1, "wrong number of arguments for fn")
	bodyNodes := nodes[1:]
	fn := func(paramNodes []Node, _ *Env) (Node, *Env, bool) {
		env := ChildEnv(defsideEnv)
		destructure(nodes[0], VectorNode{paramNodes}, env)
		if len(bodyNodes) == 0 {
			return LiteralNode{nil}, env, true
		}
		for _, n := range bodyNodes[:len(bodyNodes)-1] {
			Eval(n, env)
		}
		return bodyNodes[len(bodyNodes)-1], env, false
	}
	return LiteralNode{Fn(fn)}, defsideEnv, true
}

func newMacro(nodes []Node, defsideEnv *Env) (Node, *Env, bool) {
	ln, _, _ := newFn(nodes, defsideEnv)
	fn := ln.(LiteralNode).Value.(Fn)
	return LiteralNode{Macro(func(ns []Node, env *Env) Node {
		n, env, isFinal := fn(ns, env)
		if !isFinal {
			n = Eval(n, env)
		}
		return n
	})}, defsideEnv, true
}

func try(nodes []Node, parentEnv *Env) (node Node, _ *Env, _ bool) {
	assert(len(nodes) >= 1, "wrong number of arguments for try")
	catch, ok := nodes[len(nodes)-1].(ListNode)
	body := nodes[:len(nodes)-1]
	assert(ok && CallTo(catch) == "catch", "last form of try must be a catch clause")
	assert(len(catch.Nodes) >= 2, "invalid catch clause (inside try)")
	catchSymbol, catchBody := catch.Nodes[1], catch.Nodes[2:]
	sn, ok := catchSymbol.(SymbolNode)
	assert(ok, "catch clause must have symbol as first element")
	defer func() {
		if err := recover(); err != nil {
			env := ChildEnv(parentEnv)
			env.Set(sn.Value, err.(error).Error())
			for _, n := range catchBody {
				node = Eval(n, env)
			}
		}
	}()
	env := ChildEnv(parentEnv)
	for _, n := range body {
		node = Eval(n, env)
	}
	return node, parentEnv, true
}

func quote(nodes []Node, env *Env) (Node, *Env, bool) {
	assert(len(nodes) == 1, "wrong number of arguments for quote")
	return nodes[0], env, true
}
