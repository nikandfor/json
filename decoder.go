package json

import (
	"errors"
	"io"

	"nikand.dev/go/skip"
)

type (
	// Decoder is a group of methods to parse JSON.
	// Decoder is stateless.
	// All the needed state is passed though arguments and return values.
	//
	// Most of the methods take buffer with json and start position
	// and return a value, end position and possible error.
	Decoder struct{}
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
}

// Decoder errors. Plus Str errors from skip module.
var (
	ErrBadNumber   = errors.New("bad number")
	ErrShortBuffer = io.ErrShortBuffer
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

	return None, i, ErrShortBuffer
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

	return st, ErrShortBuffer
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

	ss, w, _, i := skip.DecodeString(b, i, skip.Quo|skip.ErrRune, buf)
	if ss.Is(skip.ErrBuffer) {
		return w, st, ErrShortBuffer
	}
	if ss.Err() {
		return w, i, ss
	}

	return w, i, nil
}

// DecodedStringLength reads and decodes the next string but only return the result length.
// It doesn't allocate while DecodeString does.
func (d *Decoder) DecodedStringLength(b []byte, st int) (bs, rs, i int, err error) {
	tp, i, err := d.Type(b, st)
	if err != nil {
		return
	}

	if tp != String {
		return 0, 0, i, ErrType
	}

	ss, bs, rs, i := skip.String(b, i, skip.Quo|skip.ErrRune)
	if ss.Is(skip.ErrBuffer) {
		return bs, rs, st, ErrShortBuffer
	}
	if ss.Err() {
		return bs, rs, i, ss
	}

	return bs, rs, i, nil
}

// Enter enters an Array or an Object. typ is checked to match with the actual container type.
// Use More or, more convenient form, ForMore to iterate over container.
// See examples to better understand usage pattern.
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
		return false, i, ErrShortBuffer
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
	ss, _, _, i := skip.String(b, st, skip.Quo)
	if ss.Is(skip.ErrBuffer) {
		return st, ErrShortBuffer
	}
	if ss.Err() {
		return i, ss
	}

	return i, nil
}

func (d *Decoder) skipNum(b []byte, st int) (i int, err error) {
	n, i := skip.Number(b, st, 0)
	if !n.Ok() {
		return i, ErrBadNumber
	}

	return i, nil
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
		return st, ErrShortBuffer
	}

	return st, ErrSyntax
}

// SkipSpaces skips whitespaces.
func SkipSpaces(b []byte, i int) int {
	return (&Decoder{}).SkipSpaces(b, i)
}

func isWhitespace(b byte) bool {
	return b <= 0x20 && whitespaces&(1<<b) != 0
}
