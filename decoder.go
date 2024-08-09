package json

import (
	"errors"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

// Value types returned by Decoder.
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
	// Decoder is a group of methods to parse JSON buffers.
	// Decoder is stateless.
	// All the needed state is passed though arguments and return values.
	//
	// Most of the methods take buffer with json and start position
	// and return a value, end position and possible error.
	Decoder struct{}
)

var ( // bitsets
	whitespaces uint64
	decimals    uint64
	hexdecimals uint64 // -64 offset
)

func init() {
	for _, b := range []byte{'\n', '\r', '\t', ' '} {
		whitespaces |= 1 << b
	}

	for b := '0'; b <= '9'; b++ {
		decimals |= 1 << b
	}

	for b := 'a'; b <= 'z'; b++ {
		hexdecimals |= 1 << b
		hexdecimals |= 1 << (b - 'a' + 'A')
	}

	_, _ = isDigit1('f', true), isDigit2('f', true) // keep it used
}

// Errors returned by Decoder.
var (
	ErrBadNumber   = errors.New("bad number")
	ErrBadRune     = errors.New("bad rune")
	ErrBadString   = errors.New("bad string")
	ErrEndOfBuffer = errors.New("unexpected end of buffer")
	ErrNoSuchKey   = errors.New("no such object key")
	ErrOutOfBounds = errors.New("out of array bounds")
	ErrSyntax      = errors.New("syntax error")
	ErrType        = errors.New("incompatible type")
)

// Type finds the beginning of the next value and detects its type.
// It doesn't parse the value so it can't detect if it's incorrect.
func (d *Decoder) Type(b []byte, st int) (tp byte, i int, err error) {
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
func (d *Decoder) Skip(b []byte, st int) (i int, err error) {
	return d.Break(b, st, 0)
}

// Raw skips the next value and returns subslice with the value trimming whitespaces.
func (d *Decoder) Raw(b []byte, st int) (v []byte, i int, err error) {
	_, st, err = d.Type(b, st)
	if err != nil {
		return nil, st, err
	}

	i, err = d.Break(b, st, 0)
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
func (d *Decoder) Break(b []byte, st, depth int) (i int, err error) {
	i = st

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
			i, err = d.skipString(b, i)
		case 'n', 't', 'f':
			i, err = d.skipLit(b, i)
		case '[', '{':
			i++
			depth++
		case ']', '}':
			i++
			depth--
		default:
			i, err = d.skipNum(b, i)
		}

		if err != nil {
			return
		}

		if depth == 0 {
			return
		}
	}

	return st, ErrEndOfBuffer
}

// Key reads the next string removing quotes but not decoding the string value.
// So escape sequences (\n, \uXXXX) are not decoded. They are returned as is.
// This is intended for object keys as they usually contain alpha-numeric symbols only.
// This is faster and does not require additional buffer for decoding.
func (d *Decoder) Key(b []byte, st int) (k []byte, i int, err error) {
	tp, i, err := d.Type(b, st)
	if err != nil {
		return
	}

	if tp != String {
		return nil, i, ErrType
	}

	raw, i, err := d.Raw(b, i)
	if err != nil {
		return
	}

	return raw[1 : len(raw)-1], i, nil
}

// DecodeString reads the next string, decodes escape sequences (\n, \uXXXX),
// and appends the result to the buf.
func (d *Decoder) DecodeString(b []byte, st int, buf []byte) (s []byte, i int, err error) {
	tp, i, err := d.Type(b, st)
	if err != nil {
		return buf, i, err
	}

	if tp != String {
		return buf, i, ErrType
	}

	if buf == nil {
		buf = []byte{}
	}

	s, _, i, err = d.decodeString(b, i, buf)

	return
}

// DecodedStringLength reads and decodes the next string but only return the result length.
// It doesn't allocate while DecodeString does.
func (d *Decoder) DecodedStringLength(b []byte, st int) (n, i int, err error) {
	tp, i, err := d.Type(b, st)
	if err != nil {
		return
	}

	if tp != String {
		return 0, i, ErrType
	}

	_, n, i, err = d.decodeString(b, i, nil)

	return
}

// Enter enters an Array or an Object. typ is checked to match with the actual container type.
// Use More or, more convenient form, ForMore to iterate over container.
// See examples to understand the usage pattern more.
func (d *Decoder) Enter(b []byte, st int, typ byte) (i int, err error) {
	tp, i, err := d.Type(b, st)
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
func (d *Decoder) More(b []byte, st int, typ byte) (more bool, i int, err error) {
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

	tp, i, err := d.Type(b, i)
	if err != nil {
		return false, i, err
	}

	if typ == Object && tp != String {
		return false, i, ErrSyntax
	}

	return true, i, nil
}

// ForMore is a convenient wrapper for More which makes iterating code shorter and simpler.
func (d *Decoder) ForMore(b []byte, i *int, typ byte, errp *error) bool { //nolint:gocritic
	more, j, err := d.More(b, *i, typ)
	*i = j

	if errp != nil {
		*errp = err
	}

	return more
}

// Length calculates number of elements in Array or Object.
func (d *Decoder) Length(b []byte, st int) (n, i int, err error) {
	tp, i, err := d.Type(b, st)
	if err != nil {
		return 0, i, err
	}

	switch tp {
	case Array, Object:
	default:
		return 0, i, ErrType
	}

	i, err = d.Enter(b, i, tp)
	if err != nil {
		return 0, i, err
	}

	for d.ForMore(b, &i, tp, &err) {
		if tp == Object {
			_, i, err = d.Key(b, i)
			if err != nil {
				return n, i, err
			}
		}

		i, err = d.Skip(b, i)
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
func (d *Decoder) SkipSpaces(b []byte, i int) int {
	for i < len(b) && isWhitespace(b[i]) {
		i++
	}

	return i
}

func (d *Decoder) skipString(b []byte, st int) (i int, err error) {
	i = st + 1 // opening "

	for i < len(b) {
		switch b[i] {
		case '"':
			return i + 1, nil
		case '\\':
			i++
			if i == len(b) {
				return st, ErrEndOfBuffer
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
				return st, ErrEndOfBuffer
			}

			i += size
		case '\n', '\r', '\b':
			return i, ErrBadString
		default:
			if b[i] < utf8.RuneSelf {
				i++
				break
			}

			if !utf8.FullRune(b[i:]) {
				return st, ErrEndOfBuffer
			}

			_, size := utf8.DecodeRune(b[i:])
			//	if r == utf8.RuneError && size == 1 {
			//		return i, ErrBadRune
			//	}

			i += size
		}
	}

	return st, ErrEndOfBuffer
}

func (d *Decoder) decodeString(b []byte, st int, w []byte) (_ []byte, n, i int, err error) {
	i = st + 1 // opening "
	done := i

	add := func(d []byte, s ...byte) []byte {
		n += len(s)

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
			return w, n, i, nil
		case '\\':
			w = add(w, b[done:i]...)

			i++
			if i == len(b) {
				return w, n, st, ErrEndOfBuffer
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
				if err == ErrEndOfBuffer {
					i = st
				}
				if err != nil {
					return w, n, i, err
				}

				i--
			default:
				return w, n, i, ErrBadString
			}

			i++
			done = i
		case '\n', '\r', '\b':
			return w, n, i, ErrBadString
		default:
			if b[i] < utf8.RuneSelf {
				i++
				break
			}

			if !utf8.FullRune(b[i:]) {
				return w, n, st, ErrEndOfBuffer
			}

			r, size := utf8.DecodeRune(b[i:])
			if r == utf8.RuneError && size == 1 {
				w = add(w, b[done:i]...)
				w = utf8.AppendRune(w, utf8.RuneError)

				i += size
				done = i

				break
			}

			i += size
		}
	}

	return w, n, st, ErrEndOfBuffer
}

func (d *Decoder) skipLit(b []byte, st int) (i int, err error) {
	var lit string

	switch b[st] {
	case 't':
		lit = "true"
	case 'f':
		lit = "false"
	case 'n':
		lit = "null"
	}

	return d.skipVal(b, st, lit)
}

func (d *Decoder) skipVal(b []byte, st int, val string) (i int, err error) {
	end := st + len(val)

	if end <= len(b) && string(b[st:end]) == val {
		return end, nil
	}

	if end > len(b) && string(b[st:]) == val[:len(b)-st] {
		return st, ErrEndOfBuffer
	}

	return st, ErrSyntax
}

func (d *Decoder) skipNum(b []byte, st int) (i int, err error) {
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
	var hex bool

	if i+2 < len(b) && b[i] == '0' && (b[i+1] == 'x' || b[i+1] == 'X') {
		hex = true
		i += 2
	}

	// integer
	i, digit := skipInt(b, i, hex)

	var dot bool

	if dot = i < len(b) && b[i] == '.'; dot {
		i++
		i, _ = skipInt(b, i, hex)
	}

	var exp, exptail bool

	if exp = i < len(b) && (b[i] == 'e' || b[i] == 'E'); exp {
		i++
		i = skipSign(b, i)
		i, exptail = skipInt(b, i, false)
	}

	var pexp, pexptail bool

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
	return (&Decoder{}).SkipSpaces(b, i)
}

func skipSign(b []byte, i int) int {
	if i < len(b) && (b[i] == '+' || b[i] == '-') {
		i++
	}

	return i
}

func skipInt(b []byte, i int, hex bool) (_ int, ok bool) {
	for i < len(b) && (b[i] == '_' || isDigit1(b[i], hex)) {
		ok = true
		i++
	}

	return i, ok
}

func decodeRune(w, b []byte, st int) (_ []byte, i int, err error) {
	i = st

	var size int

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
		return w, st - 1, ErrEndOfBuffer
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
		return w, st - 1, ErrBadRune
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

func isDigit1(b byte, hex bool) bool {
	return b >= '0' && b <= '9' || hex && (b >= 'a' && b <= 'f' || b >= 'A' && b <= 'F')
}

func isDigit2(b byte, hex bool) bool {
	return b < 64 && decimals&(1<<b) != 0 || b >= 64 && b < 128 && hexdecimals&(1<<b) != 0
}
