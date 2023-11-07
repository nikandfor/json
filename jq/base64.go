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

func (f *Base64) Apply(w, r []byte, st int) (_ []byte, i int, err error) {
	w, f.Buf, i, err = base64Apply(w, r, st, f.Encoding, true, f.Buf)

	return w, i, err
}

func (f *Base64d) Apply(w, r []byte, st int) (_ []byte, i int, err error) {
	w, f.Buf, i, err = base64Apply(w, r, st, f.Encoding, false, f.Buf)

	return w, i, err
}

func base64Apply(w, r []byte, st int, e *base64.Encoding, enc bool, buf []byte) (res, buf1 []byte, i int, err error) {
	var p json.Parser

	s, i, err := p.DecodeString(r, st, buf)
	if err != nil {
		return w, s, i, pe(err, i)
	}

	if e == nil {
		e = base64.StdEncoding
	}

	var n int
	wst := len(w)

	if enc {
		n = e.EncodedLen(len(s))
	} else {
		n = e.DecodedLen(len(s))
	}

	n += 3

	if (wst+n)-cap(w) >= 4*1024 {
		q := make([]byte, wst+n)
		copy(q, w)
		w = q
	} else {
		for wst+n > cap(w) {
			w = append(w[:cap(w)], 0, 0, 0, 0, 0, 0, 0, 0)
		}
	}

	w = w[:wst+n]
	wst++

	if enc {
		e.Encode(w[wst:], s)
	} else {
		n, err = e.Decode(w[wst:], s)
		n += 2
		w = w[:wst+n]
		if err != nil {
			return w, s, i, err
		}
	}

	w[wst-1] = '"'
	w[len(w)-2] = '"'
	w[len(w)-1] = '\n'

	return w, s, i, nil
}
