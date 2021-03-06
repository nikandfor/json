package json

// Copy copies one next value from r into w
func Copy(w *Writer, r *Reader) {
	t := r.Type()
	switch t {
	case String:
		w.String(r.NextString())
	case Object:
		w.ObjStart()
		for r.HasNext() {
			w.ObjKey(r.NextString())
			Copy(w, r)
		}
		w.ObjEnd()
	case Array:
		w.ArrayStart()
		for r.HasNext() {
			Copy(w, r)
		}
		w.ArrayEnd()
	default:
		w.RawBytes(r.NextAsBytes())
	}
}
