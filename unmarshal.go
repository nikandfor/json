package json

import (
	"encoding/base64"
	"fmt"
	"io"
	"reflect"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
	"unsafe"
)

var (
	interfaceVal interface{}
	sliceType    = reflect.SliceOf(reflect.TypeOf(&interfaceVal).Elem())
	mapType      = reflect.MapOf(reflect.TypeOf("string"), reflect.TypeOf(&interfaceVal).Elem())
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
	//	log.Printf("unmarshal: %d+%d/%d  -> %v (%v)", r.ref, r.i, r.end, rv, rv.Type())
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

	var fptr uintptr
	if rv.CanAddr() {
		fptr = rv.UnsafeAddr()
	}

	switch k := rv.Kind(); k {
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
		switch k {
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
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8,
		reflect.Uintptr:
		q, err := r.Uint64()
		if err != nil {
			return err
		}
		switch k {
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
		if k == reflect.Float64 {
			*(*float64)(unsafe.Pointer(fptr)) = float64(q)
		} else {
			*(*float32)(unsafe.Pointer(fptr)) = float32(q)
		}
	case reflect.Map:
		return r.unmarshalMap(rv)
	case reflect.Interface:
		var v interface{}
		switch r.Type() {
		case String:
			v = string(r.NextString())
		case Number:
			v = Num(r.NextNumber())
		case Bool:
			q, err := r.Bool()
			if err != nil {
				return err
			}
			v = q
		case Array:
			q := reflect.New(sliceType).Elem()
			err := r.unmarshal(q)
			if err != nil {
				return err
			}
			v = q.Interface()
		case Object:
			q := reflect.New(mapType).Elem()
			err := r.unmarshal(q)
			if err != nil {
				return err
			}
			v = q.Interface()
		case Null:
			// we'll set `v` that is nil
		case None:
			return r.Err()
		}
		rv.Set(reflect.ValueOf(&v).Elem())
	default:
		panic(rv.Kind())
	}
	return nil
}

func (r *Reader) unmarshalMap(rv reflect.Value) error {
	rv.Set(reflect.MakeMap(rv.Type()))

	for r.HasNext() {
		k := r.NextString()
		rk := reflect.ValueOf(string(k))

		//	log.Printf("set map %s -> %v  %v", k, r.Type(), rv.Type())

		switch r.Type() {
		case String:
			v := string(r.NextString())
			rv.SetMapIndex(rk, reflect.ValueOf(v))
		case Number:
			elt := rv.Type().Elem()
			rval := reflect.New(elt).Elem()
			num := Num(r.NextNumber())
			switch elt.Kind() {
			case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
				q, err := num.Int64()
				if err != nil {
					return err
				}
				rval.SetInt(q)
				rv.SetMapIndex(rk, rval)
			case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
				q, err := num.Uint64()
				if err != nil {
					return err
				}
				rval.SetUint(q)
				rv.SetMapIndex(rk, rval)
			case reflect.Float64, reflect.Float32:
				q, err := num.Float64()
				if err != nil {
					return err
				}
				rval.SetFloat(q)
				rv.SetMapIndex(rk, rval)
			default:
				rv.SetMapIndex(rk, reflect.ValueOf(num))
			}
		case Bool:
			v, err := r.Bool()
			if err != nil {
				return err
			}
			rv.SetMapIndex(rk, reflect.ValueOf(v))
		case Null:
			r.Skip()
			var v interface{}
			rv.SetMapIndex(rk, reflect.ValueOf(&v).Elem())
		case Array:
			v := reflect.New(sliceType).Elem()
			err := r.unmarshal(v)
			if err != nil {
				return err
			}
			rv.SetMapIndex(rk, v)
		case Object:
			v := reflect.New(mapType).Elem()
			err := r.unmarshal(v)
			if err != nil {
				return err
			}
			rv.SetMapIndex(rk, v)
		case None:
			return r.Err()
		default:
			panic(r.Type())
		}
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
			case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8,
				reflect.Uintptr:
				q, err := r.Uint64()
				if err != nil {
					return err
				}
				switch f.Kind {
				case reflect.Uint:
					*(*uint)(unsafe.Pointer(fptr)) = uint(q)
				case reflect.Uintptr:
					*(*uintptr)(unsafe.Pointer(fptr)) = uintptr(q)
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
		case reflect.Uintptr:
			*(*uintptr)(unsafe.Pointer(fptr)) = 0
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
		buf := rv.Bytes()
		res := buf
		rn := 0
		sr := r.Base64Reader(base64.RawStdEncoding)
		for {
			n, err := sr.Read(buf)
			rn += n
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}

			res = append(res, 0)
			buf = res[rn:]
		}
		if res == nil {
			res = make([]byte, 0)
		}
		rv.SetBytes(res[:rn])
		return sr.Close()
	}

	if tp != Array {
		return r.setErr(ErrIncompatibleTypes)
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
	Name      []byte
	Kind      reflect.Kind
	Ptr       uintptr
	FastPath  bool
	FastErase bool
	OmitEmpty bool
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
			reflect.Float64, reflect.Float32,
			reflect.Uintptr:
			sf.FastPath = true
			sf.FastErase = true
		}

		n := f.Name
		if t, ok := f.Tag.Lookup("json"); ok {
			t := strings.Split(t, ",")
			if t[0] == "-" {
				continue
			}
			n = t[0]
			for _, o := range t[1:] {
				switch o {
				case "omitempty":
					sf.OmitEmpty = true
				}
			}
		} else {
			r, sz := utf8.DecodeRuneInString(n)
			if unicode.IsLower(r) {
				continue
			}
			ln := string(unicode.ToLower(r)) + n[sz:]
			m.m[ln] = sf
		}
		sf.Name = []byte(n)
		m.m[n] = sf

		m.s = append(m.s, sf)
	}

	structMaps[t] = m

	return m
}

func (m *structMap) String() string {
	var b strings.Builder
	b.WriteByte('[')
	for i, f := range m.s {
		if i != 0 {
			b.WriteByte(' ')
		}
		b.WriteString(f.String())
	}
	b.WriteByte(']')
	return b.String()
}

func (f structField) String() string {
	return fmt.Sprintf("{%d %s %v}", f.I, f.Name, f.Kind)
}
