package jq

import (
	"nikand.dev/go/json"
)

type (
	Object struct {
		Keys []ObjectKey

		pool []*objectState
	}

	ObjectKey struct {
		Key    string
		Filter Filter
	}

	objectState struct {
		sub []objectSub
	}

	objectSub struct {
		i int
		j int
		s State
		n State
	}
)

func NewObject(keys ...ObjectKey) *Object {
	return &Object{Keys: keys}
}

func (f *Object) Next(w, r []byte, st int, state State) ([]byte, int, State, error) {
	var d json.Decoder
	var err error
	var i int

	st = d.SkipSpaces(r, st)
	if st == len(r) {
		return w, st, nil, nil
	}

	w, i, state, err = f.next(w, r, st, state, state == nil)
	if err != nil {
		return w, i, nil, err
	}

	return w, i, state, nil
}

func (f *Object) next(w, r []byte, st int, state State, first bool) (_ []byte, i int, _ State, err error) {
	var d json.Decoder
	var substate any
	var comma bool
	var stack []objectSub

	ss, _ := state.(*objectState)
	if ss != nil {
		stack = ss.sub
	}

	wreset := len(w)

back:
	for {
		j := len(stack) - 1
		w = w[:wreset]

		for ; j >= 0; j-- {
			if stack[j].n != nil {
				stack[j].i = stack[j].j
				stack[j].s = stack[j].n
				break
			}
		}

		if j < 0 && !first {
			break
		}

		//	log.Printf("objectback  %v  %v  %+v", j, first, stack)

		w = append(w, '{')

		comma = false

		for j, kf := range f.Keys {
			pairStart := len(w)

			if comma {
				w = append(w, ',')
			}

			w = append(w, '"')
			w = append(w, kf.Key...)
			w = append(w, '"', ':')

			pairVal := len(w)
			i = st

			if stack != nil {
				i = stack[j].i
				substate = stack[j].s
			}

			w, i, substate, err = kf.Filter.Next(w[:pairVal], r, i, substate)
			//	log.Printf("object key  %v/%v  %v  %v  %+v  %v  %d/%d  %s", j, -1, kf.Key, i, substate, err, len(w), pairVal, w[pairStart:])
			if err != nil {
				return w, i, nil, err
			}

			if substate != nil && stack == nil {
				ss = f.state()
				stack = ss.sub
			}

			if stack != nil {
				stack[j].j = i
				stack[j].n = substate
			}

			if stack != nil && stack[j].n == nil && stack[j].s != nil {
				stack[j].i = 0
				stack[j].s = nil
				//	log.Printf("exhausted key %v  %+v", kf.Key, stack[j])
				continue back
			}

			if len(w) != pairVal {
				comma = true
			} else {
				w = w[:pairStart]
			}
		}

		w = append(w, '}')

		break
	}

	state = nil

	for _, s := range stack {
		if s.n != nil {
			state = ss
			break
		}
	}

	if state == nil && ss != nil {
		f.pool = append(f.pool, ss)
	}

	i = st

	if state == nil {
		i, err = d.Skip(r, st)
		if err != nil {
			return w, i, nil, pe(err, i)
		}
	}

	return w, i, state, nil
}

func (f *Object) state() (ss *objectState) {
	l := len(f.pool)

	if l == 0 {
		ss = &objectState{}
	} else {
		ss = f.pool[l-1]
		f.pool = f.pool[:l-1]
	}

	if cap(ss.sub) < len(f.Keys) {
		ss.sub = make([]objectSub, len(f.Keys))
	}

	ss.sub = ss.sub[:len(f.Keys)]

	return ss
}
