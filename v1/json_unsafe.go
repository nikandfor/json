package json

import "unsafe"

func WrapStringUnsafe(s string) *Value {
	b := *(*[]byte)(unsafe.Pointer(&s))
	return Wrap(b)
}
