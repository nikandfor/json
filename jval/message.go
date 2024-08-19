package jval

type (
	Message struct {
		b    []byte
		root int
	}
)

func (m *Message) Decode(r []byte, st int) (i int, err error) {
	var d Decoder

	m.b, m.root, i, err = d.Decode(m.b[:0], r, st)

	return
}

func (m *Message) Encode(w []byte) []byte {
	var e Encoder

	return e.Encode(w, m.b, m.root)
}

func (m *Message) BytesRoot() ([]byte, int) { return m.b, m.root }
func (m *Message) Reset(b []byte, root int) { m.b, m.root = b, root }
