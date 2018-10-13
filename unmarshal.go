package json

import (
	"encoding/base64"
	"reflect"
	"strings"
	"sync"
	"unsafe"
)

func Unmarshal(data []byte, v interface{}) error {
	return Wrap(data).Unmarshal(v)
}

func (v *Value) Unmarshal(r interface{}) error {
	/*
		switch r := r.(type) {
		case *string:
			s, err := v.CheckString()
			if err != nil {
				return err
			}
			*r = s
				case *int, *int64, *int32, *int16, *int8:
					s, err := v.Int64()
					if err != nil {
						return err
					}
					switch r := r.(type) {
					case *int:
						*r = int(s)
					case *int64:
						*r = s
					case *int32:
						*r = int32(s)
					case *int16:
						*r = int16(s)
					case *int8:
						*r = int8(s)
					}
		}
	*/
	rv := reflect.ValueOf(r)
	return v.unmarshal(rv)
}

func (v *Value) unmarshal(rv reflect.Value) error {
	for rv.Kind() == reflect.Ptr {
		ok, err := v.IsNull()
		isNil := err == nil && ok

		if isNil && rv.IsNil() {
			return nil
		}

		if isNil {
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
		return v.unmarshalStruct(rv)
	case reflect.String:
		q, err := v.CheckString()
		if err != nil {
			return err
		}
		rv.SetString(q)
	case reflect.Slice:

		return v.unmarshalArray(rv)
	case reflect.Array:
		return v.unmarshalArray(rv)
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		q, err := v.Int64()
		if err != nil {
			return err
		}
		rv.SetInt(q)
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		q, err := v.Uint64()
		if err != nil {
			return err
		}
		rv.SetUint(q)
	}
	return nil
}

func (v *Value) unmarshalStruct(rv reflect.Value) error {
	m := getStructMap(rv.Type())
	vis := make([]bool, rv.NumField())

	ptr := rv.UnsafeAddr()

	i, err := v.ObjectIter()
	if err != nil {
		return err
	}

	for i.HasNext() {
		k, val := i.Next()

		f, ok := m.m[string(k.MustCheckBytes())]
		if !ok {
			continue
		}

		vis[f.I] = true

		if f.FastPath {
			switch f.Kind {
			case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
				q, err := val.Int64()
				if err != nil {
					return err
				}
				switch f.Kind {
				case reflect.Int:
					*(*int)(unsafe.Pointer(ptr + f.Ptr)) = int(q)
				case reflect.Int64:
					*(*int64)(unsafe.Pointer(ptr + f.Ptr)) = q
				case reflect.Int32:
					*(*int32)(unsafe.Pointer(ptr + f.Ptr)) = int32(q)
				case reflect.Int16:
					*(*int16)(unsafe.Pointer(ptr + f.Ptr)) = int16(q)
				case reflect.Int8:
					*(*int8)(unsafe.Pointer(ptr + f.Ptr)) = int8(q)
				}
			case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
				q, err := v.Uint64()
				if err != nil {
					return err
				}
				switch f.Kind {
				case reflect.Uint:
					*(*uint)(unsafe.Pointer(ptr + f.Ptr)) = uint(q)
				case reflect.Uint64:
					*(*uint64)(unsafe.Pointer(ptr + f.Ptr)) = q
				case reflect.Uint32:
					*(*uint32)(unsafe.Pointer(ptr + f.Ptr)) = uint32(q)
				case reflect.Uint16:
					*(*uint16)(unsafe.Pointer(ptr + f.Ptr)) = uint16(q)
				case reflect.Uint8:
					*(*uint8)(unsafe.Pointer(ptr + f.Ptr)) = uint8(q)
				}
			case reflect.Float64, reflect.Float32:
				q, err := v.Float64()
				if err != nil {
					return err
				}
				if f.Kind == reflect.Float64 {
					*(*float64)(unsafe.Pointer(ptr + f.Ptr)) = float64(q)
				} else {
					*(*float32)(unsafe.Pointer(ptr + f.Ptr)) = float32(q)
				}
			case reflect.Bool:
				q, err := v.Bool()
				if err != nil {
					return err
				}
				*(*bool)(unsafe.Pointer(ptr + f.Ptr)) = q
			}
			continue
		}

		err = val.unmarshal(rv.Field(f.I))
		if err != nil {
			return err
		}
	}

	for i, vis := range vis {
		if vis {
			continue
		}
		if f := m.s[i]; f.FastErase {
			switch f.Kind {
			case reflect.Ptr:
				*(*uintptr)(unsafe.Pointer(ptr + f.Ptr)) = 0
			case reflect.Slice:
				sl := (*sliceHeader)(unsafe.Pointer(ptr + f.Ptr))
				sl.Data = 0
				sl.Len = 0
				sl.Cap = 0
			case reflect.Int:
				*(*int)(unsafe.Pointer(ptr + f.Ptr)) = 0
			case reflect.Int64:
				*(*int64)(unsafe.Pointer(ptr + f.Ptr)) = 0
			case reflect.Int32:
				*(*int32)(unsafe.Pointer(ptr + f.Ptr)) = 0
			case reflect.Int16:
				*(*int16)(unsafe.Pointer(ptr + f.Ptr)) = 0
			case reflect.Int8:
				*(*int8)(unsafe.Pointer(ptr + f.Ptr)) = 0
			case reflect.Uint:
				*(*uint)(unsafe.Pointer(ptr + f.Ptr)) = 0
			case reflect.Uint64:
				*(*uint64)(unsafe.Pointer(ptr + f.Ptr)) = 0
			case reflect.Uint32:
				*(*uint32)(unsafe.Pointer(ptr + f.Ptr)) = 0
			case reflect.Uint16:
				*(*uint16)(unsafe.Pointer(ptr + f.Ptr)) = 0
			case reflect.Uint8:
				*(*uint8)(unsafe.Pointer(ptr + f.Ptr)) = 0
			case reflect.Float64:
				*(*float64)(unsafe.Pointer(ptr + f.Ptr)) = 0
			case reflect.Float32:
				*(*float32)(unsafe.Pointer(ptr + f.Ptr)) = 0
			case reflect.Bool:
				*(*bool)(unsafe.Pointer(ptr + f.Ptr)) = false
			}
		} else {
			rv.Field(i).Set(reflect.Zero(rv.Field(i).Type()))
		}
	}

	return nil
}

func (v *Value) unmarshalArray(rv reflect.Value) error {
	elt := rv.Type().Elem()

	tp, err := v.Type()
	if err != nil {
		return err
	}

	if elt.Kind() == reflect.Uint8 && tp == String {
		bs := v.Bytes()
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

	i, err := v.ArrayIter()
	if err != nil {
		return err
	}

	j := 0
	for i.HasNext() {
		n := i.Next()

		if j < rv.Len() {
			err = n.unmarshal(rv.Index(j))
		} else if j < rv.Cap() {
			rv.Set(rv.Slice(0, j+1))
			err = n.unmarshal(rv.Index(j))
		} else {
			rv.Set(reflect.Append(rv, zero))
			err = n.unmarshal(rv.Index(j))
		}
		if err != nil {
			return err
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

type StructMap struct {
	s []StructField
	m map[string]StructField
}

type StructField struct {
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
	structMaps = map[reflect.Type]*StructMap{}
)

func getStructMap(t reflect.Type) *StructMap {
	mapMu.Lock()
	m, ok := structMaps[t]
	if ok {
		mapMu.Unlock()
		return m
	}
	defer mapMu.Unlock()

	m = &StructMap{
		m: make(map[string]StructField),
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		sf := StructField{
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
			m.m[n] = sf
			m.m[strings.ToLower(n)] = sf
		}
	}

	structMaps[t] = m

	return m
}
