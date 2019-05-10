package json

// Inspect inspects requested value and returns back or skips object according
// to callback returned value: true - return back, false - skip the object.
// Whole data from the beginning of the object until inspected value (including)
// is stored in memory and buffer doesn't shrunk back
func (r *Reader) Inspect(cb func(r *Reader) bool, ks ...interface{}) bool {
	r.Lock()
	r.Search(ks...)
	rev := cb(r)
	if rev {
		r.Return()
	} else {
		r.Unlock()
		r.GoOut(len(ks))
	}
	return rev
}

// Search is a shortcut for Wrap(data).Search(keys...)
func Search(data []byte, keys ...interface{}) *Reader {
	return Wrap(data).Search(keys...)
}
