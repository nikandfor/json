package json

import (
	"strings"
	"unsafe"
)

type (
	structProg struct {
		enc []*structField
		dec map[string]*structField
	}

	structField struct {
		ptp unsafe.Pointer

		off uintptr

		un  unmarshaler
		unr unmarshalerReader
	}
)

var (
	// protected by mu
	progs = map[unsafe.Pointer]*structProg{}
)

func compileStruct(tp unsafe.Pointer, d *Decoder, r *Reader) (*structProg, error) {
	if pr, ok := progs[tp]; ok {
		return pr, nil
	}

	pr := &structProg{
		dec: map[string]*structField{},
	}

	err := compileStructFields(tp, pr, d, r)
	if err != nil {
		return nil, err
	}

	progs[tp] = pr

	return pr, nil
}

func compileStructFields(tp unsafe.Pointer, p *structProg, d *Decoder, r *Reader) error {
	rt := toType(tp)

	for i := 0; i < rt.NumField(); i++ {
		sf := rt.Field(i)

		tag := strings.Split(sf.Tag.Get("json"), ",")

		if len(tag) == 0 || tag[0] == "-" {
			continue
		}

		_, tp := unpack(sf.Type)

		if sf.Anonymous {
			err := compileStructFields(tp, p, d, r)
			if err != nil {
				return err
			}

			continue
		}

		f := &structField{
			ptp: tpPtrTo(tp),

			off: sf.Offset,
		}

		if un := d.unCustom(f.ptp); un != nil {
			f.un = un
		}

		if unr := r.unCustom(f.ptp); unr != nil {
			f.unr = unr
		}

		if f.un == nil {
			_, err := d.compile(tp)
			if err != nil {
				return err
			}
		}

		if f.unr == nil {
			_, err := r.compile(tp)
			if err != nil {
				return err
			}
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

		fp := unsafe.Add(p, f.off)

		//	log.Printf("field   %14v %10x    -> %10x is %10x + %4x  name %s", tpString(f.tp), f.tp, fp, p, f.off, k)

		if f.un != nil {
			i, err = f.un(d, b, i, f.ptp, fp)
		} else {
			i, err = unPtr(d, b, i, f.ptp, fp)
		}
	}

	if err != nil {
		return
	}

	return i, nil
}
