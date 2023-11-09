package jq

import (
	"encoding/base64"

	"github.com/nikandfor/json"
)

type (
	Base64 struct {
		*base64.Encoding
		Buf []byte
	}

	Base64d struct {
		*base64.Encoding
		Buf []byte
	}
)

func (f *Base64) Next(w, r []byte, st int, state State) (_ []byte, i int, _ State, err error) {
	w, f.Buf, i, err = base64Apply(w, r, st, f.Encoding, true, f.Buf[:0])

	return w, i, nil, err
}

func (f *Base64d) Next(w, r []byte, st int, state State) (_ []byte, i int, _ State, err error) {
	w, f.Buf, i, err = base64Apply(w, r, st, f.Encoding, false, f.Buf[:0])

	return w, i, nil, err
}

func base64Apply(w, r []byte, st int, e *base64.Encoding, enc bool, buf []byte) (res, buf1 []byte, i int, err error) {
	var p json.Parser

	st = p.SkipSpaces(r, st)
	if st == len(r) {
		return w, buf, st, nil
	}

	s, i, err := p.DecodeString(r, st, buf)
	if err != nil {
		return w, s, i, pe(err, i)
	}

	if e == nil {
		e = base64.StdEncoding
	}

	if enc {
		n := e.EncodedLen(len(s))

		wst := len(w) + 1 // open "
		w = grow(w, n+2)

		e.Encode(w[wst:], s)

		w[wst-1] = '"'
		w[len(w)-1] = '"'
		//	w[len(w)-1] = '\n'
	} else {
		n := e.DecodedLen(len(s))

		ssize := len(s)
		s = grow(s, ssize+n)

		n, err = e.Decode(s[ssize:], s[:ssize])
		//	log.Printf("decoded base64 (err %v): %q -> %q", err, s[:ssize], s[ssize:ssize+n])
		if err != nil {
			return w, s, i, err
		}

		w = (&json.Generator{}).EncodeString(w, s[ssize:ssize+n])
		//	w = append(w, '\n')
	}

	return w, s, i, nil
}

func grow(b []byte, size int) []byte {
	if size-cap(b) >= 4*1024 {
		q := make([]byte, size)
		copy(q, b)
		return q
	}

	for size > cap(b) {
		b = append(b[:cap(b)], 0, 0, 0, 0, 0, 0, 0, 0)
	}

	return b[:size]
}
