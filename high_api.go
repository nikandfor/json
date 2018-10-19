package json

// Inspect inspects requested value and returns back or skips object according
// to callback returned value: true - return back, false - skip the object.
// Whole data from the beginning of the object until inspected value (including)
// is stored in memory and buffer doesn't shrinked back
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

// Get is a shortcut for Wrap(data).Get(keys...)
func Get(data []byte, keys ...interface{}) *Reader {
	return Wrap(data).Get(keys...)
}
