package gowen

import "reflect"

func init() {
	Register(values, "(def version \"not even 0\")")
}

var values = map[string]Any{
	"if":         SpecialFn(fi),
	"def":        SpecialFn(def),
	"fn":         SpecialFn(newFn),
	"macro":      SpecialFn(newMacro),
	"try":        SpecialFn(try),
	"quote":      SpecialFn(quote),
	"quasiquote": MacroFn(quasiquote),

	"get": func(ns []Node, env *Env) Node {
		v := get(ns[0], ns[1])
		if ln, ok := v.(LiteralNode); ok && ln.Value == nil && len(ns) == 3 {
			return ns[2]
		}
		return v
	},

	"seq":    func(ns []Node, env *Env) Node { return ListNode{seq(ns[0])} },
	"cons":   func(ns []Node, env *Env) Node { return cons(ns[0], ns[1]) },
	"conj":   func(ns []Node, env *Env) Node { return conj(ns[0], ns[1]) },
	"concat": func(ns []Node, env *Env) Node { return concat(ns...) },
	"slice": func(ns []Node, env *Env) Node {
		t := reflect.TypeOf(0)
		i := reflect.ValueOf(ns[1].(LiteralNode).Value).Convert(t).Int()
		j := reflect.ValueOf(ns[2].(LiteralNode).Value).Convert(t).Int()
		return ListNode{ns[0].Seq()[i:j]}
	},

	"count": func(ns []Node, env *Env) Node { return LiteralNode{float64(len(seq(ns[0])))} },

	"macroexpand": func(ns []Node, env *Env) Node { return expand(ns, env)[0] },
	"parse":       func(in string) []Node { return parse(in) },
	"eval":        func(ns []Node, env *Env) Node { return eval(ns[0], env) },
	"apply":       func(ns []Node, env *Env) (Node, *Env, bool) { return apply(ns[0], seq(ns[1]), env) },
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
			return wrapInCall("vec", []Node{out}), false
		case ListNode:
			if lvl == 0 {
				return n, false
			}
			switch callTo(n) {
			case "quasiquote":
				return qq(n.Nodes[1], lvl+1)
			case "unquote", "unquote-splicing":
				lvl--
				assert(lvl >= 0, "call to unquote outside of quasiquote")
				assert(len(n.Nodes) == 2, "wrong number of arguments for unquote/unquote-splicing")
				qn, splicing := qq(n.Nodes[1], lvl)
				return qn, callTo(n) == "unquote-splicing" || splicing
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
	ln, isLn := eval(nodes[0], env).(LiteralNode)
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
	env.Set(sn.Value, eval(nodes[1], env))
	return LiteralNode{nil}, env, true
}

func buildFn(nodes []Node, defsideEnv *Env) (ComplexFn, *Env, string) {
	assert(len(nodes) >= 1, "wrong number of arguments for fn")
	fnEnv := ChildEnv(defsideEnv)
	name := "_"
	paramNodes := nodes[0]
	bodyNodes := nodes[1:]
	if sn, ok := nodes[0].(SymbolNode); ok {
		name = sn.Value
		paramNodes = nodes[1]
		bodyNodes = nodes[2:]
	}
	fn := func(argumentNodes []Node, _ *Env) (Node, *Env, bool) {
		env := ChildEnv(fnEnv)
		destructure(paramNodes, VectorNode{argumentNodes}, env)
		if len(bodyNodes) == 0 {
			return LiteralNode{nil}, env, true
		}
		for _, n := range bodyNodes[:len(bodyNodes)-1] {
			eval(n, env)
		}
		return bodyNodes[len(bodyNodes)-1], env, false
	}
	return fn, fnEnv, name
}

func newFn(nodes []Node, defsideEnv *Env) (Node, *Env, bool) {
	fn, fnEnv, name := buildFn(nodes, defsideEnv)
	fnEnv.Set(name, fn)
	return LiteralNode{fn}, defsideEnv, true
}

func newMacro(nodes []Node, defsideEnv *Env) (Node, *Env, bool) {
	fn, fnEnv, name := buildFn(nodes, defsideEnv)
	macroFn := MacroFn(func(ns []Node, env *Env) Node {
		n, env, isFinal := fn(ns, env)
		if !isFinal {
			n = eval(n, env)
		}
		return n
	})
	fnEnv.Set(name, macroFn)
	return LiteralNode{macroFn}, defsideEnv, true
}

func try(nodes []Node, parentEnv *Env) (node Node, env *Env, isFinal bool) {
	assert(len(nodes) >= 1, "wrong number of arguments for try")
	catch, ok := nodes[len(nodes)-1].(ListNode)
	body := nodes[:len(nodes)-1]
	assert(ok && callTo(catch) == "catch", "last form of try must be a catch clause")
	assert(len(catch.Nodes) >= 2, "invalid catch clause (inside try)")
	catchSymbol, catchBody := catch.Nodes[1], catch.Nodes[2:]
	sn, ok := catchSymbol.(SymbolNode)
	assert(ok, "catch clause must have symbol as first element")
	defer func() {
		if err := recover(); err != nil {
			env, isFinal = ChildEnv(parentEnv), true
			env.Set(sn.Value, err.(error).Error())
			for _, n := range catchBody {
				node = eval(n, env)
			}
		}
	}()
	env = ChildEnv(parentEnv)
	for _, n := range body {
		node = eval(n, env)
	}
	return node, env, true
}

func quote(nodes []Node, env *Env) (Node, *Env, bool) {
	assert(len(nodes) == 1, "wrong number of arguments for quote")
	return nodes[0], env, true
}
