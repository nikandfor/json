package json

import (
	"errors"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

// Value types returned by Parser.
const (
	None   = 0 // never returned in successful case
	Null   = 'n'
	Bool   = 'b'
	String = 's'
	Array  = '['
	Object = '{'
	Number = 'N'
)

type (
	// Parser is a group of methods to parse JSON buffers.
	// Parser is stateless.
	// All the needed state is passed though arguments and return values.
	//
	// Most of the methods take buffer with json and start position
	// and return a value, end position and possible error.
	Parser struct{}
)

var whitespaces uint64

func init() {
	for _, b := range []byte{'\n', '\r', '\t', ' '} {
		whitespaces |= 1 << b
	}
}

// Errors returned by Parser.
var (
	ErrBadNumber   = errors.New("bad number")
	ErrBadRune     = errors.New("bad rune")
	ErrBadString   = errors.New("bad string")
	ErrEndOfBuffer = errors.New("unexpected end of buffer")
	ErrSyntax      = errors.New("syntax error")
	ErrType        = errors.New("incompatible type")
)

// Type finds the beginning of the next value and detects its type.
// It doesn't parse the value so it can't detect if it's incorrect.
func (p *Parser) Type(b []byte, st int) (tp byte, i int, err error) {
	for i = st; i < len(b); i++ {
		if isWhitespace(b[i]) {
			continue
		}

		switch b[i] {
		case ',', ':':
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
			'i', 'I': // Inf
			return Number, i, nil
		}

		return None, i, ErrSyntax
	}

	return None, i, ErrEndOfBuffer
}

// Skip skips the next value.
func (p *Parser) Skip(b []byte, st int) (i int, err error) {
	i, _, err = p.break_(b, st, 0)
	return
}

// Raw skips the next value and returns subslice with the value trimming whitespaces.
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

// Break breaks from inside the object to the end of it on depth levels.
// As a special case with depth=0 it skips the next value.
// Skip and Raw do exactly that.
//
// It's intended for exiting out of arrays and objects when their content is not needed anymore
// (all the needed indexes or keys are already parsed) and we want to parse the next array or object.
func (p *Parser) Break(b []byte, st, depth int) (i int, err error) {
	i, _, err = p.break_(b, st, depth)
	return
}

// Key reads the next string removing quotes but not decoding the string value.
// So escape sequences (\n, \uXXXX) are not decoded. They are returned as is.
// This is intended for object keys as they usually contain alpha-numeric symbols only.
// This is faster and does not require additional buffer for decoding.
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

// DecodeString reads the next string, decodes escape sequences (\n, \uXXXX),
// and appends the result to the buf.
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

// DecodedStringLength reads and decodes the next string but only return the result length.
// It doesn't allocate while DecodeString does.
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
		if isWhitespace(b[i]) {
			i++
			continue
		}

		switch b[i] {
		case ',', ':':
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

// Enter enters an Array or an Object. typ is checked to match with the actual container type.
// Use More or, more convenient form, ForMore to iterate over container.
// See examples to understand the usage pattern more.
func (p *Parser) Enter(b []byte, st int, typ byte) (i int, err error) {
	tp, i, err := p.Type(b, st)
	if err != nil {
		return
	}

	if tp != typ || typ != Array && typ != Object {
		return i, ErrType
	}

	i++

	return
}

// More iterates over an Array or an Object elements entered by the Enter method.
func (p *Parser) More(b []byte, st int, typ byte) (more bool, i int, err error) {
	for i = st; i < len(b); i++ {
		if isWhitespace(b[i]) || b[i] == ',' {
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

	if typ == Array {
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

// ForMore is a convenient wrapper for More which makes iterating code shorter and simpler.
func (p *Parser) ForMore(b []byte, i *int, typ byte, errp *error) bool {
	more, j, err := p.More(b, *i, typ)
	*i = j

	if errp != nil {
		*errp = err
	}

	return more
}

// Length calculates String length (runes in decoded form) or number of elements in Array or Object.
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
			_, i, err = p.Key(b, i)
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

// SkipSpaces skips whitespaces.
func (p *Parser) SkipSpaces(b []byte, i int) int {
	for i < len(b) && isWhitespace(b[i]) {
		i++
	}

	return i
}

func (p *Parser) skipString(b []byte, st int) (i int, err error) {
	i = st + 1 // opening "

	for i < len(b) {
		switch b[i] {
		case '"':
			return i + 1, nil
		case '\\':
			i++
			if i == len(b) {
				return i, ErrEndOfBuffer
			}

			size := 0

			switch b[i] {
			case '"', '\\', '/', 'n', 'r', 't', 'b', 'f':
			case 'x':
				size = 2
			case 'u':
				size = 4
			case 'U':
				size = 8
			default:
				return i, ErrBadString
			}

			i++

			if i+size > len(b) {
				return i - 2, ErrEndOfBuffer
			}

			i += size
		case '\n', '\r', '\b':
			return i, ErrBadString
		default:
			if b[i] < utf8.RuneSelf {
				i++
				break
			}

			_, size := utf8.DecodeRune(b[i:])
			//	if r == utf8.RuneError && size == 1 {
			//		return i, ErrBadRune
			//	}

			i += size
		}
	}

	return i, ErrEndOfBuffer
}

func (p *Parser) decodeString(b []byte, st int, w []byte) (_ []byte, n, i int, err error) {
	i = st + 1 // opening "
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
			case '\\', '"', '/':
				w = add(w, b[i])
			case 'n':
				w = add(w, '\n')
			case 'r':
				w = add(w, '\r')
			case 't':
				w = add(w, '\t')
			case 'b':
				w = add(w, '\b')
			case 'f':
				w = add(w, '\f')
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
		case '\n', '\r', '\b':
			return w, n, i, ErrBadString
		default:
			if b[i] < utf8.RuneSelf {
				i++
				n++
				break
			}

			r, size := utf8.DecodeRune(b[i:])
			if r == utf8.RuneError && size == 1 {
				w = add(w, b[done:i]...)
				w = utf8.AppendRune(w, utf8.RuneError)

				i += size
				n++
				done = i

				break
			}

			i += size
			n++
		}
	}

	return w, n, i, ErrEndOfBuffer
}

func (p *Parser) skipLit(b []byte, st int) (i int, err error) {
	var lit string

	switch b[st] {
	case 't':
		lit = "true"
	case 'f':
		lit = "false"
	case 'n':
		lit = "null"
	}

	return p.skipVal(b, st, lit)
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

	// Infinity
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
	i, digit := skipInt(b, i, hex)

	dot := false

	if dot = i < len(b) && b[i] == '.'; dot {
		i++
		i, _ = skipInt(b, i, hex)
	}

	exp, exptail := false, false

	if exp = i < len(b) && (b[i] == 'e' || b[i] == 'E'); exp {
		i++
		i = skipSign(b, i)
		i, exptail = skipInt(b, i, false)
	}

	pexp, pexptail := false, false

	if pexp = i < len(b) && (b[i] == 'p' || b[i] == 'P'); pexp {
		i++
		i = skipSign(b, i)
		i, pexptail = skipInt(b, i, false)
	}

	ok := (digit || dot) && (!exp || !pexp) && (hex || !pexp) && exp == exptail && pexp == pexptail

	//	log.Printf("parseNum %10q -> %5v  ditdot %5v %5v  hex %5v  exp %5v %5v  pexp %5v %5v", b[st:i], ok, digit, dot, hex, exp, exptail, pexp, pexptail)

	if !ok {
		return st, ErrBadNumber
	}

	return i, nil
}

// SkipSpaces skips whitespaces.
func SkipSpaces(b []byte, i int) int {
	return (&Parser{}).SkipSpaces(b, i)
}

func skipInt(b []byte, i int, hex bool) (_ int, ok bool) {
	for i < len(b) && (b[i] >= '0' && b[i] <= '9' || b[i] == '_' ||
		hex && (b[i] >= 'a' && b[i] <= 'f' || b[i] >= 'A' && b[i] <= 'F')) {
		ok = true
		i++
	}

	return i, ok
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

func decodeRune(w, b []byte, i int) ([]byte, int, error) {
	size := 0

	switch b[i] {
	case 'x':
		size = 2
	case 'u':
		size = 4
	case 'U':
		size = 8
	default:
		panic(b[i])
	}

	i++

	if i+size > len(b) {
		return w, i - 2, ErrEndOfBuffer
	}

	decode := func(i, size int) (r rune) {
		for j := 0; j < size; j++ {
			c := b[i+j]

			switch {
			case c >= '0' && c <= '9':
				c -= '0'
			case c >= 'a' && c <= 'f':
				c = 10 + c - 'a'
			case c >= 'A' && c <= 'F':
				c = 10 + c - 'A'
			default:
				return -1
			}

			r = r<<4 | rune(c)
		}

		return r
	}

	r := decode(i, size)
	if r < 0 {
		return w, i - 2, ErrBadRune
	}

	// we are after first \u
	if b[i-1] == 'u' && utf16.IsSurrogate(r) &&
		i+size+6 <= len(b) &&
		b[i+4] == '\\' && b[i+5] == 'u' {

		r1 := decode(i+6, size)
		dec := utf16.DecodeRune(r, r1)

		if dec != unicode.ReplacementChar {
			r = dec
			i += 6
		}
	}

	w = utf8.AppendRune(w, r)

	return w, i + size, nil
}

func isWhitespace(b byte) bool {
	return b <= 0x20 && whitespaces&(1<<b) != 0
}
