package json

import (
	"errors"
	"io"
	"unicode/utf8"
)

type (
	Reader struct {
		r io.Reader

		lock []int

		b []byte
		i int

		off int64
	}
)

// NewReader creates a new Reader.
// It first reads from b and then from rd.
// If you just want to provide a buffer slice it to zero length.
func NewReader(b []byte, rd io.Reader) *Reader {
	return &Reader{
		r: rd,
		b: b,
	}
}

// Reset resets reader.
// As in NewReader it makes reading first from b and then from rd.
func (r *Reader) Reset(b []byte, rd io.Reader) {
	r.r = rd
	r.b = b
	r.lock = r.lock[:0]
	r.i = 0
	r.off = 0
}

// Offset returns current position in the stream.
func (r *Reader) Offset() int64 {
	return r.off + int64(r.i)
}

// Type finds the beginning of the next value and detects its type.
// It doesn't parse the value so it can't detect if it's incorrect.
func (r *Reader) Type() (tp byte, err error) {
again:
	for r.i < len(r.b) {
		if isWhitespace(r.b[r.i]) {
			r.i++
			continue
		}

		switch r.b[r.i] {
		case ',', ':':
			r.i++
			continue
		case 't', 'f':
			return Bool, nil
		case '"':
			return String, nil
		case Null, Array, Object:
			return r.b[r.i], nil
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
			'+', '-', '.',
			'N',      // NaN
			'i', 'I': // Inf
			return Number, nil
		}

		return None, ErrSyntax
	}

	err = r.more()
	if err != nil {
		return None, err
	}

	goto again
}

// Skip skips the next value.
func (r *Reader) Skip() error {
	return r.Break(0)
}

// Raw skips the next value and returns subslice with the value trimming whitespaces.
//
// Returned buffer is only valid until the next reading method is called.
// It can be reused if more data needed to be read from underlying reader.
func (r *Reader) Raw() ([]byte, error) {
	_, err := r.Type()
	if err != nil {
		return nil, err
	}

	l := r.Lock()
	defer r.Unlock()

	err = r.Break(0)
	if err != nil {
		return nil, nil
	}

	st := r.lock[l-1]

	return r.b[st:r.i], nil
}

// Break breaks from inside the object to the end of it on depth levels.
// As a special case with depth=0 it skips the next value.
// Skip and Raw do exactly that.
//
// It's intended for exiting out of arrays and objects when their content is not needed anymore
// (all the needed indexes or keys are already parsed) and we want to parse the next array or object.
func (r *Reader) Break(depth int) (err error) {
	var d Decoder

again:
	for err == nil && r.i < len(r.b) {
		if isWhitespace(r.b[r.i]) {
			r.i++
			continue
		}

		switch r.b[r.i] {
		case ',', ':':
			r.i++
			continue
		case '"':
			err = r.skipString()
		case 'n', 't', 'f':
			r.i, err = d.skipLit(r.b, r.i)
			if err == ErrEndOfBuffer { //nolint:errorlint
				err = nil
				break again
			}
		case '[', '{':
			r.i++
			depth++
		case ']', '}':
			r.i++
			depth--
		default:
			r.i, err = d.skipNum(r.b, r.i)
			if err == ErrBadNumber && r.i == len(r.b) { //nolint:errorlint
				err = nil
				break again
			}
		}

		if depth == 0 {
			return nil
		}
	}
	if err != nil {
		return err
	}

	err = r.more()

	goto again
}

// Key reads the next string removing quotes but not decoding the string value.
// So escape sequences (\n, \uXXXX) are not decoded. They are returned as is.
// This is intended for object keys as they usually contain alpha-numeric symbols only.
// This is faster and does not require additional buffer for decoding.
//
// Returned buffer is only valid until the next reading method is called.
// It can be reused if more data needed to be read from underlying reader.
func (r *Reader) Key() ([]byte, error) {
	tp, err := r.Type()
	if err != nil {
		return nil, err
	}
	if tp != String {
		return nil, ErrType
	}

	l := r.Lock()
	defer r.Unlock()

	if err := r.skipString(); err != nil {
		return nil, err
	}

	st := r.lock[l-1]

	return r.b[st+1 : r.i-1], nil
}

// DecodeString reads the next string, decodes escape sequences (\n, \uXXXX),
// and appends the result to the buf.
//
// Data is appended to the provided buffer. And the buffer will not be preserved by Reader.
func (r *Reader) DecodeString(buf []byte) (s []byte, err error) {
	tp, err := r.Type()
	if err != nil {
		return buf, err
	}
	if tp != String {
		return buf, ErrType
	}

	if buf == nil {
		buf = []byte{}
	}

	s, _, err = r.decodeString(buf)

	return
}

// DecodedStringLength reads and decodes the next string but only return the result length.
// It doesn't allocate while DecodeString does.
func (r *Reader) DecodedStringLength() (n int, err error) {
	tp, err := r.Type()
	if err != nil {
		return 0, err
	}
	if tp != String {
		return 0, ErrType
	}

	_, n, err = r.decodeString(nil)

	return
}

// Enter enters an Array or an Object. typ is checked to match with the actual container type.
// Use More or, more convenient form, ForMore to iterate over container.
// See examples to better understand usage pattern.
func (r *Reader) Enter(typ byte) (err error) {
	tp, err := r.Type()
	if err != nil {
		return
	}

	if tp != typ || typ != Array && typ != Object {
		return ErrType
	}

	r.i++

	return
}

// More iterates over an Array or an Object elements entered by the Enter method.
func (r *Reader) More(typ byte) (more bool, err error) {
again:
	for r.i < len(r.b) {
		if isWhitespace(r.b[r.i]) || r.b[r.i] == ',' {
			r.i++
			continue
		}

		break
	}

	if r.i == len(r.b) {
		if err := r.more(); err != nil {
			return false, err
		}

		goto again
	}

	if r.b[r.i] == typ+2 {
		r.i++
		return false, nil
	}

	tp, err := r.Type()
	if err != nil {
		return false, err
	}

	if typ == Object && tp != String {
		return false, ErrSyntax
	}

	return true, nil
}

// ForMore is a convenient wrapper for More which makes iterating code shorter and simpler.
func (r *Reader) ForMore(typ byte, errp *error) bool { //nolint:gocritic
	more, err := r.More(typ)
	if err != nil {
		*errp = err
	}

	return more
}

// Length calculates number of elements in Array or Object.
func (r *Reader) Length() (n int, err error) {
	tp, err := r.Type()
	if err != nil {
		return 0, err
	}

	switch tp {
	case Array, Object:
	default:
		return 0, ErrType
	}

	err = r.Enter(tp)
	if err != nil {
		return 0, err
	}

	for r.ForMore(tp, &err) {
		if tp == Object {
			_, err = r.Key()
			if err != nil {
				return n, err
			}
		}

		err = r.Skip()
		if err != nil {
			return n, err
		}

		n++
	}
	if err != nil {
		return n, err
	}

	return n, nil
}

// Lock locks the buffer from rewriting by reading more data from Reader.
// Lock also remembers position in the stream and allows rewinding to it.

// Lock locks internal buffer so the data is not overwritten when more data is read from underlaying reader.
// It's used to return to the locked position in a stream and reread some part of it.
// Internal buffer grows to the size of data locked plus additional space for the next Read.
// Lock must be followed by Unlock just like for sync.Mutex.
// Rewind is used to return to the latest Lock position.
// Multiple nested locks are allowed.
// It returns the number of locks acquired and not released so far; kinda Lock depth.
func (r *Reader) Lock() int {
	r.lock = append(r.lock, r.i)

	return len(r.lock)
}

// Unlock releases the latest buffer Lock.
// It returns the number of remaining active Locks.
func (r *Reader) Unlock() int {
	r.lock = r.lock[:len(r.lock)-1]

	return len(r.lock)
}

// Rewind returns stream position to the latest Lock.
func (r *Reader) Rewind() {
	r.i = r.lock[len(r.lock)-1]
}

func (r *Reader) skipString() (err error) {
	b := r.b
	i := r.i + 1 // opening "

again:
	for err == nil && i < len(b) {
		switch b[i] {
		case '"':
			r.i = i + 1
			return nil
		case '\\':
			if i+1 >= len(b) {
				break again
			}

			size := 0

			switch b[i+1] {
			case '"', '\\', '/', 'n', 'r', 't', 'b', 'f':
			case 'x':
				size = 2
			case 'u':
				size = 4
			case 'U':
				size = 8
			default:
				r.i = i
				return ErrBadString
			}

			if i+1+1+size > len(b) {
				break again
			}

			i += 1 + 1 + size
		case '\n', '\r', '\b':
			r.i = i
			return ErrBadString
		default:
			if b[i] < utf8.RuneSelf {
				i++
				break
			}

			if !utf8.FullRune(b[i:]) {
				break again
			}

			_, size := utf8.DecodeRune(b[i:])
			//	if r == utf8.RuneError && size == 1 {
			//		return i, ErrBadRune
			//	}

			i += size
		}
	}
	r.b = b
	r.i = i
	if err != nil {
		return
	}

	err = r.more()

	goto again
}

func (r *Reader) decodeString(w []byte) (_ []byte, n int, err error) { //nolint:gocognit
	r.i += 1 // opening "
	done := r.i

	add := func(d []byte, s ...byte) []byte {
		n += len(s)

		if w == nil {
			return nil
		}

		return append(w, s...)
	}

again:
	for err == nil && r.i < len(r.b) {
		switch r.b[r.i] {
		case '"':
			w = add(w, r.b[done:r.i]...)
			r.i++
			return w, n, nil
		case '\\':
			w = add(w, r.b[done:r.i]...)

			if r.i+1 == len(r.b) {
				break again
			}

			r.i++

			switch r.b[r.i] {
			case '\\', '"', '/':
				w = add(w, r.b[r.i])
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
				w, r.i, err = decodeRune(w, r.b, r.i)
				if err == ErrEndOfBuffer { //nolint:errorlint
					break again
				}
				if err != nil {
					return w, n, err
				}

				r.i--
			default:
				return w, n, ErrBadString
			}

			r.i++
			done = r.i
		case '\n', '\r', '\b':
			return w, n, ErrBadString
		default:
			if r.b[r.i] < utf8.RuneSelf {
				r.i++
				break
			}

			if !utf8.FullRune(r.b[r.i:]) {
				break again
			}

			rr, size := utf8.DecodeRune(r.b[r.i:])
			if rr == utf8.RuneError && size == 1 {
				w = add(w, r.b[done:r.i]...)
				w = utf8.AppendRune(w, utf8.RuneError)

				r.i += size
				done = r.i

				break
			}

			r.i += size
		}
	}
	if err != nil {
		return w, n, err
	}

	err = r.more()

	goto again
}

func (r *Reader) more() error {
	if r.r == nil && r.i == len(r.b) {
		return io.EOF
	}

	if r.r == nil {
		return io.ErrUnexpectedEOF
	}

	end := len(r.b)

	st := r.i
	if len(r.lock) > 0 {
		st = r.lock[0]
	}

	if st != 0 {
		r.off += int64(st)

		copy(r.b, r.b[st:end])
		end -= st
		r.i -= st

		for i := range r.lock {
			r.lock[i] -= st
		}
	}

	if cap(r.b) == 0 {
		r.b = make([]byte, 16<<10)
	}

	r.b = r.b[:cap(r.b)]

	n, err := r.r.Read(r.b[end:])
	end += n

	r.b = r.b[:end]

	if n != 0 && errors.Is(err, io.EOF) {
		err = nil
	}

	return err
}
