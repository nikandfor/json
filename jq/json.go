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

func (f *JSONDecoder) Next(w, r []byte, st int, state State) (_ []byte, i int, _ State, err error) {
	var p json.Parser

	st = p.SkipSpaces(r, st)
	if st == len(r) {
		return w, st, nil, nil
	}

	f.Buf, i, err = p.DecodeString(r, st, f.Buf[:0])
	if err != nil {
		return w, i, state, pe(err, i)
	}

	//	log.Printf("JSONDecoder string\n%s", f.Buf)

	var raw []byte

	for j := p.SkipSpaces(f.Buf, 0); j < len(f.Buf); j = p.SkipSpaces(f.Buf, j) {
		raw, j, err = p.Raw(f.Buf, j)
		if errors.Is(err, json.ErrEndOfBuffer) {
			break
		}
		if err != nil {
			return w, st, state, pe(pe(err, j), st)
		}

		w = append(w, raw...)
	}

	return w, i, nil, nil
}