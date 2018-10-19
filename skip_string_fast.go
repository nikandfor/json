// +build !strict

package json

func (r *Reader) skipString(esc bool) {
	//	log.Printf("Skip stri %d+%d/%d", r.ref, r.i, r.end)
start:
	s := r.i
	b := r.b[s:r.end]
	for i, c := range b {
		//	log.Printf("skip str0 %d+%d/%d '%c'", r.ref, r.i, r.end, c)
		if c == '"' {
			if esc {
				esc = false
				continue
			}
			r.i = s + i + 1
			return
		} else if c == '\\' {
			if esc {
				esc = false
				continue
			}
			esc = true
		}
	}
	r.i = r.end
	if r.more() {
		goto start
	}
}
