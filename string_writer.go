package json

type StringWriter struct {
	*Writer
}

func (w *Writer) StringWriter() StringWriter {
	s := StringWriter{w}
	w.valueStart()
	w.add([]byte{'"'})
	return s
}

func (w StringWriter) Write(p []byte) (int, error) {
	w.add(p)
	return len(p), nil
}

func (w StringWriter) Close() error {
	w.add([]byte{'"'})
	w.valueEnd()
	return w.Err()
}
