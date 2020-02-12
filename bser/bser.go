package bser

import (
	"encoding/binary"
	"errors"
	"reflect"
	"strings"
)

var (
	protocolPrefix  = []byte{0, 1}
	order           = binary.LittleEndian
	typString       = reflect.TypeOf("")
	typInt          = reflect.TypeOf(int(0))
	typInt8         = reflect.TypeOf(int8(0))
	typInt16        = reflect.TypeOf(int16(0))
	typInt32        = reflect.TypeOf(int32(0))
	typInt64        = reflect.TypeOf(int64(0))
	typFloat32      = reflect.TypeOf(float32(0))
	typFloat64      = reflect.TypeOf(float64(0))
	typGenericSlice = reflect.TypeOf([]interface{}{})
	typGenericMap   = reflect.TypeOf(map[string]interface{}{})
	typBool         = reflect.TypeOf(true)
	typMarshaler    = reflect.TypeOf((*Marshaler)(nil)).Elem()
	typUnmarshaler  = reflect.TypeOf((*Unmarshaler)(nil)).Elem()
)

type structFields map[string]field

func (s structFields) field(name string) (field, bool) {
	if f, ok := s[name]; ok {
		return f, true
	}

	for k, f := range s {
		if strings.EqualFold(k, name) {
			return f, true
		}
	}

	return field{}, false
}

type field struct {
	Name  string
	Index []int
}

func fields(t reflect.Type) structFields {
	type ftyp struct {
		typ   reflect.Type
		index []int
	}

	fields := structFields{}
	typs := []ftyp{{t, []int{}}}

	for len(typs) > 0 {
		typ := typs[0]
		typs = typs[1:]

		for i := 0; i < typ.typ.NumField(); i++ {
			f := typ.typ.Field(i)
			unexp := f.PkgPath != ""
			if unexp && !f.Anonymous {
				continue
			}

			tag := f.Tag.Get("bser")
			if tag == "-" {
				continue
			}

			name := tag
			if name == "" {
				name = f.Name
			}

			idx := append(append([]int(nil), typ.index...), i)

			fields[name] = field{
				Name:  name,
				Index: idx,
			}

			if f.Anonymous {
				typs = append(typs, ftyp{f.Type, idx})
			}
		}
	}

	return fields
}

// Marshaler allows a type to define a custom marshal mechanism
type Marshaler interface {
	MarshalBSER() ([]byte, error)
}

// Unmarshaler allows a type to define a custom unmarshal mechanism
type Unmarshaler interface {
	UnmarshalBSER([]byte) error
}

// RawMessage is a raw encoded BSER value
type RawMessage []byte

// MarshalBSER implements Marshaler
func (r RawMessage) MarshalBSER() ([]byte, error) {
	if r == nil {
		return []byte{0x0a}, nil
	}
	return r, nil
}

// UnmarshalBSER implements Unmarshaler
func (r *RawMessage) UnmarshalBSER(b []byte) error {
	if r == nil {
		return errors.New("bser.RawMessage: UnmarshalBSER on nil pointer")
	}

	*r = append((*r)[0:0], b...)
	return nil
}
