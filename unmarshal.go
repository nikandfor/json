package json

import (
	"encoding/base64"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"sync"
	"unsafe"
)

type (
	UnmarshalerAt interface {
		UnmarshalJSONAt(b []byte, st int) (i int, err error)
	}

	UnmarshalerAtDecoder interface {
		UnmarshalJSONAt(d *Decoder, b []byte, st int) (i int, err error)
	}

	Unmarshaler interface {
		UnmarshalJSON(b []byte) (err error)
	}

	unmarshaler = func(d *Decoder, b []byte, st int, tp, p unsafe.Pointer) (int, error)

	RawMessage []byte

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

	unAt    = tpElem(vtp((*UnmarshalerAt)(nil)))
	unAtDec = tpElem(vtp((*UnmarshalerAtDecoder)(nil)))
	unStd   = tpElem(vtp((*Unmarshaler)(nil)))
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
	//	log.Printf("compile %14v [%10x]", tpString(tp), tp)

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

	if un := d.unCustom(tp); un != nil {
		return un, nil
	}

	switch tpKind(tp) {
	case reflect.Pointer:
		tp := tpElem(tp)

		_, err := d.compile(tp)
		if err != nil {
			return nil, err
		}

		return unPtr, nil

	case reflect.Array:
		tp := tpElem(tp)

		_, err := d.compile(tp)
		if err != nil {
			return nil, err
		}

		return unArr, nil

	case reflect.Slice:
		if tpElem(tp) == vtp(byte(0)) {
			return unBytes, nil
		}

		tp := tpElem(tp)

		_, err := d.compile(tp)
		if err != nil {
			return nil, err
		}

		return unSlice, nil

	case reflect.Struct:
		p, err := compileStruct(tp, d, nil)
		if err != nil {
			return nil, err
		}

		return p.unmarshal, nil

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

	default:
		return nil, errNew("unsupported type: %v", tpString(tp))
	}
}

func (d *Decoder) unCustom(tp unsafe.Pointer) unmarshaler {
	switch {
	case toType(tp).Implements(toType(unAt)):
		return unUnmarshalerAt
	case toType(tp).Implements(toType(unAtDec)):
		return unUnmarshalerAtDecoder
	case toType(tp).Implements(toType(unStd)):
		return unJSONUnmarshaler
	}

	return nil
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

	//	undbg[T](Number, tp, p, raw)

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

	//	undbg[string](String, tp, p, x)

	*(*string)(p) = string(x)

	return i, nil
}

func unBytes(d *Decoder, b []byte, st int, tp, p unsafe.Pointer) (i int, err error) {
	x, i, err := d.DecodeString(b, st, nil)
	if err != nil {
		return
	}

	dl := base64.StdEncoding.DecodedLen(len(x))

	dec := *(*[]byte)(p)

	if dl <= cap(dec) {
		dec = dec[:dl]
	} else {
		dec = make([]byte, dl)
	}

	n, err := base64.StdEncoding.Decode(dec, x)
	if err != nil {
		return st, err
	}

	*(*[]byte)(p) = dec[:n]

	return i, nil
}

func unPtr(d *Decoder, b []byte, st int, t, p unsafe.Pointer) (i int, err error) {
	jtp, i, err := d.Type(b, st)
	if err != nil {
		return i, err
	}

	tp := tpElem(t)
	isptr := tpKind(tp) == reflect.Pointer

	pp := (*unsafe.Pointer)(p)

	if jtp == Null {
		if isptr {
			*pp = nil
		}

		i, err = d.Skip(b, i)
		return
	}

	if isptr && *pp == nil {
		*pp = unsafe_New(tpElem(tp))
	}

	p2 := p
	if isptr {
		p2 = *pp
	}

	//	log.Printf("unPtr   %14v %10x %s => %14v %10x %s  ptr %x -> %x : %x %c", tpString(t), t, flags(t), tpString(tp), tp, flags(tp), p, *pp, p2, al)

	un := d.un(tp)

	return un(d, b, st, tp, p2)
}

func unArr(d *Decoder, b []byte, st int, t, p unsafe.Pointer) (i int, err error) {
	tp := tpElem(t)
	ptp := tpPtrTo(tp)
	size := tpSize(tp)

	l := arrLen(t)

	i, err = d.Enter(b, st, Array)
	if err != nil {
		return
	}

	for j := 0; err == nil && d.ForMore(b, &i, Array, &err); j++ {
		fp := unsafe.Add(p, uintptr(j)*size)

		if j >= l {
			typedmemclr(tp, fp)
			i, err = d.Skip(b, i)
			continue
		}

		i, err = unPtr(d, b, i, ptp, fp)
	}

	return
}

func unSlice(d *Decoder, b []byte, st int, t, p unsafe.Pointer) (i int, err error) {
	tp := tpElem(t)
	ptp := tpPtrTo(tp)
	size := tpSize(tp)

	s := *(*sliceHeader)(p)

	i, err = d.Enter(b, st, Array)
	if err != nil {
		return
	}

	j := 0

	for err == nil && d.ForMore(b, &i, Array, &err) {
		if j == s.l {
			s = growslice(tp, s, 1)
		}

		fp := unsafe.Add(s.p, uintptr(j)*size)
		i, err = unPtr(d, b, i, ptp, fp)

		s.l++
		j++
	}

	s.l = j

	*(*sliceHeader)(p) = s

	return
}

func unUnmarshalerAt(d *Decoder, b []byte, st int, t, p unsafe.Pointer) (i int, err error) {
	return pack(t, p).(UnmarshalerAt).UnmarshalJSONAt(b, st)
}

func unUnmarshalerAtDecoder(d *Decoder, b []byte, st int, t, p unsafe.Pointer) (i int, err error) {
	return pack(t, p).(UnmarshalerAtDecoder).UnmarshalJSONAt(d, b, st)
}

func unJSONUnmarshaler(d *Decoder, b []byte, st int, t, p unsafe.Pointer) (i int, err error) {
	raw, i, err := d.Raw(b, st)
	if err != nil {
		return i, err
	}

	err = pack(t, p).(Unmarshaler).UnmarshalJSON(raw)
	if err != nil {
		return st, err
	}

	return i, nil
}

func (r *RawMessage) UnmarshalJSONAt(d *Decoder, b []byte, st int) (i int, err error) {
	raw, i, err := d.Raw(b, st)
	if err != nil {
		return i, err
	}

	*r = (*r)[:0]
	*r = append(*r, raw...)

	return i, nil
}

func errNew(f string, args ...interface{}) error {
	return fmt.Errorf(f, args...)
}

func errWrap(err error, f string) error {
	return fmt.Errorf("%v: %w", f, err)
}

func undbg[T any](jtp byte, tp, p unsafe.Pointer, raw []byte) {
	log.Printf("unm %c   %14v %10x    -> %10x  (%s)", jtp, tpString(tp), tp, p, raw)
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
