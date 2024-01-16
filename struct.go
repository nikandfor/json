package json

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"
)

type (
	structProg struct {
		dec map[string]*field
	}

	field struct {
		name string

		tp  unsafe.Pointer
		off uintptr

		make func(unsafe.Pointer) unsafe.Pointer

		bits fieldBits
	}

	fieldBits uint8
)

const (
	fieldPtr = 1 << iota
)

func (d *Decoder) compileStruct(r reflect.Type) (unmarshaler, error) {
	if p, ok := progs[r]; ok {
		return p.unmarshal, nil
	}

	p := &structProg{
		dec: map[string]*field{},
	}

	progs[r] = p

	err := d.compileStructFields(p, r)
	if err != nil {
		return nil, err
	}

	//	log.Printf("compiled struct %v %+v", r, p)

	return p.unmarshal, nil
}

func (d *Decoder) compileStructFields(p *structProg, r reflect.Type) error {
	for i := 0; i < r.NumField(); i++ {
		sf := r.Field(i)

		//	log.Printf("field %+v", sf)

		tag := strings.Split(sf.Tag.Get("json"), ",")

		if len(tag) != 0 && tag[0] == "-" {
			continue
		}

		if sf.Anonymous {
			err := d.compileStructFields(p, sf.Type)
			if err != nil {
				return err
			}

			continue
		}

		_, err := d.compile(sf.Type)
		if err != nil {
			return fmt.Errorf("%v: %w", sf.Name, err)
		}

		f := &field{
			name: sf.Name,
			off:  sf.Offset,
		}

		_, f.tp = unpack(sf.Type)

		if sf.Type.Kind() == reflect.Pointer {
			f.bits |= fieldPtr
			f.make = unsafeNew
		}

		if len(tag) != 0 {
			f.name = tag[0]
		}

		p.dec[f.name] = f
	}

	return nil
}

func (p *structProg) unmarshal(d *Decoder, b []byte, st int, v interface{}) (i int, err error) {
	_, ptr := unpack(v)

	//	log.Printf("unmarshal struct %15T  ptr %12x", v, ptr)

	i, err = d.Enter(b, st, Object)
	if err != nil {
		return
	}

	var k []byte

	for err == nil && d.ForMore(b, &i, Object, &err) {
		k, i, err = d.Key(b, i)
		if err != nil {
			return
		}

		f, ok := p.dec[string(k)]
		if !ok {
			i, err = d.Skip(b, i)
			continue
		}

		fptr := unsafe.Add(ptr, f.off)

		//	log.Printf("unmarshal struct %15T  ptr %12x  off %12x  bits %x  field %s", v, fptr, f.off, f.bits, k)

		if f.bits&fieldPtr != 0 {
			fptr2 := (*unsafe.Pointer)(fptr) // cast field offset to a pointer

			if *fptr2 == nil {
				*fptr2 = f.make(f.tp)
				//	log.Printf("allocated new object for field at %x -> %x", fptr2, *fptr2)
			}

			fptr = *fptr2

			//	log.Printf("unmarshal field pointer dereferenced to %x", fptr)
		}

		vf := pack(f.tp, fptr)

		un := d.un(f.tp)
		i, err = un(d, b, i, vf)
	}

	if err != nil {
		return
	}

	return i, nil
}
