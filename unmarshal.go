package json

import (
	"encoding/json"
	"reflect"
	"strings"
)

func (v *Value) Unmarshal(r interface{}) error {
	if u, ok := r.(json.Unmarshaler); ok {
		return u.UnmarshalJSON(v.buf)
	}
	rv := reflect.ValueOf(r)
	return v.unmarshal(rv)
}

func (v *Value) unmarshal(rv reflect.Value) error {
	for rv.Kind() == reflect.Ptr {
		ok, err := v.IsNull()
		if err == nil && ok {
			rv.Set(reflect.Zero(rv.Type()))
			return nil
		}

		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Uint,
		reflect.Uint64,
		reflect.Uint32,
		reflect.Uint16,
		reflect.Uint8:
		r, err := v.Uint64()
		if err != nil {
			return err
		}
		rv.SetUint(r)
	case reflect.Int,
		reflect.Int64,
		reflect.Int32,
		reflect.Int16,
		reflect.Int8:
		r, err := v.Int64()
		if err != nil {
			return err
		}
		rv.SetInt(int64(r))
	case reflect.String:
		r, err := v.CheckString()
		if err != nil {
			return err
		}
		rv.SetString(r)
	case reflect.Slice:
		elt := rv.Type().Elem()

		res := reflect.MakeSlice(rv.Type(), 0, 0)

		i, err := v.ArrayIter()
		if err != nil {
			return err
		}
		for i.HasNext() {
			n := i.Next()
			re := reflect.New(elt)
			err := n.unmarshal(re)
			if err != nil {
				return err
			}
			res = reflect.Append(res, re.Elem())
		}

		if res.Len() != 0 {
			rv.Set(res)
		} else {
			rv.Set(reflect.Zero(rv.Type()))
		}

	case reflect.Array:
		elt := rv.Type().Elem()

		j := 0
		i, err := v.ArrayIter()
		if err != nil {
			return err
		}
		for i.HasNext() {
			n := i.Next()
			err := n.unmarshal(rv.Index(j))
			if err != nil {
				return err
			}
			j++
		}

		for j < rv.Len() {
			rv.Index(j).Set(reflect.Zero(elt))
			j++
		}

	case reflect.Struct:
		tp := rv.Type()
		rv.Set(reflect.Zero(rv.Type()))

		i, err := v.ObjectIter()
		if err != nil {
			return err
		}

		for i.HasNext() {
			k, val := i.Next()
			kk := k.MustCheckString()

			ft, ok := tp.FieldByName(kk)
			if !ok {
				ft, ok = tp.FieldByName(strings.Title(kk))
			}
			if ok {
				fv := rv.Field(ft.Index[0])
				err = val.unmarshal(fv)
				if err != nil {
					return err
				}
			} else {
				panic("slow path needed")
			}
		}

	default:
		panic(rv)
	}
	return nil
}
