package gowen

import (
	"reflect"
)

type Any = interface{}
type Vector []Any
type List []Any
type Map map[Any]Any

func applyInterop(fln LiteralNode, argns []Node) Node {
	fnv := reflect.ValueOf(fln.Value)
	fnt := fnv.Type()
	argvs := make([]reflect.Value, len(argns))
	for i, argn := range argns {
		if n := fnt.NumIn(); fnt.IsVariadic() && i >= n-1 {
			argvs[i] = reflectArg(argn.ToGo(), fnt.In(n-1).Elem())
		} else {
			argvs[i] = reflectArg(argn.ToGo(), fnt.In(i))
		}
	}
	switch retvs := fnv.Call(argvs); len(retvs) {
	case 0:
		return LiteralNode{nil}
	case 1:
		return FromGo(retvs[0].Interface())
	case 2:
		err := retvs[1].Interface()
		assert(err == nil, "call returned err: %s", err)
		return FromGo(retvs[0].Interface())
	default:
		panic(errorf("too many return values: %s", retvs))
	}
}

func reflectArg(arg Any, paramType reflect.Type) reflect.Value {
	defer func() {
		if err := recover(); err != nil {
			panic(errorf("reflectArg: converting %s to %s - %s", arg, paramType, err))
		}
	}()
	if arg == nil {
		switch paramType.Kind() {
		case reflect.Slice, reflect.Map, reflect.Func, reflect.Chan, reflect.Ptr, reflect.Interface:
			return reflect.Zero(paramType)
		default:
			return reflect.ValueOf((*Any)(nil))
		}
	}
	argValue := reflect.ValueOf(arg)
	argType := argValue.Type()
	switch {
	case argType.AssignableTo(paramType):
		return argValue
	case argType.ConvertibleTo(paramType):
		return argValue.Convert(paramType)
	case paramType.Kind() == reflect.Slice:
		paramElemType := paramType.Elem()
		l := argValue.Len()
		slice := reflect.MakeSlice(paramType, l, l)
		for i := 0; i < l; i++ {
			argValue.Index(i).Elem()
			slice.Index(i).Set(argValue.Index(i).Elem().Convert(paramElemType))
		}
		return slice
	case paramType.Kind() == reflect.Map:
		m := reflect.MakeMap(paramType)
		mValueType := paramType.Elem()
		mKeyType := paramType.Key()
		for _, k := range argValue.MapKeys() {
			m.SetMapIndex(k.Elem().Convert(mKeyType), argValue.MapIndex(k).Elem().Convert(mValueType))
		}
		return m
	default:
		return argValue
	}
}

func (n ListNode) ToGo() Any {
	values := make(List, len(n.Nodes))
	for i, cn := range n.Nodes {
		values[i] = cn.ToGo()
	}
	return values
}

func (n ArrayMapNode) ToGo() Any {
	m := make(Map, len(n.Nodes))
	for i := 0; i < len(n.Nodes); i += 2 {
		m[n.Nodes[i].ToGo()] = n.Nodes[i+1].ToGo()
	}
	return m
}

func (n MapNode) ToGo() Any {
	m := make(Map, len(n.Nodes))
	for k, v := range n.Nodes {
		m[k.ToGo()] = v.ToGo()
	}
	return m
}

func (n VectorNode) ToGo() Any {
	values := make(Vector, len(n.Nodes))
	for i, cn := range n.Nodes {
		values[i] = cn.ToGo()
	}
	return values
}

func (n LiteralNode) ToGo() Any { return n.Value }
func (n SymbolNode) ToGo() Any  { return n }
func (n KeywordNode) ToGo() Any { return n }

func ToGo(x Any) Any {
	switch x := x.(type) {
	case Node:
		return x.ToGo()
	default:
		return x
	}
}

func FromGo(x Any) Node {
	switch x := x.(type) {
	case Node:
		return x
	case List:
		nodes := make([]Node, len(x))
		for i, v := range x {
			nodes[i] = FromGo(v)
		}
		return ListNode{nodes}
	case Vector:
		nodes := make([]Node, len(x))
		for i, v := range x {
			nodes[i] = FromGo(v)
		}
		return VectorNode{nodes}
	case Map:
		m := map[Node]Node{}
		for k, v := range x {
			m[FromGo(k)] = FromGo(v)
		}
		return MapNode{m}
	case string, float64:
		return LiteralNode{x}
	case int:
		return LiteralNode{float64(x)}
	}

	switch reflect.ValueOf(x).Kind() {
	case reflect.Slice:
		xv := reflect.ValueOf(x)
		ns := make([]Node, xv.Len())
		for i := 0; i < xv.Len(); i++ {
			ns[i] = FromGo(xv.Index(i).Interface())
		}
		return ListNode{ns}
	case reflect.Map:
		xv := reflect.ValueOf(x)
		m := map[Node]Node{}
		for _, k := range xv.MapKeys() {
			m[FromGo(k.Interface())] = FromGo(xv.MapIndex(k).Interface())
		}
		return MapNode{m}
	default:
		return LiteralNode{x}
	}
}
