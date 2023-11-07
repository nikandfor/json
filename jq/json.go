package jq

import (
	"errors"

	"github.com/nikandfor/json"
)

type (
	//	JSONEncode struct{}

	JSONDecoder struct {
		Buf []byte
	}
)

func (f *JSONDecoder) Apply(w, r []byte, st int) (_ []byte, i int, err error) {
	var p json.Parser

	s, i, err := p.DecodeString(r, st, f.Buf[:0])
	f.Buf = s
	if err != nil {
		return w, i, pe(err, i)
	}

	var raw []byte

	for j := 0; j < len(s); {
		raw, j, err = p.Raw(s, j)
		if errors.Is(err, json.ErrEndOfBuffer) {
			break
		}
		if err != nil {
			return w, st, pe(pe(err, j), st)
		}

		w = append(w, raw...)
		w = append(w, '\n')
	}

	return w, i, nil
}
