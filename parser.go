package json

import (
	"errors"
	"fmt"
	"unicode/utf8"
)

const (
	None   = 0
	Null   = 'n'
	Bool   = 'b'
	String = 's'
	Array  = '['
	Object = '{'
	Number = 'N'
)

type (
	Parser struct{}
)

var (
	ErrEndOfBuffer = errors.New("unexpected end of buffer")
	ErrSyntax      = errors.New("syntax error")
	ErrBadString   = errors.New("bad string")
	ErrBadRune     = errors.New("bad rune")
	ErrType        = errors.New("incompatible type")
)

func (p *Parser) Type(b []byte, st int) (tp byte, i int, err error) {
	for i = st; i < len(b); i++ {
		switch b[i] {
		case ' ', '\t', '\n', '\v',
			',', ':':
			continue
		case 't', 'f':
			return Bool, i, nil
		case '"':
			return String, i, nil
		case Null, Array, Object:
			return b[i], i, nil
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
			'+', '-', '.',
			'N',      // NaN
			'i', 'I': // inf
			return Number, i, nil
		}

		return None, i, ErrSyntax
	}

	return None, i, ErrEndOfBuffer
}

func (p *Parser) Skip(b []byte, st int) (i int, err error) {
	i, _, err = p.break_(b, st, 0)
	return
}

func (p *Parser) Raw(b []byte, st int) (v []byte, i int, err error) {
	_, st, err = p.Type(b, st)
	if err != nil {
		return nil, st, err
	}

	i, _, err = p.break_(b, st, 0)
	if err != nil {
		return
	}

	return b[st:i], i, nil
}

func (p *Parser) Break(b []byte, st, depth int) (i int, err error) {
	i, _, err = p.break_(b, st, depth)
	return
}

func (p *Parser) Key(b []byte, st int) (k []byte, i int, err error) {
	tp, i, err := p.Type(b, st)
	if err != nil {
		return
	}

	if tp != String {
		return nil, i, ErrType
	}

	raw, i, err := p.Raw(b, i)
	if err != nil {
		return
	}

	return raw[1 : len(raw)-1], i, nil
}

func (p *Parser) DecodeString(b []byte, st int, buf []byte) (s []byte, i int, err error) {
	tp, i, err := p.Type(b, st)
	if err != nil {
		return
	}

	if tp != String {
		return nil, i, ErrType
	}

	if buf == nil {
		buf = []byte{}
	}

	s, _, i, err = p.decodeString(b, i, buf)

	return
}

func (p *Parser) DecodedStringLength(b []byte, st int) (n, i int, err error) {
	tp, i, err := p.Type(b, st)
	if err != nil {
		return
	}

	if tp != String {
		return 0, i, ErrType
	}

	_, n, i, err = p.decodeString(b, i, nil)

	return
}

func (p *Parser) break_(b []byte, st, depth int) (i, maxDepth int, err error) {
	i = st
	d := depth
	maxDepth = depth

	for i < len(b) {
		switch b[i] {
		case ' ', '\t', '\n', '\v',
			',', ':':
			i++
			continue
		case '"':
			i, err = p.skipString(b, i)
		case 'n', 't', 'f':
			i, err = p.skipLit(b, i)
		case '[', '{':
			i++
			d++

			if d > maxDepth {
				maxDepth = d
			}
		case ']', '}':
			i++
			d--
		default:
			i, err = p.skipNum(b, i)
		}

		if err != nil {
			return
		}

		if d == 0 {
			return
		}
	}

	return i, maxDepth, ErrEndOfBuffer
}

func (p *Parser) Enter(b []byte, st int, typ byte) (i int, err error) {
	if typ != Array && typ != Object {
		return st, ErrType
	}

	tp, i, err := p.Type(b, st)
	if err != nil {
		return
	}

	if tp == typ {
		i++
		return
	}

	return i, ErrType
}

func (p *Parser) More(b []byte, st int, typ byte) (more bool, i int, err error) {
	for i = st; i < len(b); i++ {
		switch b[i] {
		case ' ', '\n', '\t', '\v',
			',':
			continue
		}

		break
	}

	if i == len(b) {
		return false, i, ErrEndOfBuffer
	}

	if b[i] == typ+2 {
		i++
		return false, i, nil
	}

	if typ != Object {
		return true, i, nil
	}

	tp, i, err := p.Type(b, i)
	if err != nil {
		return false, i, err
	}

	if tp != String {
		return false, i, ErrSyntax
	}

	return true, i, nil
}

func (p *Parser) ForMore(b []byte, i *int, typ byte, errp *error) bool {
	more, j, err := p.More(b, *i, typ)
	*i = j

	if errp != nil {
		*errp = err
	}

	return more
}

func (p *Parser) Length(b []byte, st int) (n, i int, err error) {
	tp, i, err := p.Type(b, st)
	if err != nil {
		return 0, i, err
	}

	switch tp {
	case String:
		_, n, i, err = p.decodeString(b, i, nil)
		return
	case Array, Object:
	default:
		return 0, i, ErrType
	}

	i, err = p.Enter(b, i, tp)
	if err != nil {
		return 0, i, err
	}

	for p.ForMore(b, &i, tp, &err) {
		if tp == Object {
			i, err = p.Skip(b, i)
			if err != nil {
				return n, i, err
			}
		}

		i, err = p.Skip(b, i)
		if err != nil {
			return n, i, err
		}

		n++
	}
	if err != nil {
		return n, i, err
	}

	return n, i, nil
}

func (p *Parser) SkipSpaces(b []byte, i int) int {
	for i < len(b) && (b[i] == ' ' || b[i] == '\n' || b[i] == '\t') {
		i++
	}

	return i
}

func (p *Parser) skipString(b []byte, st int) (i int, err error) {
	i = st + 1 // open "

	for i < len(b) {
		switch b[i] {
		case '"':
			return i + 1, nil
		case '\\':
			i++
			if i == len(b) {
				return i, ErrEndOfBuffer
			}

			wid := 0

			switch b[i] {
			case '\\', 'n', 't', '"':
			case 'x':
				wid = 2
			case 'u':
				wid = 4
			case 'U':
				wid = 8
			default:
				return i, ErrBadString
			}

			i++

			if i+wid > len(b) {
				return i - 2, ErrEndOfBuffer
			}

			i += wid
		case '\n':
			return i, ErrBadString
		default:
			if b[i] < utf8.RuneSelf {
				i++
				break
			}

			r, w := utf8.DecodeRune(b[i:])
			if r == utf8.RuneError {
				return i, ErrBadRune
			}

			i += w
		}
	}

	return i, ErrEndOfBuffer
}

func (p *Parser) decodeString(b []byte, st int, w []byte) (_ []byte, n, i int, err error) {
	i = st + 1 // open "
	done := i

	add := func(d []byte, s ...byte) []byte {
		if w == nil {
			return nil
		}

		return append(w, s...)
	}

	for i < len(b) {
		switch b[i] {
		case '"':
			w = add(w, b[done:i]...)
			i++
			return w, n, i, err
		case '\\':
			w = add(w, b[done:i]...)

			i++
			if i == len(b) {
				return w, n, i, ErrEndOfBuffer
			}

			switch b[i] {
			case '\\':
				w = add(w, '\\')
			case '"':
				w = add(w, '"')
			case 'n':
				w = add(w, '\n')
			case 't':
				w = add(w, '\t')
			case 'x', 'u', 'U':
				w, i, err = decodeRune(w, b, i)
				if err != nil {
					return w, n, i, err
				}

				i--
			default:
				return w, n, i, ErrBadString
			}

			i++
			n++
			done = i
		case '\n':
			return w, n, i, ErrBadString
		default:
			if b[i] < utf8.RuneSelf {
				i++
				n++
				break
			}

			r, wid := utf8.DecodeRune(b[i:])
			if r == utf8.RuneError {
				return w, n, i, ErrBadRune
			}

			i += wid
			n++
		}
	}

	return w, n, i, ErrEndOfBuffer
}

func (p *Parser) skipLit(b []byte, st int) (i int, err error) {
	switch b[st] {
	case 't':
		return p.skipVal(b, st, "true")
	case 'f':
		return p.skipVal(b, st, "false")
	case 'n':
		return p.skipVal(b, st, "null")
	default:
		panic(fmt.Sprintf("%q", b[st]))
	}
}

func (p *Parser) skipVal(b []byte, st int, val string) (i int, err error) {
	end := st + len(val)
	if end <= len(b) && string(b[st:end]) == val {
		return end, nil
	}

	if end > len(b) && string(b[st:]) == val[:len(b)-st] {
		return len(b), ErrEndOfBuffer
	}

	return st, ErrSyntax
}

func (p *Parser) skipNum(b []byte, st int) (i int, err error) {
	i = st

	// NaN

	if i+3 < len(b) && string(b[i:i+3]) == "NaN" {
		return i + 3, nil
	}

	// sign
	i = skipSign(b, i)

	// infinity

	if i+3 < len(b) && (b[i] == 'i' || b[i] == 'I') && string(b[i+1:i+3]) == "nf" {
		i += 3

		if i+5 < len(b) && string(b[i:i+5]) == "inity" {
			i += 5
		}

		return i, nil
	}

	// 0x

	hex := false

	if i+2 < len(b) && b[i] == '0' && (b[i+1] == 'x' || b[i+1] == 'X') {
		hex = true
		i += 2
	}

	// integer

	digit := false

	for i < len(b) && (b[i] >= '0' && b[i] <= '9' || hex && (b[i] >= 'a' && b[i] <= 'f' || b[i] >= 'A' && b[i] <= 'F')) {
		digit = true
		i++
	}

	tail := false

	switch {
	case i == len(b):
	case b[i] == '.':
		i++
		i, tail = skipDec(b, i)
	case b[i] == 'e' || b[i] == 'E':
		i++
		i = skipSign(b, i)
		i, tail = skipDec(b, i)
	case b[i] == 'p' || b[i] == 'P':
		i++
		i = skipSign(b, i)
		i, tail = skipDec(b, i)
	default:
	}

	if !(digit || tail) {
		return st, ErrSyntax
	}

	return i, nil
}

func SkipSpaces(b []byte, i int) int {
	return (&Parser{}).SkipSpaces(b, i)
}

func skipSign(b []byte, i int) int {
	if i < len(b) && (b[i] == '+' || b[i] == '-') {
		i++
	}

	return i
}

func skipDec(b []byte, i int) (int, bool) {
	ok := false

	for i < len(b) && b[i] >= '0' && b[i] <= '9' {
		ok = true
		i++
	}

	return i, ok
}

func decodeRune(w, r []byte, i int) ([]byte, int, error) {
	wid := 0

	switch r[i] {
	case 'x':
		wid = 2
	case 'u':
		wid = 4
	case 'U':
		wid = 8
	default:
		panic(r[i])
	}

	i++

	if i+wid > len(r) {
		return w, i - 2, ErrEndOfBuffer
	}

	var x rune

	sym := func(c byte) bool {
		switch {
		case c >= '0' && c <= '9':
			c -= '0'
		case c >= 'a' && c <= 'f':
			c = 10 + c - 'a'
		case c >= 'A' && c <= 'F':
			c = 10 + c - 'A'
		default:
			return false
		}

		x = x<<4 | rune(c)

		return true
	}

	for j := 0; j < wid; j += 2 {
		if !sym(r[i+j]) || !sym(r[i+j+1]) {
			return w, i - 2, ErrBadRune
		}
	}

	w = utf8.AppendRune(w, x)

	return w, i + wid, nil
}
