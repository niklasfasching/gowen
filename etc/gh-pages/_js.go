package main

import (
	"fmt"
	"syscall/js"

	"github.com/kr/pretty"
	"github.com/niklasfasching/gowen"
	_ "github.com/niklasfasching/gowen/lib/core"
)

func main() {
	env := gowen.NewEnv(false)
	js.Global().Call("gowenInitialized")

	document := js.Global().Get("document")
	in := document.Call("getElementById", "gowen-input")
	out := document.Call("getElementById", "gowen-output")
	js.Global().Set("gowenRun", js.FuncOf(func(js.Value, []js.Value) interface{} {
		nodes, err := gowen.Parse(in.Get("value").String())
		if err != nil {
			out.Set("textContent", fmt.Sprintf("%s", err))
			return nil
		}
		results, err := gowen.EvalMultiple(nodes, env)
		if err != nil {
			out.Set("textContent", fmt.Sprintf("%s", err))
			return nil
		}
		s := ""
		for _, result := range results {
			s += pretty.Sprint(result.ToGo()) + "\n"
		}
		out.Set("textContent", s)
		return nil
	}))

	<-make(chan struct{}) // stay alive
}
