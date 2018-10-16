package json

import "io"

type Writer struct {
	b      []byte
	ref, i int

	d         int
	pref, ind []byte

	ncomma bool
	naopen bool
	prefln bool

	err error

	w io.Writer
}

func NewWriter(b []byte) *Writer {
	return NewIndentWriter(b, nil, nil)
}

func NewIndentWriter(b []byte, p, i []byte) *Writer {
	w := &Writer{
		b:      b,
		pref:   p,
		ind:    i,
		prefln: true,
	}
	return w
}

func NewStreamWriter(w io.Writer) *Writer {
	return NewStreamWriterBuffer(w, nil)
}

func NewStreamWriterBuffer(w io.Writer, b []byte) *Writer {
	if len(b) == 0 {
		b = make([]byte, 1000)
	}
	return &Writer{
		b:      b,
		prefln: true,
		w:      w,
	}
}

func (w *Writer) ArrayStart() {
	w.valueStart()
	w.add([]byte{'['})
	w.d++
	w.naopen = true
}

func (w *Writer) ArrayEnd() {
	if w.naopen {
		w.naopen = false
	} else {
		w.newline(0)
	}
	w.d--
	if w.prefln {
		w.prefln = false
		w.addpref()
	}
	w.add([]byte{']'})
	w.valueEnd()
}

func (w *Writer) ObjStart() {
	w.valueStart()
	w.add([]byte{'{'})
	w.d++
	w.naopen = true
}

func (w *Writer) ObjEnd() {
	if w.naopen {
		w.naopen = false
	} else {
		w.newline(-1)
	}
	w.d--
	if w.prefln {
		w.prefln = false
		w.addpref()
	}
	w.add([]byte{'}'})
	w.valueEnd()
}

func (w *Writer) ObjKey(k []byte) {
	if w.naopen {
		w.naopen = false
		w.newline(0)
	}
	w.naopen = false
	w.String(k)
	w.add([]byte{':'})
	if len(w.ind) != 0 {
		w.add([]byte{' '})
	}
	w.ncomma = false
}

func (w *Writer) String(v []byte) {
	w.valueStart()
	w.add([]byte{'"'})
	w.add(v)
	w.add([]byte{'"'})
	w.valueEnd()
}

func (w *Writer) Number(v []byte) {
	w.RawBytes(v)
}

func (w *Writer) True() {
	w.RawBytes([]byte("true"))
}

func (w *Writer) False() {
	w.RawBytes([]byte("false"))
}

func (w *Writer) Null() {
	w.RawBytes([]byte("null"))
}

func (w *Writer) RawBytes(v []byte) {
	w.valueStart()
	w.add(v)
	w.valueEnd()
}

func (w *Writer) valueStart() {
	if w.naopen {
		w.naopen = false
		w.newline(0)
	}
	w.comma()
	if w.prefln {
		w.prefln = false
		w.addpref()
	}
}

func (w *Writer) valueEnd() {
	w.ncomma = true
}

func (w *Writer) SetIndent(pref, ind []byte) {
	w.pref, w.ind = pref, ind
}

func (w *Writer) ObjKeyString(k string) {
	w.ObjKey([]byte(k))
}

func (w *Writer) StringString(v string) {
	w.String([]byte(v))
}

func (w *Writer) comma() {
	if !w.ncomma {
		return
	}
	w.add([]byte{','})
	w.newline(0)
	w.prefln = true
}

func (w *Writer) newline(d int) {
	if len(w.ind) != 0 {
		w.d += d
		w.add([]byte{'\n'})
		w.prefln = true
	}
}

func (w *Writer) addpref() {
	w.add(w.pref)
	for i := 0; i < w.d; i++ {
		w.add(w.ind)
	}
}

func (w *Writer) add(t []byte) {
	for {
		n := copy(w.b[w.i:], t)
		w.i += n
		if n == len(t) {
			return
		}

		if !w.more() {
			return
		}
		t = t[n:]
	}
}

func (w *Writer) more() bool {
	if w.w == nil {
		w.b = append(w.b, 0)
		w.b = w.b[:cap(w.b)]
		return true
	}
	r, err := w.w.Write(w.b[:w.i])
	w.ref += r
	if r < w.i {
		copy(w.b, w.b[r:])
		w.i -= r
	} else {
		w.i = 0
	}
	w.err = err
	return r > 0
}

func (w *Writer) Flush() error {
	if w.i == 0 {
		return nil
	}
	w.more()
	return w.Err()
}

func (w *Writer) Bytes() []byte {
	return w.b[:w.i]
}

func (w *Writer) Err() error {
	return w.err
}