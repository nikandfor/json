package json

func (r *Reader) Inspect(cb func(r *Reader) bool, ks ...interface{}) bool {
	r.Lock()
	r.Get(ks...)
	rev := cb(r)
	if rev {
		r.Return()
	} else {
		r.Unlock()
		r.GoOut(len(ks))
	}
	return rev
}
