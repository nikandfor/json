package json

import (
	"encoding/base64"
	"reflect"
	"strconv"
	"unsafe"
)

func Marshal(r interface{}) ([]byte, error) {
	w := NewWriter(nil)
	err := w.Marshal(r)
	return w.Bytes(), err
}

func (w *Writer) Marshal(r interface{}) error {
	if r == nil {
		w.Null()
		return w.Err()
	}
	return w.marshal(reflect.ValueOf(r))
}

func (w *Writer) marshal(rv reflect.Value) error {
	//	log.Printf("marshal: %v %v", rv.Type(), rv)
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			w.Null()
			return nil
		}
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Struct:
		return w.marshalStruct(rv)
	case reflect.Array, reflect.Slice:
		return w.marshalSlice(rv)
	case reflect.String:
		w.String(UnsafeStringToBytes(rv.String()))
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		i := rv.Int()
		s := strconv.FormatInt(i, 10)
		w.Number(UnsafeStringToBytes(s))
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		i := rv.Uint()
		s := strconv.FormatUint(i, 10)
		w.Number(UnsafeStringToBytes(s))
	case reflect.Float64, reflect.Float32:
		bits := 64
		if rv.Kind() == reflect.Float32 {
			bits = 32
		}
		f := rv.Float()
		s := strconv.FormatFloat(f, 'g', -1, bits)
		w.Number(UnsafeStringToBytes(s))
	case reflect.Bool:
		w.Bool(rv.Bool())
	default:
		panic(rv.Kind())
	}
	return w.Err()
}

func (w *Writer) marshalStruct(rv reflect.Value) error {
	m := getStructMap(rv.Type())
	ptr := rv.UnsafeAddr()
	//	log.Printf("struct: %+v", m)

	w.ObjStart()
	for i, f := range m.s {
		w.ObjKey(f.Name)

		fptr := ptr + f.Ptr

		switch f.Kind {
		case reflect.String:
			w.String([]byte(rv.Field(i).String()))
		case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
			var q int64
			switch f.Kind {
			case reflect.Int:
				q = (int64)(*(*int)(unsafe.Pointer(fptr)))
			case reflect.Int64:
				q = (int64)(*(*int64)(unsafe.Pointer(fptr)))
			case reflect.Int32:
				q = (int64)(*(*int32)(unsafe.Pointer(fptr)))
			case reflect.Int16:
				q = (int64)(*(*int16)(unsafe.Pointer(fptr)))
			case reflect.Int8:
				q = (int64)(*(*int8)(unsafe.Pointer(fptr)))
			}
			s := strconv.FormatInt(q, 10)
			w.Number(UnsafeStringToBytes(s))
		case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
			var q uint64
			switch f.Kind {
			case reflect.Int:
				q = (uint64)(*(*uint)(unsafe.Pointer(fptr)))
			case reflect.Int64:
				q = (uint64)(*(*uint64)(unsafe.Pointer(fptr)))
			case reflect.Int32:
				q = (uint64)(*(*uint32)(unsafe.Pointer(fptr)))
			case reflect.Int16:
				q = (uint64)(*(*uint16)(unsafe.Pointer(fptr)))
			case reflect.Int8:
				q = (uint64)(*(*uint8)(unsafe.Pointer(fptr)))
			}
			s := strconv.FormatUint(q, 10)
			w.Number(UnsafeStringToBytes(s))
		case reflect.Float64, reflect.Float32:
			bits := 64
			if f.Kind == reflect.Float32 {
				bits = 32
			}
			var q float64
			if f.Kind == reflect.Float64 {
				q = (float64)(*(*float64)(unsafe.Pointer(fptr)))
			} else {
				q = (float64)(*(*float32)(unsafe.Pointer(fptr)))
			}
			s := strconv.FormatFloat(q, 'g', -1, bits)
			w.Number(UnsafeStringToBytes(s))
		case reflect.Bool:
			q := *(*bool)(unsafe.Pointer(fptr))
			w.Bool(q)
		default:
			w.marshal(rv.Field(i))
		}
	}
	w.ObjEnd()

	return w.Err()
}

func (w *Writer) marshalSlice(rv reflect.Value) error {
	elk := rv.Type().Elem().Kind()
	if elk == reflect.Uint8 {
		sw := w.Base64Writer(base64.RawStdEncoding)
		sw.Write(rv.Bytes())
		return sw.Close()
	}
	w.ArrayStart()
	l := rv.Len()
	for i := 0; i < l; i++ {
		vi := rv.Index(i)
		switch elk {
		case reflect.String:
			w.String(UnsafeStringToBytes(vi.String()))
		case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
			i := vi.Int()
			s := strconv.FormatInt(i, 10)
			w.Number(UnsafeStringToBytes(s))
		case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
			i := vi.Uint()
			s := strconv.FormatUint(i, 10)
			w.Number(UnsafeStringToBytes(s))
		case reflect.Float64, reflect.Float32:
			bits := 64
			if rv.Kind() == reflect.Float32 {
				bits = 32
			}
			f := vi.Float()
			s := strconv.FormatFloat(f, 'g', -1, bits)
			w.Number(UnsafeStringToBytes(s))
		default:
			w.marshal(vi.Index(i))
		}
	}
	w.ArrayEnd()
	return w.Err()
}
