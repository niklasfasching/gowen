package main

import (
	"flag"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/niklasfasching/gowen"
	_ "github.com/niklasfasching/gowen/lib/core"
)

func main() {
	log.SetFlags(0) // do not prefix log with timestamp

	var in string
	flag.StringVar(&in, "eval", "", "Evaluate the input")
	flag.StringVar(&in, "e", "", "Evaluate the input")
	flag.Parse()
	env := gowen.NewEnv(false)
	evalFiles(flag.Args(), env)

	if in != "" {
		nodes, err := gowen.Parse(in)
		if err != nil {
			log.Fatal("ERROR:", err)
		}
		results, err := gowen.EvalMultiple(nodes, env)
		if err != nil {
			log.Fatal("ERROR:", err)
		}

		if len(results) != 0 {
			log.Println(results[len(results)-1])
		}
		return
	}
	repl()
}

func evalFiles(paths []string, env *gowen.Env) {
	input := ""
	for _, path := range paths {
		if filepath.Ext(path) == ".gow" {
			b, err := ioutil.ReadFile(path)
			if err != nil {
				panic(err)
			}
			input += string(b)
		}
	}
	nodes, err := gowen.Parse(input)
	if err != nil {
		log.Fatal("ERROR:", err)
	}
	gowen.EvalMultiple(nodes, env)
}
