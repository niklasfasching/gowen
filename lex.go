package gowen

// Here we implement a lexer as described in "Lexical Scanning In Go" by Rob Pike and used in text/template.
// In short we split the lexing step into many little lex fns (stateFn) that each do two things:
// 1. do something with the input (like emit a token)
// 2. return the next lex fn (or nil if there is nothing left to do) once they are done.
// As most tokens are delimited by space lexSpace ends up being our dispatching function.

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type token struct {
	category tokenCategory
	string   string
	index    int
}

type tokenCategory int

const (
	tokenError tokenCategory = iota
	tokenEOF
	tokenParenOpen
	tokenParenClose
	tokenBracketOpen
	tokenBracketClose
	tokenBraceOpen
	tokenBraceClose
	tokenSpace
	tokenSymbol
	tokenKeyword
	tokenFloat
	tokenString
	tokenUnquote
	tokenUnquoteSplicing
	tokenQuote
	tokenQuasiQuote
)

const eof = -1
const digits = "0123456789"

type stateFn func(*lexer) stateFn

type lexer struct {
	input  string
	index  int
	start  int
	width  int
	tokens chan token
}

func lex(input string) *lexer {
	l := &lexer{
		input:  input,
		tokens: make(chan token),
	}
	go func() {
		for state := lexSpace; state != nil; {
			state = state(l)
		}
		close(l.tokens)
	}()
	return l
}

func lexSpace(l *lexer) stateFn {
	l.acceptRun(", \t\n")
	l.ignore()
	switch r := l.next(); {
	case r == eof:
		l.emit(tokenEOF)
		return nil
	case r == '(':
		l.emit(tokenParenOpen)
		return lexSpace
	case r == ')':
		l.emit(tokenParenClose)
		return lexSpace
	case r == '[':
		l.emit(tokenBracketOpen)
		return lexSpace
	case r == ']':
		l.emit(tokenBracketClose)
		return lexSpace
	case r == '{':
		l.emit(tokenBraceOpen)
		return lexSpace
	case r == '}':
		l.emit(tokenBraceClose)
		return lexSpace
	case r == '"':
		return lexString
	case r == '\'':
		l.emit(tokenQuote)
		return lexSpace
	case r == '`':
		l.emit(tokenQuasiQuote)
		return lexSpace
	case ('0' <= r && r <= '9'):
		return lexNumber
	case r == '~':
		return lexUnquote
	case r == ';':
		return lexComment
	case r == ':':
		return lexKeyword
	case r == '+' || r == '-':
		if r2 := l.peek(); strings.ContainsRune(digits, r2) {
			return lexNumber
		}
		return lexSymbol
	case isValidIdentifierRune(r):
		return lexSymbol
	default:
		return l.errorf("bad rune: %q", r)
	}
}

func lexUnquote(l *lexer) stateFn {
	if r2 := l.peek(); r2 == '@' {
		l.next()
		l.emit(tokenUnquoteSplicing)
	} else {
		l.emit(tokenUnquote)
	}
	return lexSpace
}

func lexString(l *lexer) stateFn {
	for r := l.next(); r != '"'; r = l.next() {
		if r == '\\' {
			r = l.next()
		}
		if r == eof {
			return l.errorf("unterminated quoted string")
		}
	}
	l.emit(tokenString)
	return lexSpace
}

func lexComment(l *lexer) stateFn {
	offset := strings.Index(l.input[l.index:], "\n")
	if offset == -1 {
		return l.errorf("unterminated comment")
	}
	l.index += offset
	l.ignore()
	return lexSpace
}

func lexSymbol(l *lexer) stateFn {
	for r := l.next(); isValidIdentifierRune(r); r = l.next() {
	}
	l.backup()
	l.emit(tokenSymbol)
	return lexSpace
}

func lexKeyword(l *lexer) stateFn {
	for r := l.next(); isValidIdentifierRune(r); r = l.next() {
	}
	l.backup()
	l.emit(tokenKeyword)
	return lexSpace
}

func lexNumber(l *lexer) stateFn {
	l.accept("+-")
	l.acceptRun(digits)
	if l.accept(".") {
		l.acceptRun(digits)
	}
	if r := l.peek(); isValidIdentifierRune(r) {
		return lexSymbol
	}
	l.emit(tokenFloat)
	return lexSpace
}

func isValidIdentifierRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || strings.ContainsRune("?&@~<>=-+*/_#:.", r)
}

func (l *lexer) next() rune {
	if l.index >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.index:])
	l.width = w
	l.index += l.width
	return r
}

func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *lexer) backup() {
	l.index -= l.width
	_, w := utf8.DecodeRuneInString(l.input[l.index:])
	l.width = w
}

func (l *lexer) emit(c tokenCategory) {
	l.tokens <- token{c, l.input[l.start:l.index], l.start}
	l.start = l.index
}

func (l *lexer) ignore() {
	l.start = l.index
}

func (l *lexer) accept(valid string) bool {
	if strings.ContainsRune(valid, l.next()) {
		return true
	}
	l.backup()
	return false
}

func (l *lexer) acceptRun(valid string) {
	for strings.ContainsRune(valid, l.next()) {
	}
	l.backup()
}

func (l *lexer) errorf(format string, args ...Any) stateFn {
	l.tokens <- token{tokenError, fmt.Sprintf(format, args...), l.start}
	return nil
}
