package jval

import (
	"fmt"

	"nikand.dev/go/json"
)

type (
	Encoder struct {
		json.Encoder
	}
)

func (e Encoder) Encode(w, r []byte, off int) []byte {
	tp := r[off]

	if tp&String == String {
		l := int(r[off] &^ String)
		str := off + 1

		return e.AppendString(w, r[str:str+l])
	}

	if tp&Number == Number {
		return fmt.Appendf(w, "%d", tp&0b0011_1111)
	}

	if tp&special == 0 {
		return append(w, []string{
			"false",
			"true",
			"",
			"null",
		}[tp&0x3]...)
	}

	l := int(tp &^ arrObj)
	base := off + 1
	tp &= arrObj

	p := byte('[')
	if tp == Object {
		p = '{'
		l *= 2
	}

	w = append(w, p)

	for j := 0; j < l; {
		if j != 0 {
			w = append(w, ',')
		}

		if tp == Object {
			sub := off - int(r[base+j])

			w = e.Encode(w, r, sub)
			w = append(w, ':')

			j++
		}

		sub := off - int(r[base+j])

		w = e.Encode(w, r, sub)
		j++
	}

	w = append(w, p+2)

	return w
}
