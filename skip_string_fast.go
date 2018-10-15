// +build !strict

package json

func (r *Reader) skipString() {
	//	log.Printf("Skip stri %d+%d/%d", r.ref, r.i, r.end)
	esc := true
start:
	i := r.i
	for i < r.end {
		c := r.b[i]
		i++
		//	log.Printf("skip str0 %d+%d/%d '%c'", r.ref, r.i, r.end, c)
		switch {
		case c == '\\':
			if esc {
				esc = false
				continue
			}
			esc = true
		case c == '"':
			if esc {
				esc = false
				continue
			}
			r.i = i
			return
		}
	}
	r.i = i
	if r.more() {
		goto start
	}
}
