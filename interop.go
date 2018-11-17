package gowen

import (
	"reflect"
	"strings"
)

type Any = interface{}

func applyInterop(fln LiteralNode, argns []Node) Node {
	var retvs []reflect.Value
	if sn, ok := fln.Value.(SymbolNode); ok {
		retvs = applyMemberInterop(sn, argns)
	} else {
		fnv := reflect.ValueOf(fln.Value)
		fnt := fnv.Type()
		retvs = fnv.Call(reflectArgs(fnt, argns))
	}
	switch len(retvs) {
	case 0:
		return LiteralNode{nil}
	case 1:
		return ToNode(retvs[0].Interface())
	case 2:
		err := retvs[1].Interface()
		assert(err == nil, "call returned err: %s", err)
		return ToNode(retvs[0].Interface())
	default:
		panic(errorf("too many return values: %s", retvs))
	}
}

func applyMemberInterop(sn SymbolNode, argns []Node) []reflect.Value {
	it := reflect.ValueOf(argns[0].(LiteralNode).Value)
	itPointer := it
	if it.Kind() == reflect.Ptr {
		it = it.Elem()
	} else {
		itPointer = reflect.New(it.Type())
		itPointer.Elem().Set(it)
	}
	argns = argns[1:]
	name := strings.Title(sn.Value[1:])
	method := it.MethodByName(name)
	if !method.IsValid() {
		method = itPointer.MethodByName(name)
	}
	if method.IsValid() {
		return method.Call(reflectArgs(method.Type(), argns))
	}
	if field := it.FieldByName(name); field.IsValid() {
		return []reflect.Value{field}
	} else if field = itPointer.MethodByName(name); field.IsValid() {
		return []reflect.Value{field}
	}
	panic(errorf("member interop: %s is not a member of %s (%s)", name, it.Type(), argns))
}

func reflectArgs(fnt reflect.Type, argns []Node) []reflect.Value {
	argvs := make([]reflect.Value, len(argns))
	for i, argn := range argns {
		if n := fnt.NumIn(); fnt.IsVariadic() && i >= n-1 {
			argvs[i] = reflectArg(argn.ToGo(), fnt.In(n-1).Elem())
		} else {
			argvs[i] = reflectArg(argn.ToGo(), fnt.In(i))
		}
	}
	return argvs
}

func reflectArg(arg Any, paramType reflect.Type) reflect.Value {
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
	case paramType.Kind() == reflect.Ptr:
		pointer := reflect.New(paramType.Elem())
		pointer.Elem().Set(reflectArg(arg, paramType.Elem()))
		return pointer
	case argType.Kind() == reflect.Ptr:
		return reflectArg(argValue.Elem().Interface(), paramType)
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

func ToGo(x Any) Any {
	switch x := x.(type) {
	case Node:
		return x.ToGo()
	default:
		return x
	}
}

func ToNode(x Any) Node {
	switch x := x.(type) {
	case Node:
		return x
	default:
		return LiteralNode{x}
	}
}
