package json

import (
	"encoding/base64"
	"reflect"
	"strconv"
	"sync"
	"unsafe"
)

type (
	UnmarshalerFrom interface {
		UnmarshalJSONFrom(r *Reader) (err error)
	}

	unmarshalerReader = func(r *Reader, tp, p unsafe.Pointer) error
)

var (
	mur  sync.Mutex
	unrs = map[unsafe.Pointer]unmarshalerReader{}

	unrFrom = tpElem(vtp((*UnmarshalerFrom)(nil)))
)

func (r *Reader) Unmarshal(v interface{}) error {
	tp, p := unpack(v)

	un, err := r.unmarshaler(tp)
	if err != nil {
		return errWrap(err, "get decoder")
	}

	return un(r, tp, p)
}

func (r *Reader) unmarshaler(tp unsafe.Pointer) (un unmarshalerReader, err error) {
	defer mur.Unlock()
	mur.Lock()

	un, err = r.compile(tp)
	if err != nil {
		return nil, err
	}

	return un, nil
}

func (r *Reader) compile(tp unsafe.Pointer) (un unmarshalerReader, err error) {
	//	log.Printf("compile %14v [%10x]", tpString(tp), tp)

	if un, ok := unrs[tp]; ok {
		return un, nil
	}

	unrs[tp] = nil
	defer func(tp unsafe.Pointer) {
		unrs[tp] = un

		if err != nil {
			delete(unrs, tp)
		}
	}(tp)

	if un := r.unCustom(tp); un != nil {
		return un, nil
	}

	switch tpKind(tp) {
	case reflect.Pointer:
		tp := tpElem(tp)

		_, err := r.compile(tp)
		if err != nil {
			return nil, err
		}

		return unrPtr, nil

	case reflect.Array:
		tp := tpElem(tp)

		_, err := r.compile(tp)
		if err != nil {
			return nil, err
		}

		return unrArr, nil

	case reflect.Slice:
		if tpElem(tp) == vtp(byte(0)) {
			return unrBytes, nil
		}

		tp := tpElem(tp)

		_, err := r.compile(tp)
		if err != nil {
			return nil, err
		}

		return unrSlice, nil

	case reflect.Struct:
		p, err := compileStruct(tp, nil, r)
		if err != nil {
			return nil, err
		}

		return p.unmarshalReader, nil

	case reflect.Int:
		return unrInt[int], nil
	case reflect.Int64:
		return unrInt[int64], nil
	case reflect.Int32:
		return unrInt[int32], nil
	case reflect.Int16:
		return unrInt[int16], nil
	case reflect.Int8:
		return unrInt[int8], nil

	case reflect.Uint:
		return unrUint[uint], nil
	case reflect.Uint64:
		return unrUint[uint64], nil
	case reflect.Uint32:
		return unrUint[uint32], nil
	case reflect.Uint16:
		return unrUint[uint16], nil
	case reflect.Uint8:
		return unrUint[uint8], nil

	case reflect.Float32:
		return unrFloat[float32], nil
	case reflect.Float64:
		return unrFloat[float64], nil

	case reflect.Bool:
		return unrBool, nil

	case reflect.String:
		return unrString, nil

	default:
		return nil, errNew("unsupported type: %v", tpString(tp))
	}
}

func (r *Reader) unCustom(tp unsafe.Pointer) unmarshalerReader {
	switch {
	case toType(tp).Implements(toType(unrFrom)):
		return unrUnmarshalerFrom
	case toType(tp).Implements(toType(unStd)):
		return unrJSONUnmarshaler
	}

	return nil
}

func (r *Reader) un(tp unsafe.Pointer) unmarshalerReader {
	defer mur.Unlock()
	mur.Lock()

	return unrs[tp]
}

func unrInt[T anyInt](r *Reader, tp, p unsafe.Pointer) (err error) {
	jtp, err := r.Type()
	if err != nil {
		return err
	}
	if jtp != Number {
		return ErrType
	}

	raw, err := r.Raw()
	if err != nil {
		return err
	}

	//	undbg[T](Number, tp, p, raw)

	x, err := strconv.ParseInt(string(raw), 10, 8*int(unsafe.Sizeof(T(0))))
	if err != nil {
		return err
	}

	*(*T)(p) = T(x)

	return nil
}

func unrUint[T anyUint](r *Reader, tp, p unsafe.Pointer) (err error) {
	jtp, err := r.Type()
	if err != nil {
		return err
	}
	if jtp != Number {
		return ErrType
	}

	raw, err := r.Raw()
	if err != nil {
		return err
	}

	//	undbg[T](Number, tp, p, raw)

	x, err := strconv.ParseUint(string(raw), 10, 8*int(unsafe.Sizeof(T(0))))
	if err != nil {
		return err
	}

	*(*T)(p) = T(x)

	return nil
}

func unrFloat[T anyFloat](r *Reader, tp, p unsafe.Pointer) (err error) {
	jtp, err := r.Type()
	if err != nil {
		return err
	}
	if jtp != Number {
		return ErrType
	}

	raw, err := r.Raw()
	if err != nil {
		return err
	}

	x, err := strconv.ParseFloat(string(raw), 8*int(unsafe.Sizeof(T(0))))
	if err != nil {
		return err
	}

	*(*T)(p) = T(x)

	return nil
}

func unrBool(r *Reader, tp, p unsafe.Pointer) (err error) {
	jtp, err := r.Type()
	if err != nil {
		return err
	}
	if jtp != Bool {
		return ErrType
	}

	raw, err := r.Raw()
	if err != nil {
		return err
	}

	x, err := strconv.ParseBool(string(raw))
	if err != nil {
		return err
	}

	*(*bool)(p) = x

	return nil
}

func unrString(r *Reader, tp, p unsafe.Pointer) (err error) {
	x, err := r.DecodeString(nil)
	if err != nil {
		return
	}

	//	undbg[string](String, tp, p, x)

	*(*string)(p) = string(x)

	return nil
}

func unrBytes(r *Reader, tp, p unsafe.Pointer) (err error) {
	x, err := r.DecodeString(nil)
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
		return err
	}

	*(*[]byte)(p) = dec[:n]

	return nil
}

func unrPtr(r *Reader, t, p unsafe.Pointer) (err error) {
	jtp, err := r.Type()
	if err != nil {
		return err
	}

	tp := tpElem(t)
	isptr := tpKind(tp) == reflect.Pointer

	pp := (*unsafe.Pointer)(p)

	if jtp == Null {
		if isptr {
			*pp = nil
		}

		err = r.Skip()
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

	un := r.un(tp)

	return un(r, tp, p2)
}

func unrArr(r *Reader, t, p unsafe.Pointer) (err error) {
	tp := tpElem(t)
	ptp := tpPtrTo(tp)
	size := tpSize(tp)

	l := arrLen(t)

	err = r.Enter(Array)
	if err != nil {
		return
	}

	for j := 0; err == nil && r.ForMore(Array, &err); j++ {
		fp := unsafe.Add(p, uintptr(j)*size)

		if j >= l {
			typedmemclr(tp, fp)
			err = r.Skip()
			continue
		}

		err = unrPtr(r, ptp, fp)
	}

	return
}

func unrSlice(r *Reader, t, p unsafe.Pointer) (err error) {
	tp := tpElem(t)
	ptp := tpPtrTo(tp)
	size := tpSize(tp)

	s := *(*sliceHeader)(p)

	err = r.Enter(Array)
	if err != nil {
		return
	}

	j := 0

	for err == nil && r.ForMore(Array, &err) {
		if j == s.l {
			s = growslice(tp, s, 1)
		}

		fp := unsafe.Add(s.p, uintptr(j)*size)
		err = unrPtr(r, ptp, fp)

		s.l++
		j++
	}

	s.l = j

	*(*sliceHeader)(p) = s

	return
}

func unrUnmarshalerFrom(r *Reader, t, p unsafe.Pointer) (err error) {
	return pack(t, p).(UnmarshalerFrom).UnmarshalJSONFrom(r)
}

func unrJSONUnmarshaler(r *Reader, t, p unsafe.Pointer) (err error) {
	raw, err := r.Raw()
	if err != nil {
		return err
	}

	err = pack(t, p).(Unmarshaler).UnmarshalJSON(raw)
	if err != nil {
		return err
	}

	return nil
}

func (pr *structProg) unmarshalReader(r *Reader, tp, p unsafe.Pointer) (err error) {
	err = r.Enter(Object)
	if err != nil {
		return
	}

	var k []byte

	for err == nil && r.ForMore(Object, &err) {
		k, err = r.Key()
		if err != nil {
			return
		}

		f, ok := pr.dec[string(k)]
		if !ok {
			err = r.Skip()
			continue
		}

		fp := unsafe.Add(p, f.off)

		//	log.Printf("field   %14v %10x    -> %10x is %10x + %4x  name %s", tpString(f.tp), f.tp, fp, p, f.off, k)

		if f.unr != nil {
			err = f.unr(r, f.ptp, fp)
		} else {
			err = unrPtr(r, f.ptp, fp)
		}
	}

	if err != nil {
		return
	}

	return nil
}

func (rm *RawMessage) UnmarshalJSONFrom(r *Reader) (err error) {
	raw, err := r.Raw()
	if err != nil {
		return err
	}

	*rm = (*rm)[:0]
	*rm = append(*rm, raw...)

	return nil
}
