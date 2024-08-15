package jq

import (
	"nikand.dev/go/json"
)

type (
	// Comma passes the same input to each of the Filters and combines returned values.
	Comma struct {
		Filters []Filter
		pool    []*commaState
	}

	commaState struct {
		st  int
		fi  int
		sub State
	}
)

func NewComma(fs ...Filter) *Comma {
	return &Comma{
		Filters: fs,
	}
}

func (f *Comma) Next(w, r []byte, st int, state State) (_ []byte, i int, _ State, err error) {
	st = json.SkipSpaces(r, st)
	if st == len(r) {
		return w, st, nil, nil
	}

	i = st
	wreset := len(w)

	ss, _ := state.(*commaState)
	if ss == nil {
		ss = f.state()
		ss.st = st
		state = ss
	}

	for {
		w, i, ss.sub, err = f.Filters[ss.fi].Next(w, r, i, ss.sub)
		//	log.Printf("comma filter %d  %d->%d  %v  %v  => %s", ss.fi, st, i, ss.sub, err, w[wreset:])
		if err != nil {
			return w, i, state, err
		}

		if ss.sub == nil {
			ss.fi++
		}

		if ss.sub == nil && ss.fi < len(f.Filters) {
			i = ss.st
		}

		if ss.sub == nil && ss.fi < len(f.Filters) && len(w) == wreset {
			continue
		}

		break
	}

	if ss.fi == len(f.Filters) {
		state = nil
		f.pool = append(f.pool, ss)
	}

	return w, i, state, nil
}

func (f *Comma) state() (ss *commaState) {
	l := len(f.pool)
	if l == 0 {
		ss = &commaState{}
	} else {
		ss = f.pool[l-1]
		f.pool = f.pool[:l-1]

		ss.fi = 0
	}

	return ss
}
