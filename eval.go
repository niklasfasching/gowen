package gowen

type Fn = func([]Node, *Env) Node
type MacroFn Fn

type ComplexFn = func([]Node, *Env) (Node, *Env, bool)
type SpecialFn ComplexFn

type Env struct {
	parent        *Env
	values        map[string]Any
	allowRedefine bool
}

var rootEnv = &Env{
	values: map[string]Any{
		"nil":   nil,
		"true":  true,
		"false": false,
	},
}

func NewEnv(allowRedefine bool) *Env { return &Env{rootEnv, nil, allowRedefine} }
func ChildEnv(parent *Env) *Env      { return &Env{parent, nil, parent.allowRedefine} }

func Register(m map[string]Any, ow string) {
	for k, v := range m {
		rootEnv.Set(k, v)
	}
	EvalMultiple(Parse(ow), rootEnv)
}

func (e *Env) Get(key string) (Node, bool) {
	v, exists := e.values[key]
	if exists {
		return FromGo(v), true
	}
	if !exists && e.parent != nil {
		return e.parent.Get(key)
	}
	return nil, false
}

func (e *Env) Set(key string, value Any) {
	if e.values == nil {
		e.values = map[string]Any{}
	}
	if key == "_" {
		return
	}
	_, exists := e.values[key]
	assert(e.allowRedefine || !exists, "must not redefine %s (%s)", key, value)
	e.values[key] = value
}

func (e *Env) IsTopLevel() bool {
	return e == rootEnv || e.parent == rootEnv
}

func EvalMultiple(nodes []Node, env *Env) []Node {
	results := make([]Node, len(nodes))
	for i, n := range nodes {
		results[i] = Eval(n, env)
	}
	return results
}

func handleEvalErr(n Node) {
	if err := recover(); err != nil {
		if _, ok := n.(ListNode); !ok {
			panic(err)
		}
		switch err := err.(type) {
		case Error:
			panic(err)
		case error:
			panic(Error{n, err.Error()})
		case string:
			panic(Error{n, err})
		default:
			panic(err)
		}
	}
}

func Eval(node Node, env *Env) Node {
	defer handleEvalErr(node)
	for {
		switch n := node.(type) {
		case LiteralNode, KeywordNode:
			return n
		case SymbolNode:
			vn, exists := env.Get(n.Value)
			assert(exists, "could not lookup symbol %q", n.Value)
			return vn
		case VectorNode:
			cns := make([]Node, len(n.Nodes))
			for i, cn := range n.Nodes {
				cns[i] = Eval(cn, env)
			}
			return VectorNode{cns}
		case ArrayMapNode:
			m := map[Node]Node{}
			for i := 0; i < len(n.Nodes); i += 2 {
				m[Eval(n.Nodes[i], env)] = Eval(n.Nodes[i+1], env)
			}
			return MapNode{m}
		case MapNode:
			m := map[Node]Node{}
			for k, v := range n.Nodes {
				m[Eval(k, env)] = Eval(v, env)
			}
			return MapNode{m}
		case ListNode:
			if len(n.Nodes) == 0 {
				return n
			}
			fln, ok := Eval(n.Nodes[0], env).(LiteralNode)
			assert(ok, "cannot use %s as a function", n.Nodes[0])
			argns, isFinal := n.Nodes[1:], false
			switch fn := fln.Value.(type) {
			case SpecialFn:
				node, env, isFinal = fn(argns, env)
			case MacroFn:
				node = fn(argns, env)
			default:
				node, env, isFinal = Apply(fln, EvalMultiple(argns, env), env)
			}
			if isFinal {
				return node
			}
		default:
			panic("cannot eval node")
		}
	}
}

func Apply(n Node, argns []Node, env *Env) (Node, *Env, bool) {
	fln, ok := n.(LiteralNode)
	assert(ok, "cannot use %s as a function", n)
	switch fn := fln.Value.(type) {
	case ComplexFn:
		n, env, isFinal := fn(argns, env)
		return n, env, isFinal
	case Fn:
		n := fn(argns, env)
		return n, env, true
	default:
		return applyInterop(fln, argns), env, true
	}
}
