package jval

import "nikand.dev/go/json"

type (
	Decoder struct {
		json.Decoder
	}
)

const (
	// Num
	//   Embedded
	//   Int: off
	//   Float: off
	//
	// Null, True, False
	//
	// String: off, len
	//
	// Array, Object: len, off -> []off

	String = 0b1000_0000 // 0b1xxx_xxxx
	Number = 0b0100_0000 // 0b01xx_xxxx

	Array  = 0b0010_0000 // 0b0010_xxxx
	Object = 0b0011_0000 // 0b0011_xxxx

	arrObj = 0b0011_0000

	_ = 0b0001_0000 // 0b0001_xxxx

	False = 0b0000_1100
	True  = 0b0000_1101
	_     = 0b0000_1110
	Null  = 0b0000_1111

	special = 0b1111_0000

	strLen1 = String - 3
	strLen2 = String - 2
	strLen4 = String - 1

	intLen1 = Number - 4
	intLen2 = Number - 3
	intLen4 = Number - 2
	float   = Number - 1
)

func (d Decoder) Decode(w, r []byte, st int) (_ []byte, off, i int, err error) {
	var raw []byte

	off = len(w)

	tp, i, err := d.Type(r, st)
	if err != nil {
		return w, -1, i, err
	}

	switch tp {
	case json.Null, json.Bool:
		raw, i, err = d.Raw(r, i)
		if err != nil {
			return w, -1, i, err
		}

		var x byte

		switch {
		case tp == json.Null:
			x = Null
		case string(raw) == "true":
			x = True
		default:
			x = False
		}

		w = append(w, x)

		return w, off, i, nil
	case json.Number:
		raw, i, err = d.Raw(r, i)
		if err != nil {
			return
		}

		v := 0

		for j := 0; j < len(raw); j++ {
			if raw[j] >= '0' && raw[j] <= '9' {
				v = v*10 + int(raw[j]-'0')
			} else {
				panic("number")
			}
		}

		if v < intLen1 {
			w = append(w, Number|byte(v))

			return w, off, i, nil
		}

		panic("number")
	case json.String:
		w = append(w, String)

		str := len(w)

		w, i, err = d.DecodeString(r, i, w)
		if err != nil {
			return w, -1, i, err
		}

		l := len(w) - str

		if l < strLen1 {
			w[str-1] = String | byte(l)

			return w, off, i, nil
		}

		w[str-1] = String | strLen1
		// TODO
		panic("str")
	}

	sub := make([]int, 0, 8)

	i, err = d.Enter(r, i, tp)
	if err != nil {
		return w, -1, i, err
	}

	for d.ForMore(r, &i, tp, &err) {
		if tp == json.Object {
			w, off, i, err = d.Decode(w, r, i)
			if err != nil {
				return w, -1, i, err
			}

			sub = append(sub, off)
		}

		w, off, i, err = d.Decode(w, r, i)
		if err != nil {
			return w, -1, i, err
		}

		sub = append(sub, off)
	}
	if err != nil {
		return w, -1, i, err
	}

	if tp == json.Array {
		tp = Array
	} else {
		tp = Object
	}

	off = len(w)

	if len(sub) == 0 {
		w = append(w, tp)

		return w, off, i, nil
	}

	l := len(sub)
	if tp == Object {
		l /= 2
	}

	if l < 8 && off-sub[0] < 256 {
		w = append(w, tp|byte(l))

		for _, sub := range sub {
			w = append(w, byte(off-sub))
		}

		return w, off, i, nil
	}

	panic("obj/arr")
}
