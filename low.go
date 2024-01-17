package json

import (
	"reflect"
	"unsafe"
)

//go:linkname ifaceIndir internal/abi.(*Type).IfaceIndir
func ifaceIndir(tp unsafe.Pointer) bool

//go:linkname ptrTo reflect.ptrTo
func ptrTo(tp unsafe.Pointer) unsafe.Pointer

//go:linkname unsafeNew reflect.unsafe_New
func unsafeNew(tp unsafe.Pointer) unsafe.Pointer

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
