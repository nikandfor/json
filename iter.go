package json

type ArrayIter struct {
	*Reader
}

func (r *Reader) ArrayIter() ArrayIter {
	if r.Type() != '[' {
		return ArrayIter{}
	}
	r.i++
	return ArrayIter{r}
}

func (r ArrayIter) HasNext() bool {
	if r.Reader == nil {
		return false
	}
	//	log.Printf("HasNxt: %d+%d '%s'", r.ref, r.i, r.b)
start:
	for r.i < r.end {
		c := r.b[r.i]
		switch c {
		case ' ', '\t', '\n':
			r.i++
			continue
		}
		switch c {
		case ']':
			r.i++
			return false
		case ',':
			r.i++
			continue
		}
		return true
	}
	if r.more() {
		goto start
	}
	return false
}

type ObjectIter struct {
	*Reader
}

func (r *Reader) ObjectIter() ObjectIter {
	if r.Type() != '{' {
		return ObjectIter{}
	}
	r.i++
	return ObjectIter{r}
}

func (r ObjectIter) HasNext() bool {
	if r.Reader == nil {
		return false
	}
	//	log.Printf("HasNxt: %d+%d '%s'", r.ref, r.i, r.b)
start:
	for r.i < r.end {
		c := r.b[r.i]
		switch c {
		case ' ', '\t', '\n':
			r.i++
			continue
		}
		switch c {
		case '}':
			r.i++
			return false
		case ',':
			r.i++
			continue
		}
		return true
	}
	if r.more() {
		goto start
	}
	return false
}
