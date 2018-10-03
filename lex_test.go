package gowen

import (
	"reflect"
	"testing"
)

type lexTest struct {
	name   string
	input  string
	tokens []token
}

var lexTests = []lexTest{
	{"empty", "", []token{token{tokenEOF, "", 0}}},

	{"space", "\n\t , , ,", []token{token{tokenEOF, "", 8}}},

	{"symbols: letters", "foo bar baz + - = a-b a/b a_b a.b", []token{
		token{tokenSymbol, "foo", 0},
		token{tokenSymbol, "bar", 4},
		token{tokenSymbol, "baz", 8},
		token{tokenSymbol, "+", 12},
		token{tokenSymbol, "-", 14},
		token{tokenSymbol, "=", 16},
		token{tokenSymbol, "a-b", 18},
		token{tokenSymbol, "a/b", 22},
		token{tokenSymbol, "a_b", 26},
		token{tokenSymbol, "a.b", 30},
		token{tokenEOF, "", 33},
	}},

	{"numbers", "-42 0 +42 -42.01 0.01 +42.01", []token{
		token{tokenFloat, "-42", 0},
		token{tokenFloat, "0", 4},
		token{tokenFloat, "+42", 6},
		token{tokenFloat, "-42.01", 10},
		token{tokenFloat, "0.01", 17},
		token{tokenFloat, "+42.01", 22},
		token{tokenEOF, "", 28},
	}},

	{"strings", `
        "foo"
        "foo\nbar"
        "multi
         line
         string"
    `, []token{
		token{tokenString, "\"foo\"", 9},
		token{tokenString, "\"foo\\nbar\"", 23},
		token{tokenString, "\"multi\n         line\n         string\"", 42},
		token{tokenEOF, "", 84},
	}},

	{"lists", "(+ 1 2)", []token{
		token{tokenParenOpen, "(", 0},
		token{tokenSymbol, "+", 1},
		token{tokenFloat, "1", 3},
		token{tokenFloat, "2", 5},
		token{tokenParenClose, ")", 6},
		token{tokenEOF, "", 7},
	}},

	{"vectors", "[1 2]", []token{
		token{tokenBracketOpen, "[", 0},
		token{tokenFloat, "1", 1},
		token{tokenFloat, "2", 3},
		token{tokenBracketClose, "]", 4},
		token{tokenEOF, "", 5},
	}},

	{"maps & keywords", "{:foo bar :baz 42}", []token{
		token{tokenBraceOpen, "{", 0},
		token{tokenKeyword, ":foo", 1},
		token{tokenSymbol, "bar", 6},
		token{tokenKeyword, ":baz", 10},
		token{tokenFloat, "42", 15},
		token{tokenBraceClose, "}", 17},
		token{tokenEOF, "", 18},
	}},

	{"quotes unquotes", "'(+ 2) 'x `y `~@[a ~b]", []token{
		token{tokenQuote, "'", 0},
		token{tokenParenOpen, "(", 1},
		token{tokenSymbol, "+", 2},
		token{tokenFloat, "2", 4},
		token{tokenParenClose, ")", 5},
		token{tokenQuote, "'", 7},
		token{tokenSymbol, "x", 8},
		token{tokenQuasiQuote, "`", 10},
		token{tokenSymbol, "y", 11},
		token{tokenQuasiQuote, "`", 13},
		token{tokenUnquoteSplicing, "~@", 14},
		token{tokenBracketOpen, "[", 16},
		token{tokenSymbol, "a", 17},
		token{tokenUnquote, "~", 19},
		token{tokenSymbol, "b", 20},
		token{tokenBracketClose, "]", 21},
		token{tokenEOF, "", 22},
	}},
}

func TestLex(t *testing.T) {
	for _, test := range lexTests {
		l := lex(test.input)
		tokens := []token{}
		for token := range l.tokens {
			// token.Index = 0
			tokens = append(tokens, token)
		}
		if !reflect.DeepEqual(tokens, test.tokens) {
			t.Errorf("%s: got\n\t%#v\nexpected\n\t%#v", test.name, tokens, test.tokens)
		}
	}
}
