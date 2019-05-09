package json

import (
	"encoding/base64"
	"io"
)

// StringWriter allows to write string value as an io.WriteCloser interface
type StringWriter struct {
	*Writer
}

// StringWriter allows to write string value as an io.WriteCloser interface
func (w *Writer) StringWriter() StringWriter {
	s := StringWriter{w}
	w.valueStart()
	w.add([]byte{'"'})
	return s
}

// Write writes data into the big string
func (w StringWriter) Write(p []byte) (int, error) {
	w.safeadd(p)
	return len(p), nil
}

// Close must be called to finish the string
func (w StringWriter) Close() error {
	w.add([]byte{'"'})
	w.valueEnd()
	return w.Err()
}

// Base64Writer allows to write raw bytes stream as an base64 encoded string
type Base64Writer struct {
	StringWriter
	e io.WriteCloser
}

// Base64Writer allows to write raw bytes stream as an base64 encoded string
func (w *Writer) Base64Writer(enc *base64.Encoding) Base64Writer {
	s := Base64Writer{StringWriter: w.StringWriter()}
	s.e = base64.NewEncoder(enc, s.StringWriter)
	return s
}

// Write encodes data into base64 encoding and writes it into the big string
func (w Base64Writer) Write(p []byte) (int, error) {
	return w.e.Write(p)
}

// Close must be called to finish the string
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
