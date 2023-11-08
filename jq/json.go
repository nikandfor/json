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

	st = p.SkipSpaces(r, st)
	if st == len(r) {
		return w, st, nil
	}

	f.Buf, i, err = p.DecodeString(r, st, f.Buf[:0])
	if err != nil {
		return w, i, pe(err, i)
	}

	//	log.Printf("JSONDecoder string\n%s", f.Buf)

	var raw []byte

	for j := p.SkipSpaces(f.Buf, 0); j < len(f.Buf); j = p.SkipSpaces(f.Buf, j) {
		raw, j, err = p.Raw(f.Buf, j)
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
