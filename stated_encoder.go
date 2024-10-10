package json

import (
	"strconv"
	"unsafe"
)

type (
	StatedEncoder struct {
		Buf []byte

		Encoder

		state encState
		stack encStack
		depth int
	}

	encState int
	encStack int64
)

const (
	encComma = 1 << iota
	encColon
	encNewline
	encKey
	_
)

func NewStatedEncoder(b []byte) *StatedEncoder {
	return &StatedEncoder{Buf: b}
}

func (e *StatedEncoder) Reset() *StatedEncoder {
	e.Buf = e.Buf[:0]
	e.state = 0
	e.stack = 0
	e.depth = 0

	return e
}

func (e *StatedEncoder) Result() []byte { return e.Buf }

func (e *StatedEncoder) KeyString(k, v string) *StatedEncoder {
	return e.Key(k).String(v)
}

func (e *StatedEncoder) KeyStringBytes(k string, v []byte) *StatedEncoder {
	return e.Key(k).StringBytes(v)
}

func (e *StatedEncoder) KeyInt(k string, v int) *StatedEncoder {
	return e.Key(k).Int(v)
}

func (e *StatedEncoder) KeyInt64(k string, v int64) *StatedEncoder {
	return e.Key(k).Int64(v)
}

func (e *StatedEncoder) ObjStart() *StatedEncoder {
	if e.depth > 62 {
		panic("too deep")
	}

	e.comma()

	e.Buf = append(e.Buf, '{')
	e.state.Set(encKey)

	e.stack.SetBit(e.depth)
	e.depth++

	e.state.Set(encNewline)

	return e
}

func (e *StatedEncoder) ObjEnd() *StatedEncoder {
	e.depth--
	e.stack.UnsetBit(e.depth)

	e.Buf = append(e.Buf, '}')
	e.setcomma()

	return e
}

func (e *StatedEncoder) ArrStart() *StatedEncoder {
	e.comma()

	e.Buf = append(e.Buf, '[')
	e.depth++

	e.state.Set(encNewline)

	return e
}

func (e *StatedEncoder) ArrEnd() *StatedEncoder {
	e.depth--

	e.Buf = append(e.Buf, ']')
	e.setcomma()

	return e
}

func (e *StatedEncoder) Key(s string) *StatedEncoder {
	e.comma()

	e.Buf = e.Encoder.AppendKey(e.Buf, s2b(s))
	e.colon(true)

	return e
}

func (e *StatedEncoder) NextIsKey() *StatedEncoder {
	e.state.Set(encKey)

	return e
}

func (e *StatedEncoder) String(s string) *StatedEncoder {
	e.comma()

	e.Buf = e.Encoder.AppendString(e.Buf, s2b(s))
	e.setcomma()

	e.colon(false)

	return e
}

func (e *StatedEncoder) StringBytes(s []byte) *StatedEncoder {
	e.comma()

	e.Buf = e.Encoder.AppendString(e.Buf, s)
	e.setcomma()

	e.colon(false)

	return e
}

func (e *StatedEncoder) Int(v int) *StatedEncoder {
	e.comma()

	e.Buf = strconv.AppendInt(e.Buf, int64(v), 10)
	e.setcomma()

	return e
}

func (e *StatedEncoder) Int64(v int64) *StatedEncoder {
	e.comma()

	e.Buf = strconv.AppendInt(e.Buf, v, 10)
	e.setcomma()

	return e
}

func (e *StatedEncoder) Newline() *StatedEncoder {
	e.newline()

	return e
}

func (e *StatedEncoder) setcomma() {
	e.state.Unset(encComma)

	if e.depth != 0 {
		e.state.Set(encComma)
	}
}

func (e *StatedEncoder) comma() {
	if e.depth == 0 {
		e.newline()
		e.state.Unset(encNewline)

		return
	}

	if e.state.Is(encComma) {
		e.Buf = append(e.Buf, ',')
	}

	e.state.Unset(encComma)
	e.state.Set(encNewline)
}

func (e *StatedEncoder) colon(force bool) {
	if force || e.state.Is(encKey) {
		e.Buf = append(e.Buf, ':')
	}

	e.state.Unset(encKey)
}

func (e *StatedEncoder) newline() {
	if e.state.Is(encNewline) {
		e.Buf = append(e.Buf, '\n')
	}

	e.state.Unset(encNewline)
}

func (s encState) Is(flag encState) bool {
	return s&flag == flag
}

func (s encState) Any(flag encState) bool {
	return s&flag != 0
}

func (s *encState) Set(flag encState) {
	*s |= flag
}

func (s *encState) Unset(flag encState) {
	*s &^= flag
}

func (s *encStack) Bit(n int) bool {
	return *s&(1<<n) != 0
}

func (s *encStack) SetBit(n int) {
	*s |= 1 << n
}

func (s *encStack) UnsetBit(n int) {
	*s &^= 1 << n
}

func s2b(s string) []byte {
	data := unsafe.StringData(s)

	return unsafe.Slice(data, len(s))
}
