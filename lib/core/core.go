package core

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"reflect"
)

type Any = interface{}

func assert(assertion bool, format string, vs ...Any) {
	if !assertion {
		panic(fmt.Sprintf(format, vs...))
	}
}

var Input = ""

// TODO: go generate concat *.gow files into a go file with string literal
func init() {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = build.Default.GOPATH
	}
	path := filepath.Join(gopath, "src", "github.com/niklasfasching/gowen/lib/core", "core.gow")
	bs, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	Input = string(bs)
}

var Values = map[string]Any{
	"=":   func(x1 Any, x2 Any) bool { return reflect.DeepEqual(x1, x2) },
	"<":   func(x1 float64, x2 float64) bool { return x1 < x2 },
	">":   func(x1 float64, x2 float64) bool { return x1 > x2 },
	"<=":  func(x1 float64, x2 float64) bool { return x1 <= x2 },
	">=":  func(x1 float64, x2 float64) bool { return x1 >= x2 },
	"mod": func(x1 float64, x2 float64) float64 { return float64(int(x1) % int(x2)) },
	"+":   func(vs ...float64) float64 { return calc(func(x, y float64) float64 { return x + y }, vs) },
	"-":   func(vs ...float64) float64 { return calc(func(x, y float64) float64 { return x - y }, vs) },
	"*":   func(vs ...float64) float64 { return calc(func(x, y float64) float64 { return x * y }, vs) },
	"/":   func(vs ...float64) float64 { return calc(func(x, y float64) float64 { return x / y }, vs) },
	"min": func(vs ...float64) float64 { return calc(func(x, y float64) float64 { return math.Min(x, y) }, vs) },
	"max": func(vs ...float64) float64 { return calc(func(x, y float64) float64 { return math.Max(x, y) }, vs) },

	"print":        func(args ...Any) { fmt.Println(args...) },
	"append":       func(xs []Any, x Any) Any { return append(xs, x) },
	"slice-list":   func(x []Any, i, j float64) Any { return x[int(i):int(j)] },
	"slice-string": func(x string, i, j float64) Any { return x[int(i):int(j)] },
	"nth":          func(slice []Any, i float64) Any { return slice[int(i)] },
	"throw":        func(template string, vs ...Any) { panic(fmt.Sprintf(template, vs...)) },

	"hashmap": func(kvs ...Any) Any {
		assert(len(kvs)%2 == 0, "hashmap must be called with even number of kvs")
		m := map[Any]Any{}
		for i := 0; i < len(kvs); i += 2 {
			m[kvs[i]] = kvs[i+1]
		}
		return m
	},
	"merge": func(m1 map[Any]Any, m2 map[Any]Any) Any {
		for k, v := range m2 {
			m1[k] = v
		}
		return m1
	},

	"format": func(format string, args ...Any) string { return fmt.Sprintf(format, args...) },
	"str": func(xs ...Any) string {
		s := ""
		for _, x := range xs {
			s += fmt.Sprintf("%v", x)
		}
		return s
	},

	"spit":  spit,
	"slurp": slurp,
}

func calc(fn func(float64, float64) float64, vs []float64) float64 {
	assert(len(vs) > 0, "wrong number of arguments for calc (+, -, ...)")
	acc := vs[0]
	for _, v := range vs[1:] {
		acc = fn(acc, v)
	}
	return acc
}
