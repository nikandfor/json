package jq

import (
	"errors"

	"nikand.dev/go/json"
)

type (
	// Index returns given Object key value or Array element at the index.
	// Supported index types: string (object key), int (array index), Iter (iterate over all values of Object or Array).
	Index struct {
		Filters []any
		pool    []*indexState
	}

	indexState struct {
		sub []indexSub
	}

	indexSub State
)

var ErrUnsupportedIndexFilter = errors.New("unsupported index filter")

func NewIndex(filters ...any) *Index {
	return &Index{Filters: filters}
}

func (f *Index) Next(w, r []byte, st int, state State) (_ []byte, i int, _ State, err error) {
	var d json.Decoder

	st = d.SkipSpaces(r, st)
	if st == len(r) {
		return w, st, nil, nil
	}

	i = st
	wreset := len(w)

	if len(f.Filters) == 0 {
		w, i, _, err = Dot{}.Next(w, r, i, nil)
		if err != nil {
			return w[:wreset], i, nil, pe(err, i)
		}
	}

	var stateok *indexState
	var stack []indexSub

	if stateok, _ = state.(*indexState); stateok != nil {
		stack = stateok.sub
	} else if stateok = f.state(); stateok != nil {
		stack = stateok.sub
	}

	//	log.Printf("index state  %d  %+v", i, stateok)

	if stateok == nil {
		i, err := d.Seek(r, i, f.Filters...)
		if err == json.ErrNoSuchKey {
			w = append(w, `null`...)

			return w, i, nil, nil
		}
		if err != nil {
			return w, i, nil, pe(err, i)
		}

		w, i, _, err = Dot{}.Next(w, r, i, nil)
		if err != nil {
			return w, i, nil, pe(err, i)
		}

		i, err = d.Break(r, i, len(f.Filters))
		if err != nil {
			return w, i, nil, pe(err, i)
		}

		return w, i, nil, nil
	}

	w, i, ok, err := f.next(w, r, i, stack, state == nil)
	if err != nil {
		return w[:wreset], i, nil, err
	}

	if !ok {
		f.pool = append(f.pool, stateok)

		return w, i, nil, nil
	}

	return w, i, stateok, nil
}

func (f *Index) next(w, r []byte, st int, stack []indexSub, first bool) (_ []byte, i int, ok bool, err error) {
	var d json.Decoder

	i = st
	fi := len(stack) - 1

	if first {
		fi = -1
	}

back:
	for {
		//	log.Printf("index back %d  %d", fi, i)

		for ; fi >= 0 && stack[fi] == nil; fi-- {
			ff := f.Filters[fi]

			switch ff.(type) {
			case int, string:
				i, err = d.Break(r, i, 1)
				if err != nil {
					return w, i, false, err
				}
			default:
				break
			}
		}

		if fi < 0 && !first {
			return w, i, false, nil
		}
		if fi < 0 {
			fi++
		}

		//	log.Printf("index frwd %d  %d", fi, i)

		for ; fi < len(f.Filters); fi++ {
			ff := f.Filters[fi]

			switch ff.(type) {
			case int, string:
				i, err = d.Seek(r, i, ff)
				if err != nil {
					return w, i, false, err
				}

				continue
			case Iter:
			default:
				panic(ff)
			}

			var tp byte

			if stack[fi] == nil {
				tp, i, err = d.Type(r, i)
				if err != nil {
					return w, i, false, pe(err, i)
				}

				i, err = d.Enter(r, i, tp)
				if err != nil {
					return w, i, false, pe(err, i)
				}

				stack[fi] = tp
			} else {
				tp = stack[fi].(byte)
			}

			ok, i, err = d.More(r, i, tp)
			if err != nil {
				return w, i, false, pe(err, i)
			}

			if !ok {
				stack[fi] = nil
				continue back
			}

			if tp == json.Object {
				i, err = d.Skip(r, i) // skip key
				if err != nil {
					return w, i, false, pe(err, i)
				}
			}
		}

		w, i, _, err = Dot{}.Next(w, r, i, nil)
		if err != nil {
			return w, i, false, pe(err, i)
		}

		return w, i, true, nil
	}
}

func (f *Index) state() (ss *indexState) {
	need := false

loop:
	for _, sub := range f.Filters {
		switch sub.(type) {
		case int:
		case string:
		default:
			need = true
			break loop
		}
	}

	if !need {
		return nil
	}

	l := len(f.pool)

	if l == 0 {
		ss = &indexState{}
	} else {
		ss = f.pool[l-1]
		f.pool = f.pool[:l-1]
	}

	if cap(ss.sub) < len(f.Filters) {
		ss.sub = make([]indexSub, len(f.Filters))
	}

	ss.sub = ss.sub[:len(f.Filters)]

	return ss
}
