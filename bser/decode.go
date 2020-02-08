package bser

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"reflect"
)

// NewDecoder returns an initialized Decoder
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r}
}

// A Decoder reads and decodes BSER values from an input stream
type Decoder struct {
	r io.Reader
}

// Decode reads the next BSER-encoded value from its
// input and stores it in the value pointed to by dest.
func (d *Decoder) Decode(dest interface{}) error {
	buf := make([]byte, 2)
	if _, err := d.r.Read(buf); err != nil {
		return err
	}

	if !bytes.Equal(buf, protocolPrefix) {
		return fmt.Errorf("Expected %x, found %x", protocolPrefix, buf)
	}

	var size int
	if err := decodeValue(d.r, reflect.ValueOf(&size)); err != nil {
		return err
	}

	buf = make([]byte, size)
	if _, err := d.r.Read(buf); err != nil {
		return err
	}

	return decodeValue(bytes.NewReader(buf), reflect.ValueOf(dest))
}

func decodeValue(r io.Reader, dest reflect.Value) error {
	if dest.Kind() != reflect.Ptr {
		return fmt.Errorf("Invalid dest passed in. Expected ptr, found: %s", dest.Kind())
	}

	dest = dest.Elem()

	buf := make([]byte, 1)
	if _, err := r.Read(buf); err != nil {
		return err
	}

	switch buf[0] {
	case 0x00:
		return decodeArray(r, dest)
	case 0x01:
		return decodeObject(r, dest)
	case 0x02:
		return decodeString(r, dest)
	case 0x03:
		return decodeInt8(r, dest)
	case 0x04:
		return decodeInt16(r, dest)
	case 0x05:
		return decodeInt32(r, dest)
	case 0x06:
		return decodeInt64(r, dest)
	case 0x07:
		return decodeReal(r, dest)
	case 0x08:
		return decodeBool(true, dest)
	case 0x09:
		return decodeBool(false, dest)
	case 0x0a:
		dest.Set(reflect.ValueOf(nil))
	case 0x0b:
		return decodeTemplate(r, dest)
	default:
		return fmt.Errorf("Invalid type marker found: %x", buf[0])
	}
	return nil
}

func decodeArray(r io.Reader, dest reflect.Value) error {
	dest, err := prep(dest, typGenericSlice)
	if err != nil {
		return err
	}

	if dest.Kind() != reflect.Slice {
		return fmt.Errorf("Expected slice, found %s", dest.Kind())
	}

	var length int
	if err := decodeValue(r, reflect.ValueOf(&length)); err != nil {
		return err
	}

	dest.Set(reflect.MakeSlice(dest.Type(), length, length))
	for i := 0; i < length; i++ {
		v := reflect.New(dest.Type().Elem())
		if err := decodeValue(r, v); err != nil {
			return err
		}

		dest.Index(i).Set(v.Elem())
	}

	return nil
}

func decodeObject(r io.Reader, dest reflect.Value) error {
	dest, err := prep(dest, typGenericMap)
	if err != nil {
		return err
	}
	switch k := dest.Kind(); k {
	case reflect.Map:
		if dest.Type().Key().Kind() != reflect.String {
			return fmt.Errorf("Map must have key type of string, found: %s", dest.Type().Key())
		}

		var fields int
		if err := decodeValue(r, reflect.ValueOf(&fields)); err != nil {
			return err
		}

		for i := 0; i < fields; i++ {
			var field string
			fieldV := reflect.ValueOf(&field)
			if err := decodeValue(r, fieldV); err != nil {
				return err
			}

			v := reflect.New(dest.Type().Elem())
			if err := decodeValue(r, v); err != nil {
				return err
			}
			dest.SetMapIndex(fieldV.Elem(), v.Elem())
		}
	case reflect.Struct:
		tfields := fields(dest.Type())
		var fields int
		if err := decodeValue(r, reflect.ValueOf(&fields)); err != nil {
			return err
		}

		for i := 0; i < fields; i++ {
			var field string
			if err := decodeValue(r, reflect.ValueOf(&field)); err != nil {
				return err
			}

			f, ok := tfields.field(field)
			if !ok {
				return fmt.Errorf("Field %s not found", field)
			}

			if err := decodeValue(r, dest.FieldByIndex(f.Index).Addr()); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("Expected struct or map, found %s", k)
	}
	return nil
}

func decodeString(r io.Reader, dest reflect.Value) error {
	dest, err := prep(dest, typString)
	if err != nil {
		return err
	}

	var size int
	if err := decodeValue(r, reflect.ValueOf(&size)); err != nil {
		return err
	}

	buf := make([]byte, size)
	if _, err := r.Read(buf); err != nil {
		return err
	}

	dest.SetString(string(buf))
	return nil
}

func decodeInt8(r io.Reader, dest reflect.Value) error {
	dest, err := prep(dest, typInt8)
	if err != nil {
		return err
	}

	buf := make([]byte, 1)
	if _, err := r.Read(buf); err != nil {
		return err
	}

	dest.SetInt(int64(buf[0]))
	return nil
}

func decodeInt16(r io.Reader, dest reflect.Value) error {
	dest, err := prep(dest, typInt16)
	if err != nil {
		return err
	}

	buf := make([]byte, 2)
	if _, err := r.Read(buf); err != nil {
		return err
	}

	v := order.Uint16(buf)
	dest.SetInt(int64(v))
	return nil
}

func decodeInt32(r io.Reader, dest reflect.Value) error {
	dest, err := prep(dest, typInt32)
	if err != nil {
		return err
	}

	buf := make([]byte, 4)
	if _, err := r.Read(buf); err != nil {
		return err
	}

	v := order.Uint32(buf)
	dest.SetInt(int64(v))
	return nil
}

func decodeInt64(r io.Reader, dest reflect.Value) error {
	dest, err := prep(dest, typInt64)
	if err != nil {
		return err
	}

	buf := make([]byte, 8)
	if _, err := r.Read(buf); err != nil {
		return err
	}

	v := order.Uint64(buf)
	dest.SetInt(int64(v))
	return nil
}

func decodeReal(r io.Reader, dest reflect.Value) error {
	dest, err := prep(dest, typFloat64)
	if err != nil {
		return err
	}

	buf := make([]byte, 8)
	if _, err := r.Read(buf); err != nil {
		return err
	}

	v := order.Uint64(buf)
	dest.SetFloat(math.Float64frombits(v))
	return nil
}

func decodeBool(v bool, dest reflect.Value) error {
	dest, err := prep(dest, typBool)
	if err != nil {
		return err
	}

	dest.SetBool(v)
	return nil
}

func decodeTemplate(r io.Reader, dest reflect.Value) error {
	dest, err := prep(dest, typGenericSlice)
	if err != nil {
		return err
	}

	if dest.Kind() != reflect.Slice {
		return fmt.Errorf("Expected slice, found %s", dest.Kind())
	}

	var fieldNames []string
	if err := decodeValue(r, reflect.ValueOf(&fieldNames)); err != nil {
		return err
	}

	var length int
	if err := decodeValue(r, reflect.ValueOf(&length)); err != nil {
		return err
	}

	dest.Set(reflect.MakeSlice(dest.Type(), length, length))
	var sfields structFields
	for i := 0; i < length; i++ {
		item := dest.Index(i)
		item, err = prep(item, typGenericMap)
		if err != nil {
			return err
		}

		if item.Kind() != reflect.Map && item.Kind() != reflect.Struct {
			return fmt.Errorf("Expected slice of struct or slice of map, found slice of %s", item.Kind())
		}

		for _, field := range fieldNames {
			switch item.Kind() {
			case reflect.Map:
				v := reflect.New(item.Type().Elem())
				if err := decodeValue(r, v); err != nil {
					return err
				}
				item.SetMapIndex(reflect.ValueOf(field), v.Elem())
			case reflect.Struct:
				if sfields == nil {
					sfields = fields(item.Type())
				}

				f, ok := sfields.field(field)
				if !ok {
					return fmt.Errorf("Field %s not found", field)
				}

				if err := decodeValue(r, item.FieldByIndex(f.Index).Addr()); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func prep(v reflect.Value, typ reflect.Type) (reflect.Value, error) {
	if v.Kind() != reflect.Interface {
		return v, nil
	}

	if v.Type().NumMethod() == 0 {
		val := reflect.New(typ)
		switch typ.Kind() {
		case reflect.Slice:
			val.Elem().Set(reflect.MakeSlice(typ, 0, 0))
		case reflect.Map:
			val.Elem().Set(reflect.MakeMap(typ))
		}

		v.Set(val)
		return val.Elem(), nil
	}

	return reflect.Value{}, fmt.Errorf("Interface found, but expected %s", typ)
}
