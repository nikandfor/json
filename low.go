package json

import (
	"unsafe"
)

//go:linkname ifaceIndir internal/abi.(*Type).IfaceIndir
func ifaceIndir(tp unsafe.Pointer) bool

//go:linkname unsafeNew reflect.unsafe_New
func unsafeNew(tp unsafe.Pointer) unsafe.Pointer

//go:linkname findObject runtime.findObject
func findObject(ptr, _, _ uintptr) (base, _, _ uintptr)

func pack(t, d unsafe.Pointer) interface{} {
	return *(*interface{})(unsafe.Pointer(&struct {
		t, d unsafe.Pointer
	}{
		t: t, d: d,
	}))
}

func unpack(v interface{}) (t, d unsafe.Pointer) {
	i := *(*struct {
		t, d unsafe.Pointer
	})(unsafe.Pointer(&v))

	return i.t, i.d
}

func vptr(v interface{}) unsafe.Pointer {
	i := *(*struct {
		t, d unsafe.Pointer
	})(unsafe.Pointer(&v))

	return i.d
}
