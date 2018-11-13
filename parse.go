package gowen

import (
	"strconv"
	"strings"
)

type Node interface {
	ToGo() Any
	String() string
}

type ListNode struct{ Nodes []Node }
type VectorNode struct{ Nodes []Node }
type ArrayMapNode struct{ Nodes []Node }
type MapNode struct{ Nodes map[Node]Node }
type LiteralNode struct{ Value Any }
type SymbolNode struct{ Value string }
type KeywordNode struct{ Value string }

// Parse reads the input string into an AST (list of nodes).
// Note that literal maps are read into ArrayMapNode, not MapNode. This allows
// unhashable nodes (nodes containing a slice or map) to be used as map keys.
// Only with this do nested associative destructuring and computed keys (e.g. {(+ 1 2) 3})
// become possible. ArrayMapNode only exists until the first evaluation after which it
// becomes a normal MapNode.
func Parse(input string) (nodes []Node, err error) {
	defer handleError(&err)
	return parse(input), nil
}

func parse(input string) []Node { return parseLoop(lex(input), []Node{}, "") }
func parseLoop(l *lexer, ns []Node, inside string) []Node {
LOOP:
	for t := range l.tokens {
		switch t.category {
		case tokenParenOpen:
			ns = append(ns, ListNode{parseLoop(l, []Node{}, "()")})
		case tokenBracketOpen:
			ns = append(ns, VectorNode{parseLoop(l, []Node{}, "[]")})
		case tokenBraceOpen:
			cns := parseLoop(l, []Node{}, "{}")
			assert(len(cns)%2 == 0, "hashmap must have an even number of elements (%s)", cns)
			ns = append(ns, ArrayMapNode{cns})
		case tokenKeyword:
			assert(len(t.string) > 1, "bad keyword")
			ns = append(ns, KeywordNode{t.string[1:]})
		case tokenSymbol:
			ns = append(ns, SymbolNode{t.string})
		case tokenQuote:
			ns = append(ns, wrapInCall("quote", parseLoop(l, []Node{}, "'")))
		case tokenQuasiQuote:
			ns = append(ns, wrapInCall("quasiquote", parseLoop(l, []Node{}, "'")))
		case tokenUnquote:
			ns = append(ns, wrapInCall("unquote", parseLoop(l, []Node{}, "'")))
		case tokenUnquoteSplicing:
			ns = append(ns, wrapInCall("unquote-splicing", parseLoop(l, []Node{}, "'")))
		case tokenString:
			unquoted, err := strconv.Unquote(strings.Replace(t.string, "\n", "\\n", -1))
			assert(err == nil, "cannot parseLoop string from %v", t.string)
			ns = append(ns, LiteralNode{unquoted})
		case tokenFloat:
			float, err := strconv.ParseFloat(t.string, 64)
			assert(err == nil, "cannot parseLoop float from %q", t.string)
			ns = append(ns, LiteralNode{float})
		case tokenError:
			assert(false, "parseLoop error: %s", t.string)
		case tokenEOF:
			assert(inside == "", "unexpected EOF")
			break LOOP
		case tokenParenClose:
			assert(inside == "()", "unexpected ) %v", t)
			break LOOP
		case tokenBracketClose:
			assert(inside == "[]", "unexpected ]")
			break LOOP
		case tokenBraceClose:
			assert(inside == "{}", "unexpected }")
			break LOOP
		default:
			panic(Error{t, errorf("bad token")})
		}
		if inside == "'" || inside == "~" {
			break LOOP
		}
	}
	return ns
}
