package json

import (
	"reflect"
	"unsafe"
)

func UnsafeStringToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: (*reflect.StringHeader)(unsafe.Pointer(&s)).Data,
		Len:  len(s),
		Cap:  len(s),
	}))
}
