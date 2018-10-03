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
type MapNode struct{ Nodes map[Node]Node }
type LiteralNode struct{ Value Any }
type SymbolNode struct{ Value string }
type KeywordNode struct{ Value string }

func Parse(input string) []Node { return parse(lex(input), []Node{}, "") }
func parse(l *lexer, ns []Node, inside string) []Node {
LOOP:
	for t := range l.tokens {
		switch t.category {
		case tokenParenOpen:
			ns = append(ns, ListNode{parse(l, []Node{}, "()")})
		case tokenBracketOpen:
			ns = append(ns, VectorNode{parse(l, []Node{}, "[]")})
		case tokenBraceOpen:
			cns := parse(l, []Node{}, "{}")
			assert(len(cns)%2 == 0, "hashmap must have an even number of elements (%s)", cns)
			m := map[Node]Node{}
			for i := 0; i < len(cns); i += 2 {
				m[cns[i]] = cns[i+1]
			}
			ns = append(ns, MapNode{m})
		case tokenKeyword:
			assert(len(t.string) > 1, "bad keyword")
			ns = append(ns, KeywordNode{t.string[1:]})
		case tokenSymbol:
			ns = append(ns, SymbolNode{t.string})
		case tokenQuote:
			ns = append(ns, wrapInCall("quote", parse(l, []Node{}, "'")))
		case tokenQuasiQuote:
			ns = append(ns, wrapInCall("quasiquote", parse(l, []Node{}, "'")))
		case tokenUnquote:
			ns = append(ns, wrapInCall("unquote", parse(l, []Node{}, "'")))
		case tokenUnquoteSplicing:
			ns = append(ns, wrapInCall("unquote-splicing", parse(l, []Node{}, "'")))
		case tokenString:
			unquoted, err := strconv.Unquote(strings.Replace(t.string, "\n", "\\n", -1))
			assert(err == nil, "cannot parse string from %v", t.string)
			ns = append(ns, LiteralNode{unquoted})
		case tokenFloat:
			float, err := strconv.ParseFloat(t.string, 64)
			assert(err == nil, "cannot parse float from %q", t.string)
			ns = append(ns, LiteralNode{float})
		case tokenError:
			assert(false, "parse error: %s", t.string)
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
			panic(Error{t, "bad token"})
		}
		if inside == "'" || inside == "~" {
			break LOOP
		}
	}
	return ns
}
