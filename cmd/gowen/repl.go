package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/niklasfasching/gowen"
	"github.com/peterh/liner"
)

func repl() {
	gowen.AllowRedefine = true
	l := liner.NewLiner()
	historyFile := "/tmp/.gowen_history"
	defer l.Close()
	l.SetCtrlCAborts(true)
	if f, err := os.Open(historyFile); err == nil {
		l.ReadHistory(f)
		f.Close()
	}

	env := gowen.NewEnv()
	for in := ""; true; {
		if in == "" {
			fmt.Printf("> ")
		}
		if line, err := l.Prompt(""); err == nil {
			in += line
		} else if err == liner.ErrPromptAborted {
			in = ""
		} else if err == io.EOF {
			log.Print("Exit")
			break
		} else {
			log.Print("Error reading line", err)
			break
		}

		if isReady(in) {
			l.AppendHistory(in)
			evalPrint(in, env)
			in = ""
		}
	}

	if f, err := os.Create(historyFile); err != nil {
		log.Print("Error writing history file: ", err)
	} else {
		l.WriteHistory(f)
		f.Close()
	}
}

func evalPrint(expression string, env *gowen.Env) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("ERROR: %s\n", err)
		}
	}()
	nodes := gowen.Parse(expression)
	nodes = gowen.Expand(nodes, env)
	result := gowen.EvalMultiple(nodes, env)[len(nodes)-1]
	fmt.Printf("%v\n", result)
}

func isReady(expression string) bool {
	parens, squares, braces := 0, 0, 0
	for _, c := range expression {
		switch c {
		case '(':
			parens++
		case ')':
			parens--
		case '[':
			squares++
		case ']':
			squares--
		case '{':
			braces++
		case '}':
			braces--
		}
	}
	return parens == 0 && squares == 0 && braces == 0 && strings.TrimSpace(expression) != ""
}
