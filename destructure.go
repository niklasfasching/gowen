package gowen

func destructure(binding Node, value Node, env *Env) {
	defer func() {
		if err := recover(); err != nil {
			panic(errorf("could not destructure %s to %s: %s", binding, value, err))
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
		panic(errorf("bad node for param %s %s", binding, value))
	}
}

func destructureSeq(binding Node, value Node, env *Env) {
	cbs := binding.Seq()
	cvs := value.Seq()
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
			destructure(cb, VectorNode{cvs}.Get(LiteralNode{i}), env)
		}
	}
}

func destructureMap(binding Node, value Node, env *Env) {
	vm := toMapNode(value)
	for _, vn := range binding.Seq() {
		vns := vn.(VectorNode).Nodes
		k, v := vns[0], vns[1]
		if kn, ok := k.(KeywordNode); ok && kn.Value == "as" {
			env.Set(v.(SymbolNode).Value, vm)
		} else if ok && kn.Value == "keys" {
			for _, n := range v.(VectorNode).Nodes {
				symbol := n.(SymbolNode).Value
				env.Set(symbol, vm.Get(KeywordNode{symbol}))
			}
		} else {
			destructure(k, vm.Get(v), env)
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
	case ListNode, VectorNode:
		return ArrayMapNode{n.Seq()}
	default:
		panic(errorf("cannot use %s as MapNode", n))
	}
}
