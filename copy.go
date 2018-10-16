package json

func Copy(w *Writer, r *Reader) {
	t := r.Type()
	switch t {
	case String:
		w.String(r.NextString())
	case ObjStart:
		w.ObjStart()
		for r.HasNext() {
			w.ObjKey(r.NextString())
			Copy(w, r)
		}
		w.ObjEnd()
	case ArrayStart:
		w.ArrayStart()
		for r.HasNext() {
			Copy(w, r)
		}
		w.ArrayEnd()
	}
}
