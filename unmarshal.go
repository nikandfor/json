package json

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"sync"
	"unsafe"
)

type (
	Unmarshaler interface {
		UnmarshalJSON(b []byte, st int) (i int, err error)
	}

	unmarshaler = func(d *Decoder, b []byte, st int, tp, p unsafe.Pointer) (int, error)

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
	mu  sync.Mutex
	uns = map[unsafe.Pointer]unmarshaler{}
)

func (d *Decoder) Unmarshal(b []byte, st int, v interface{}) (int, error) {
	tp, p := unpack(v)

	un, err := d.unmarshaler(tp)
	if err != nil {
		return st, errWrap(err, "get decoder")
	}

	return un(d, b, st, tp, p)
}

func (d *Decoder) unmarshaler(tp unsafe.Pointer) (un unmarshaler, err error) {
	defer mu.Unlock()
	mu.Lock()

	un, err = d.compile(tp)
	if err != nil {
		return nil, err
	}

	return un, nil
}

func (d *Decoder) compile(tp unsafe.Pointer) (un unmarshaler, err error) {
	log.Printf("compile %14v [%10x]", tpString(tp), tp)

	if un, ok := uns[tp]; ok {
		return un, nil
	}

	uns[tp] = nil
	defer func(tp unsafe.Pointer) {
		uns[tp] = un

		if err != nil {
			delete(uns, tp)
		}
	}(tp)

	switch tpKind(tp) {
	case reflect.Pointer:
		tp := tpElem(tp)

		_, err := d.compile(tp)
		if err != nil {
			return nil, err
		}

		return unPtr, nil

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
		return d.compileStruct(tp)

	default:
		return nil, errNew("unsupported type: %v", tpString(tp))
	}
}

func (d *Decoder) un(tp unsafe.Pointer) unmarshaler {
	defer mu.Unlock()
	mu.Lock()

	return uns[tp]
}

func unInt[T anyInt](d *Decoder, b []byte, st int, tp, p unsafe.Pointer) (i int, err error) {
	jtp, i, err := d.Type(b, st)
	if err != nil {
		return i, err
	}
	if jtp != Number {
		return i, ErrType
	}

	raw, i, err := d.Raw(b, i)
	if err != nil {
		return i, err
	}

	undbg[T](Number, tp, p)

	x, err := strconv.ParseInt(string(raw), 10, 8*int(unsafe.Sizeof(T(0))))
	if err != nil {
		return st, err
	}

	*(*T)(p) = T(x)

	return i, nil
}

func unUint[T anyUint](d *Decoder, b []byte, st int, tp, p unsafe.Pointer) (i int, err error) {
	jtp, i, err := d.Type(b, st)
	if err != nil {
		return i, err
	}
	if jtp != Number {
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

	*(*T)(p) = T(x)

	return i, nil
}

func unFloat[T anyFloat](d *Decoder, b []byte, st int, tp, p unsafe.Pointer) (i int, err error) {
	jtp, i, err := d.Type(b, st)
	if err != nil {
		return i, err
	}
	if jtp != Number {
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

	*(*T)(p) = T(x)

	return i, nil
}

func unBool(d *Decoder, b []byte, st int, tp, p unsafe.Pointer) (i int, err error) {
	jtp, i, err := d.Type(b, st)
	if err != nil {
		return i, err
	}
	if jtp != Bool {
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

	*(*bool)(p) = x

	return i, nil
}

func unString(d *Decoder, b []byte, st int, tp, p unsafe.Pointer) (i int, err error) {
	x, i, err := d.DecodeString(b, st, nil)
	if err != nil {
		return
	}

	undbg[string](String, tp, p)

	s := string(x)

	*(**byte)(p) = *(**byte)(unsafe.Pointer(&s))
	*(*int)(unsafe.Add(p, unsafe.Sizeof((*byte)(nil)))) = len(s)

	return i, nil
}

func unPtr(d *Decoder, b []byte, st int, t, p unsafe.Pointer) (i int, err error) {
	jtp, i, err := d.Type(b, st)
	if err != nil {
		return i, err
	}

	tp := tpElem(t)
	isptr := tpKind(tp) == reflect.Pointer
	//indir := ifaceIndir(tp)

	pp := (*unsafe.Pointer)(p)

	if jtp == Null {
		if isptr {
			*pp = nil
		}

		i, err = d.Skip(b, i)
		return
	}

	al := '.'
	if isptr && *pp == nil {
		al = 'a'
		*pp = unsafe_New(tpElem(tp))
	}

	p2 := p
	if isptr {
		p2 = *pp
	}

	log.Printf("unPtr   %14v %10x %s => %14v %10x %s  ptr %x -> %x : %x %c", tpString(t), t, flags(t), tpString(tp), tp, flags(tp), p, *pp, p2, al)

	un := d.un(tp)

	return un(d, b, st, tp, p2)
}

func errNew(f string, args ...interface{}) error {
	return fmt.Errorf(f, args...)
}

func errWrap(err error, f string) error {
	return fmt.Errorf("%v: %w", f, err)
}

func undbg[T any](jtp byte, tp, p unsafe.Pointer) {
	//	t, d := unpack(v)
	//	size := tpSize(tp)
	//	val := *(*unsafe.Pointer)(p)
	//	base, _, _ := findObject(p, 0, 0)
	//	endBase, _, _ := findObject(unsafe.Add(p, size-1), 0, 0)
	log.Printf("unm %c   %14v %10x    -> %10x", jtp, tpString(tp), tp, p)
	// log.Printf("unmarshal %q to %14v  %x %x  base %x %x  size %x  val %x", jtp, tpString(tp), tp, p, base, endBase, size, val)
}

func flags(tp unsafe.Pointer) string {
	isptr := tpKind(tp) == reflect.Pointer
	indir := ifaceIndir(tp)

	ptr := "v"
	if isptr {
		ptr = "p"
	}

	idr := "d"
	if indir {
		idr = "i"
	}

	return ptr + idr
}
