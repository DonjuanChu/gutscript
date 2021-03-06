package gutscript

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type stateFn func(*GutsLex) stateFn

type GutsLex struct {
	// the line
	input string
	line  int
	start int
	pos   int
	state stateFn

	// rollback position
	rollbackPos int

	// current space prefix length (used to calculate indentation level)
	space     int
	lastSpace int

	lastTokenType TokenType

	// XXX
	width int
	items chan *GutsSymType
}

type TokenType int

const eof = -1
const lexError = -99

func (self *GutsSymType) String() string {
	switch self.typ {
	case T_EOF, eof:
		return "EOF"
	}
	return fmt.Sprintf("Token: %s %q", GetTokenName(int(self.typ)), self.val)
}

// remember rollback position
func (l *GutsLex) remember() int {
	l.rollbackPos = l.pos
	return l.pos
}

func (l *GutsLex) rollback() {
	l.pos = l.rollbackPos
}

func (l *GutsLex) error(msg string) {
	fmt.Println("Syntax Error", msg)
	fmt.Println("Line", l.line)
	fmt.Println("Pos", l.pos)
	fmt.Println("Input", l.input)
}

func (l *GutsLex) emit(t TokenType) {

	l.lastTokenType = t
	l.items <- &GutsSymType{
		typ:  t,
		val:  l.input[l.start:l.pos],
		pos:  l.start,
		line: l.line,
	}
	l.start = l.pos
}

// peek returns but does not consume
// the next rune in the input.
func (l *GutsLex) peek() (r rune) {
	r = l.next()
	l.backup()
	return r
}

func (l *GutsLex) peekMore(p int) (r rune) {
	var w = 0
	for i := p; i > 0; i-- {
		r = l.next()
		w += l.width
	}
	l.pos -= w
	return r
}

// ignore skips over the pending input before this point.
func (l *GutsLex) ignore() {
	l.start = l.pos
}

// accept consumes the next rune
// if it's from the valid set.
// e.g.,
//    l.accept("1234567890")
func (l *GutsLex) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

// backup steps back one rune.
// Can be called only once per call of next.
func (l *GutsLex) backup() {
	l.pos -= l.width
}

// next returns the next rune in the input.
func (l *GutsLex) next() (r rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

// get the last rune in the input
func (l *GutsLex) last() (r rune) {
	_, w := utf8.DecodeRuneInString(l.input[l.pos-l.width:])
	r, _ = utf8.DecodeRuneInString(l.input[l.pos-l.width-w:])
	return r
}

func (l *GutsLex) run() {
	for l.state = lexStart; l.state != nil; {
		fn := l.state(l)
		if fn != nil {
			l.state = fn
		} else {
			break
		}
	}
	l.items <- nil
}

func (l *GutsLex) close() {
	close(l.items)
}

func (l *GutsLex) emitIfMatch(str string, typ TokenType) bool {
	if l.consumeIfMatch(str) {
		l.emit(typ)
		return true
	}
	return false
}

func (l *GutsLex) consumeIfMatch(str string) bool {
	var c rune
	var width = 0
	for _, sc := range str {
		c = l.next()
		width += l.width
		if sc != c {
			l.pos -= width
			return false
		}
	}
	return true
}

/*
lookahead match method
*/
func (l *GutsLex) match(str string) bool {
	var c rune
	var width = 0
	for _, sc := range str {
		c = l.next()
		width += l.width
		if sc != c {
			l.pos -= width
			return false
		}
	}
	l.pos -= width
	return true
}

// set token in lval, return the token type id
func (l *GutsLex) Lex(lval *GutsSymType) int {
	var item *GutsSymType
	item = <-l.items
	if item != nil {
		*lval = *item
		// fmt.Println("Lex", item.String())
		// fmt.Printf("%s %s", GutsTokname(int(item.typ)), item.val)
		return int(item.typ)
	}
	return 0
}

func GetTokenName(typ int) string {
	var c int = typ
	if c >= GutsPrivate {
		if c < GutsPrivate+len(GutsTok2) {
			c = GutsTok2[c-GutsPrivate]
		}
	}
	return GutsTokname(c)
}

func (l *GutsLex) Error(s string) {
	fmt.Printf("syntax error: %s\n", s)
}

func dumpLexItems(items chan *GutsSymType) {
	for {
		item := <-items
		if item == nil {
			break
		}
		fmt.Printf("Got token %s: '%s'\n", GetTokenName(int(item.typ)), item.val)
		if item.typ == T_EOF || item.typ == eof {
			break
		}
	}
}
