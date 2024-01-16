package json

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"unsafe"
)

type (
	Unmarshaler interface {
		UnmarshalJSON(b []byte, st int) (i int, err error)
	}

	unmarshaler = func(d *Decoder, b []byte, st int, x interface{}) (int, error)

	anyInt interface {
		~int | ~int64 | ~int32 | ~int16 | ~int8
	}

	anyUint interface {
		~uint | ~uint64 | ~uint32 | ~uint16 | ~uint8
	}

	anyFloat interface {
		~float32 | ~float64
	}
)

var (
	mu    sync.Mutex
	uns   = map[unsafe.Pointer]unmarshaler{}
	progs = map[reflect.Type]*structProg{}
)

func (d *Decoder) Unmarshal(b []byte, st int, v interface{}) (int, error) {
	un, err := d.unmarshaler(v)
	if err != nil {
		return st, errWrap(err, "get decoder")
	}

	return un(d, b, st, v)
}

func (d *Decoder) un(tp unsafe.Pointer) unmarshaler {
	defer mu.Unlock()
	mu.Lock()

	return uns[tp]
}

func (d *Decoder) unmarshaler(v interface{}) (unmarshaler, error) {
	tp, _ := unpack(v)

	defer mu.Unlock()
	mu.Lock()

	un := uns[tp]
	if un != nil {
		return un, nil
	}

	r := reflect.TypeOf(v)

	if r.Kind() != reflect.Pointer {
		return nil, errors.New("expected pointer to a value")
	}

	return d.compile(r)
}

func (d *Decoder) compile(r reflect.Type) (un unmarshaler, err error) {
	orig := r

	_, tp := unpack(r)
	//	log.Printf("compile %10v  ptr %5v  indir %5v", r, r.Kind() == reflect.Pointer, ifaceIndir(tp))

	uns[tp] = nil // break recursion circle

	defer func() {
		uns[tp] = un
	}()

	if r.Kind() == reflect.Pointer {
		r = r.Elem()

		un, err := d.compile(r)
		if err != nil {
			return nil, err
		}

		if r.Kind() != reflect.Pointer {
			return un, nil
		}

		return unPtr(tp, un), nil
	}

	switch r.Kind() {
	case reflect.Int:
		return unInt[int], nil
	case reflect.Int64:
		return unInt[int64], nil
	case reflect.Int32:
		return unInt[int32], nil
	case reflect.Int16:
		return unInt[int16], nil
	case reflect.Int8:
		return unInt[int8], nil

	case reflect.Uint:
		return unUint[uint], nil
	case reflect.Uint64:
		return unUint[uint64], nil
	case reflect.Uint32:
		return unUint[uint32], nil
	case reflect.Uint16:
		return unUint[uint16], nil
	case reflect.Uint8:
		return unUint[uint8], nil

	case reflect.Float32:
		return unFloat[float32], nil
	case reflect.Float64:
		return unFloat[float64], nil

	case reflect.Bool:
		return unBool, nil

	case reflect.String:
		return unString, nil

	case reflect.Struct:
		return d.compileStruct(r)

	default:
		return nil, errNew("unsupported type: %v", orig)
	}
}

func unInt[T anyInt](d *Decoder, b []byte, st int, v interface{}) (i int, err error) {
	tp, i, err := d.Type(b, st)
	if err != nil {
		return i, err
	}
	if tp != Number {
		return i, ErrType
	}

	raw, i, err := d.Raw(b, i)
	if err != nil {
		return i, err
	}

	//	{
	//		t, d := unpack(v)
	//		log.Printf("unmarshal %q to %T  %x %x", tp, T(0), t, d)
	//	}

	x, err := strconv.ParseInt(string(raw), 10, 8*int(unsafe.Sizeof(T(0))))
	if err != nil {
		return st, err
	}

	*(*T)(vptr(v)) = T(x)

	return i, nil
}

func unUint[T anyUint](d *Decoder, b []byte, st int, v interface{}) (i int, err error) {
	tp, i, err := d.Type(b, st)
	if err != nil {
		return i, err
	}
	if tp != Number {
		return i, ErrType
	}

	raw, i, err := d.Raw(b, i)
	if err != nil {
		return i, err
	}

	x, err := strconv.ParseUint(string(raw), 10, 8*int(unsafe.Sizeof(T(0))))
	if err != nil {
		return st, err
	}

	*(*T)(vptr(v)) = T(x)

	return i, nil
}

func unFloat[T anyFloat](d *Decoder, b []byte, st int, v interface{}) (i int, err error) {
	tp, i, err := d.Type(b, st)
	if err != nil {
		return i, err
	}
	if tp != Number {
		return i, ErrType
	}

	raw, i, err := d.Raw(b, i)
	if err != nil {
		return i, err
	}

	x, err := strconv.ParseFloat(string(raw), 8*int(unsafe.Sizeof(T(0))))
	if err != nil {
		return st, err
	}

	*(*T)(vptr(v)) = T(x)

	return i, nil
}

func unBool(d *Decoder, b []byte, st int, v interface{}) (i int, err error) {
	tp, i, err := d.Type(b, st)
	if err != nil {
		return i, err
	}
	if tp != Bool {
		return i, ErrType
	}

	raw, i, err := d.Raw(b, i)
	if err != nil {
		return i, err
	}

	x, err := strconv.ParseBool(string(raw))
	if err != nil {
		return st, err
	}

	*(*bool)(vptr(v)) = x

	return i, nil
}

func unString(d *Decoder, b []byte, st int, v interface{}) (i int, err error) {
	x, i, err := d.DecodeString(b, st, nil)
	if err != nil {
		return
	}

	*(*string)(vptr(v)) = string(x)

	return i, nil
}

func unPtr(tp unsafe.Pointer, un unmarshaler) unmarshaler {
	return func(d *Decoder, b []byte, st int, v interface{}) (i int, err error) {
		ptr := (*unsafe.Pointer)(vptr(v))

		if *ptr == nil {
			*ptr = unsafeNew(tp)
		}

		return un(d, b, st, pack(tp, *ptr))
	}
}

func errNew(f string, args ...interface{}) error {
	return fmt.Errorf(f, args...)
}

func errWrap(err error, f string) error {
	return fmt.Errorf("%v: %w", f, err)
}
