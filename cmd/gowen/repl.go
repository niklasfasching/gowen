package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/niklasfasching/gowen"
	"github.com/peterh/liner"
)

func repl() {
	l := liner.NewLiner()
	historyFile := "/tmp/.gowen_history"
	defer l.Close()
	l.SetCtrlCAborts(true)
	if f, err := os.Open(historyFile); err == nil {
		l.ReadHistory(f)
		f.Close()
	}

	env := gowen.NewEnv(true)
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
	results := make(chan gowen.Node)
	errors := make(chan error)
	go func() {
		node, err := gowen.ParseAndEval(expression, env)
		if err != nil {
			errors <- err
			return
		}
		results <- node
	}()
	select {
	case err := <-errors:
		fmt.Printf("ERROR: %s\n", err)
	case node := <-results:
		fmt.Printf("%v\n", node)
	case <-time.After(3 * time.Second):
		env.Interrupt()
		err := <-errors
		fmt.Printf("ERROR: Timeout evaling %s - %s\n", expression, err)
	}
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
