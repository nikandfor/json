package json

import (
	"reflect"
	"unsafe"
)

type (
	sliceHeader struct {
		p    unsafe.Pointer
		l, c int
	}
)

//go:linkname ifaceIndir internal/abi.(*Type).IfaceIndir
func ifaceIndir(tp unsafe.Pointer) bool

//go:linkname tpPtrTo reflect.ptrTo
func tpPtrTo(tp unsafe.Pointer) unsafe.Pointer

//go:linkname unsafe_New reflect.unsafe_New
func unsafe_New(tp unsafe.Pointer) unsafe.Pointer

//go:linkname findObject runtime.findObject
func findObject(ptr unsafe.Pointer, _, _ uintptr) (base, _, _ uintptr)

//go:linkname tpSize reflect.(*rtype).Size
func tpSize(tp unsafe.Pointer) uintptr

//go:linkname tpKind reflect.(*rtype).Kind
func tpKind(tp unsafe.Pointer) reflect.Kind

//go:linkname tpElem internal/abi.(*Type).Elem
func tpElem(tp unsafe.Pointer) unsafe.Pointer

//go:linkname tpName reflect.(*rtype).Name
func tpName(tp unsafe.Pointer) string

//go:linkname tpString reflect.(*rtype).String
func tpString(tp unsafe.Pointer) string

//go:linkname toType reflect.toType
func toType(tp unsafe.Pointer) reflect.Type

//go:linkname arrLen reflect.(*rtype).Len
func arrLen(tp unsafe.Pointer) int

//go:linkname growslice reflect.growslice
func growslice(tp unsafe.Pointer, s sliceHeader, more int) sliceHeader

//go:linkname typedmemclr runtime.typedmemclr
func typedmemclr(tp, ptr unsafe.Pointer)

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

func vtp(v interface{}) unsafe.Pointer {
	i := *(*struct {
		t, d unsafe.Pointer
	})(unsafe.Pointer(&v))

	return i.t
}

func vptr(v interface{}) unsafe.Pointer {
	i := *(*struct {
		t, d unsafe.Pointer
	})(unsafe.Pointer(&v))

	return i.d
}
