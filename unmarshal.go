package json

import (
	"encoding/base64"
	"reflect"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
	"unsafe"
)

// Unmarshal unmarshals data info r
func Unmarshal(data []byte, r interface{}) error {
	return Wrap(data).Unmarshal(r)
}

// UnmarshalNoZero unmarshals data info r but not clean struct fields that were not set in data
func UnmarshalNoZero(data []byte, r interface{}) error {
	w := Wrap(data)
	w.nozero = true
	return w.Unmarshal(r)
}

// Unmarshal reads and unmarshals value into res
func (r *Reader) Unmarshal(res interface{}) error {
	rv := reflect.ValueOf(res)
	return r.unmarshal(rv)
}

func (r *Reader) unmarshal(rv reflect.Value) error {
	//	log.Printf("unmarshal: %d+%d/%d  -> %v", r.ref, r.i, r.end, rv)
	for rv.Kind() == reflect.Ptr {
		if r.IsNull() && rv.IsNil() {
			return nil
		}

		if r.IsNull() {
			rv = rv.Elem()
			rv.Set(reflect.Zero(rv.Type()))
			return nil
		}

		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}

		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Struct:
		return r.unmarshalStruct(rv)
	case reflect.String:
		q, err := r.CheckString()
		if err != nil {
			return err
		}
		rv.SetString(q)
	case reflect.Slice:
		return r.unmarshalArray(rv)
	case reflect.Array:
		return r.unmarshalArray(rv)
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		q, err := r.Int64()
		if err != nil {
			return err
		}
		rv.SetInt(q)
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		q, err := r.Uint64()
		if err != nil {
			return err
		}
		rv.SetUint(q)
	}
	return nil
}

func (r *Reader) unmarshalStruct(rv reflect.Value) error {
	m := getStructMap(rv.Type())
	var visbuf [20]bool
	var vis []bool
	if n := rv.NumField(); n < len(visbuf) {
		vis = visbuf[:n]
	} else {
		vis = make([]bool, n)
	}

	ptr := rv.UnsafeAddr()

	for r.HasNext() {
		k := r.NextString()

		f, ok := m.m[string(k)]
		//	log.Printf("struct key: %q ok %v", k, ok)
		if !ok {
			r.Skip()
			continue
		}

		vis[f.I] = true

		if f.FastPath {
			fptr := ptr + f.Ptr

			switch f.Kind {
			case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
				q, err := r.Int64()
				if err != nil {
					return err
				}
				switch f.Kind {
				case reflect.Int:
					*(*int)(unsafe.Pointer(fptr)) = int(q)
				case reflect.Int64:
					*(*int64)(unsafe.Pointer(fptr)) = q
				case reflect.Int32:
					*(*int32)(unsafe.Pointer(fptr)) = int32(q)
				case reflect.Int16:
					*(*int16)(unsafe.Pointer(fptr)) = int16(q)
				case reflect.Int8:
					*(*int8)(unsafe.Pointer(fptr)) = int8(q)
				}
			case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
				q, err := r.Uint64()
				if err != nil {
					return err
				}
				switch f.Kind {
				case reflect.Uint:
					*(*uint)(unsafe.Pointer(fptr)) = uint(q)
				case reflect.Uint64:
					*(*uint64)(unsafe.Pointer(fptr)) = q
				case reflect.Uint32:
					*(*uint32)(unsafe.Pointer(fptr)) = uint32(q)
				case reflect.Uint16:
					*(*uint16)(unsafe.Pointer(fptr)) = uint16(q)
				case reflect.Uint8:
					*(*uint8)(unsafe.Pointer(fptr)) = uint8(q)
				}
			case reflect.Float64, reflect.Float32:
				q, err := r.Float64()
				if err != nil {
					return err
				}
				if f.Kind == reflect.Float64 {
					*(*float64)(unsafe.Pointer(fptr)) = float64(q)
				} else {
					*(*float32)(unsafe.Pointer(fptr)) = float32(q)
				}
			case reflect.Bool:
				q, err := r.Bool()
				if err != nil {
					return err
				}
				*(*bool)(unsafe.Pointer(fptr)) = q
			}
			continue
		}

		if err := r.unmarshal(rv.Field(f.I)); err != nil {
			return err
		}
	}

	if r.nozero {
		return nil
	}

	for i, vis := range vis {
		if vis {
			continue
		}
		f := m.s[i]
		fptr := ptr + f.Ptr
		switch f.Kind {
		case reflect.Ptr:
			*(*uintptr)(unsafe.Pointer(fptr)) = 0
		case reflect.String, reflect.Slice:
			//	if rv.Field(i).Len() != 0 {
			//		rv.Field(i).Set(reflect.Zero(rv.Field(i).Type()))
			//	}
			sl := (*sliceHeader)(unsafe.Pointer(fptr))
			sl.Data = 0
			sl.Len = 0
			if f.Kind == reflect.Slice {
				sl.Cap = 0
			}
		case reflect.Int:
			*(*int)(unsafe.Pointer(fptr)) = 0
		case reflect.Int64:
			*(*int64)(unsafe.Pointer(fptr)) = 0
		case reflect.Int32:
			*(*int32)(unsafe.Pointer(fptr)) = 0
		case reflect.Int16:
			*(*int16)(unsafe.Pointer(fptr)) = 0
		case reflect.Int8:
			*(*int8)(unsafe.Pointer(fptr)) = 0
		case reflect.Uint:
			*(*uint)(unsafe.Pointer(fptr)) = 0
		case reflect.Uint64:
			*(*uint64)(unsafe.Pointer(fptr)) = 0
		case reflect.Uint32:
			*(*uint32)(unsafe.Pointer(fptr)) = 0
		case reflect.Uint16:
			*(*uint16)(unsafe.Pointer(fptr)) = 0
		case reflect.Uint8:
			*(*uint8)(unsafe.Pointer(fptr)) = 0
		case reflect.Float64:
			*(*float64)(unsafe.Pointer(fptr)) = 0
		case reflect.Float32:
			*(*float32)(unsafe.Pointer(fptr)) = 0
		case reflect.Bool:
			*(*bool)(unsafe.Pointer(fptr)) = false
		default:
			rv.Field(i).Set(reflect.Zero(rv.Field(i).Type()))
		}
	}

	return nil
}

func (r *Reader) unmarshalArray(rv reflect.Value) error {
	elt := rv.Type().Elem()

	tp := r.Type()

	if elt.Kind() == reflect.Uint8 && tp == String {
		bs := r.NextString()
		n := base64.StdEncoding.DecodedLen(len(bs))
		if n > rv.Cap() || rv.IsNil() {
			rv.Set(reflect.MakeSlice(rv.Type(), n, n))
		} else {
			rv.Set(rv.Slice(0, n))
		}
		n1, err := base64.StdEncoding.Decode(rv.Bytes(), bs)
		if err != nil {
			return err
		}
		if n1 != n {
			rv.Set(rv.Slice(0, n1))
		}
		return nil
	}

	// usual array
	zero := reflect.Zero(elt)
	//	if rv.Kind() == reflect.Slice {
	//		rv.Set(rv.Slice(0, rv.Cap()))
	//	}
	var baseptr uintptr
	if rv.Len() != 0 {
		baseptr = rv.Index(0).UnsafeAddr()
	}
	size := rv.Type().Elem().Size()
	elkind := rv.Type().Elem().Kind()

	j := 0
	for r.HasNext() {
		if j == rv.Len() {
			rv.Set(reflect.Append(rv, zero))
			rv.Set(rv.Slice(0, rv.Cap()))
			baseptr = rv.Index(0).UnsafeAddr()
		}

		ptr := baseptr + uintptr(j)*size
		switch elkind {
		case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
			q, err := r.Int64()
			if err != nil {
				return err
			}
			switch elkind {
			case reflect.Int:
				*(*int)(unsafe.Pointer(ptr)) = int(q)
			case reflect.Int64:
				*(*int64)(unsafe.Pointer(ptr)) = q
			case reflect.Int32:
				*(*int32)(unsafe.Pointer(ptr)) = int32(q)
			case reflect.Int16:
				*(*int16)(unsafe.Pointer(ptr)) = int16(q)
			case reflect.Int8:
				*(*int8)(unsafe.Pointer(ptr)) = int8(q)
			}
		case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
			q, err := r.Uint64()
			if err != nil {
				return err
			}
			switch elkind {
			case reflect.Uint:
				*(*uint)(unsafe.Pointer(ptr)) = uint(q)
			case reflect.Uint64:
				*(*uint64)(unsafe.Pointer(ptr)) = q
			case reflect.Uint32:
				*(*uint32)(unsafe.Pointer(ptr)) = uint32(q)
			case reflect.Uint16:
				*(*uint16)(unsafe.Pointer(ptr)) = uint16(q)
			case reflect.Uint8:
				*(*uint8)(unsafe.Pointer(ptr)) = uint8(q)
			}
		case reflect.Float64, reflect.Float32:
			q, err := r.Float64()
			if err != nil {
				return err
			}
			if elkind == reflect.Float64 {
				*(*float64)(unsafe.Pointer(ptr)) = float64(q)
			} else {
				*(*float32)(unsafe.Pointer(ptr)) = float32(q)
			}
		case reflect.Bool:
			q, err := r.Bool()
			if err != nil {
				return err
			}
			*(*bool)(unsafe.Pointer(ptr)) = q
		default:
			if err := r.unmarshal(rv.Index(j)); err != nil {
				return err
			}
		}

		j++
	}

	if j < rv.Len() {
		if rv.Kind() == reflect.Slice {
			rv.Set(rv.Slice(0, j))
		} else {
			for i := j; i < rv.Len(); i++ {
				rv.Index(i).Set(zero)
			}
		}
	}

	if rv.Kind() == reflect.Slice && rv.IsNil() {
		rv.Set(reflect.MakeSlice(rv.Type(), 0, 0))
	}

	return nil
}

type structMap struct {
	s []structField
	m map[string]structField
}

type structField struct {
	I         int
	Kind      reflect.Kind
	Ptr       uintptr
	FastPath  bool
	FastErase bool
}

type sliceHeader struct {
	Data uintptr
	Len  int
	Cap  int
}

var (
	mapMu      sync.Mutex
	structMaps = map[reflect.Type]*structMap{}
)

func getStructMap(t reflect.Type) *structMap {
	mapMu.Lock()
	m, ok := structMaps[t]
	if ok {
		mapMu.Unlock()
		return m
	}
	defer mapMu.Unlock()

	m = &structMap{
		m: make(map[string]structField),
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		sf := structField{
			I:    i,
			Kind: f.Type.Kind(),
			Ptr:  f.Offset,
		}

		switch f.Type.Kind() {
		case reflect.Ptr,
			reflect.Slice:
			sf.FastErase = true
		case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8,
			reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8,
			reflect.Float64, reflect.Float32:
			sf.FastPath = true
			sf.FastErase = true
		}

		m.s = append(m.s, sf)

		n := f.Name
		if t, ok := f.Tag.Lookup("json"); ok {
			t := strings.Split(t, ",")
			if t[0] == "-" {
				continue
			}
			n = t[0]
		} else {
			r, sz := utf8.DecodeRuneInString(n)
			ln := string(unicode.ToLower(r)) + n[sz:]
			m.m[ln] = sf
		}
		m.m[n] = sf
	}

	structMaps[t] = m

	return m
}
