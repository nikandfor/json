package jq

type (
	// Pipe passes the input to the first Filter.
	// Then it passes the prevoius filter result(s) ot the input of the next Filter.
	// And then it returns the last filter result(s).
	Pipe struct {
		Filters []Filter
		pool    []*pipeState
	}

	pipeState struct {
		b   []byte
		sub []pipeSub
	}

	pipeSub struct {
		St    int
		State State
		Bst   int
	}
)

func NewPipe(fs ...Filter) *Pipe {
	return &Pipe{
		Filters: fs,
	}
}

func (f *Pipe) Next(w, r []byte, st int, state State) (_ []byte, i int, _ State, err error) {
	ss, _ := state.(*pipeState)
	if ss == nil {
		ss = f.state()
	}

	last := len(f.Filters) - 1
	fi := last

	for fi >= 0 && ss.sub[fi].State == nil {
		fi--
	}

	if fi == -1 && state == nil {
		fi++
	}

	if fi == -1 {
		return w, st, nil, nil
	}

	for ; fi < len(f.Filters); fi++ {
		ff := f.Filters[fi]

		if ss.sub[fi].State == nil {
			ss.sub[fi].Bst = len(ss.b)
		}

		wbst := ss.sub[fi].Bst
		ss.b = ss.b[:wbst]

		wb := cbuf(fi == last, w, ss.b)
		rb := cbuf(fi == 0, r, ss.b)
		rbst := cint(fi == 0, st, ss.sub[fi].St)

		if fi < last {
			ss.sub[fi+1].St = len(ss.b)
		}

		//	log.Printf("pipe args   #%d  %3v  %-30q  %q  sub %+v", fi, rbst, rb, ss.b, ss.sub)

		wb, ss.sub[fi].St, ss.sub[fi].State, err = ff.Next(wb, rb, rbst, ss.sub[fi].State)
		//	log.Printf("pipe filter #%d  %3v  %-30q  %q  err %v", fi, ss.sub[fi].St, wb, ss.b, err)
		if err != nil {
			return w, i, state, err
		}

		w = cbuf(fi == last, wb, w)
		ss.b = cbuf(fi != last, wb, ss.b)
	}

	i = ss.sub[0].St
	state = nil

	for _, sub := range ss.sub {
		if sub.State != nil {
			state = ss
			break
		}
	}

	if state == nil {
		f.pool = append(f.pool, ss)
	}

	return w, i, state, nil
}

func (f *Pipe) state() (ss *pipeState) {
	l := len(f.pool)

	if l == 0 {
		ss = &pipeState{}
	} else {
		ss = f.pool[l-1]
		f.pool = f.pool[:l-1]

		ss.b = ss.b[:0]
	}

	if cap(ss.sub) < len(f.Filters) {
		ss.sub = make([]pipeSub, len(f.Filters))
	}

	ss.sub = ss.sub[:len(f.Filters)]

	return ss
}
