package json

import (
	"encoding/base64"
	"io"
)

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

type Base64Writer struct {
	StringWriter
	e io.WriteCloser
}

func (w *Writer) Base64Writer(enc *base64.Encoding) Base64Writer {
	s := Base64Writer{StringWriter: w.StringWriter()}
	s.e = base64.NewEncoder(enc, s.StringWriter)
	return s
}

func (w Base64Writer) Write(p []byte) (int, error) {
	return w.e.Write(p)
}

func (w Base64Writer) Close() error {
	if w.err != nil {
		return w.Err()
	}
	err := w.e.Close()
	w.err = err
	err = w.StringWriter.Close()
	if w.err == nil {
		w.err = err
	}
	return w.Err()
}
