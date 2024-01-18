package json

import (
	"log"
	"strings"
	"unsafe"
)

type (
	structProg struct {
		enc []*structField
		dec map[string]*structField
	}

	structField struct {
		tp unsafe.Pointer

		off uintptr
	}
)

var (
	// protected by mu
	progs = map[unsafe.Pointer]*structProg{}
)

func (d *Decoder) compileStruct(tp unsafe.Pointer) (unmarshaler, error) {
	if p, ok := progs[tp]; ok {
		return p.unmarshal, nil
	}

	p := &structProg{
		dec: map[string]*structField{},
	}

	err := d.compileStructFields(tp, p)
	if err != nil {
		return nil, err
	}

	progs[tp] = p

	return p.unmarshal, nil
}

func (d *Decoder) compileStructFields(tp unsafe.Pointer, p *structProg) error {
	r := toType(tp)

	for i := 0; i < r.NumField(); i++ {
		sf := r.Field(i)

		tag := strings.Split(sf.Tag.Get("json"), ",")

		if len(tag) == 0 || tag[0] == "-" {
			continue
		}

		_, tp := unpack(sf.Type)
		//tp = tpPtrTo(tp)

		_, err := d.compile(tpPtrTo(tp))
		if err != nil {
			return err
		}

		f := &structField{
			tp: tp,

			off: sf.Offset,
		}

		p.enc = append(p.enc, f)
		p.dec[tag[0]] = f
	}

	return nil
}

func (pr *structProg) unmarshal(d *Decoder, b []byte, st int, tp, p unsafe.Pointer) (i int, err error) {
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

		f, ok := pr.dec[string(k)]
		if !ok {
			i, err = d.Skip(b, i)
			continue
		}

		tp := f.tp
		tp = tpPtrTo(tp)
		un := d.un(tp)

		fp := unsafe.Add(p, f.off)

		log.Printf("field   %14v %10x    -> %10x is %10x + %4x  name %s", tpString(f.tp), f.tp, fp, p, f.off, k)

		i, err = un(d, b, i, tp, fp)
	}

	if err != nil {
		return
	}

	return i, nil
}
