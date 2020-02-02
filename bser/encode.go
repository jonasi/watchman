package bser

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"reflect"
)

// Marshal returns the BSER encoding of d
func Marshal(d interface{}) ([]byte, error) {
	var b bytes.Buffer
	if err := NewEncoder(&b).Encode(d); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

// NewEncoder returns an initialized Encoder
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

// Encoder writes and encodes BSER values to an output stream
type Encoder struct {
	w io.Writer
}

// Encode writes the value d to the output
func (e *Encoder) Encode(d interface{}) error {
	pdu, err := encode(nil, d)
	if err != nil {
		return err
	}

	header := []byte("\x00\x01")
	header, err = encode(header, len(pdu))
	if err != nil {
		return err
	}

	if _, err = e.w.Write(header); err != nil {
		return err
	}
	_, err = e.w.Write(pdu)
	return err
}

func appendItem(buf []byte, m byte, data []byte) []byte {
	buf = append(buf, m)
	if len(data) > 0 {
		buf = append(buf, data...)
	}

	return buf
}

func encode(buf []byte, d interface{}) ([]byte, error) {
	switch v := d.(type) {
	case string:
		b, err := encode(nil, len([]byte(v)))
		if err != nil {
			return nil, err
		}
		b = append(b, []byte(v)...)
		return appendItem(buf, 0x02, b), nil
	case int8:
		b := make([]byte, 1)
		b[0] = byte(v)
		return appendItem(buf, 0x03, b), nil
	case int16:
		b := make([]byte, 2)
		order.PutUint16(b, uint16(v))
		return appendItem(buf, 0x04, b), nil
	case int32:
		b := make([]byte, 4)
		order.PutUint32(b, uint32(v))
		return appendItem(buf, 0x05, b), nil
	case int64:
		b := make([]byte, 8)
		order.PutUint64(b, uint64(v))
		return appendItem(buf, 0x06, b), nil
	case int:
		return encode(buf, fitInt(v))
	case bool:
		if v {
			return appendItem(buf, 0x08, nil), nil
		}
		return appendItem(buf, 0x09, nil), nil
	default:
		r := reflect.ValueOf(d)
		if isNillable(r.Kind()) && r.IsNil() {
			return appendItem(buf, 0x0A, nil), nil
		}

		elem := false
		for r.Kind() == reflect.Ptr || r.Kind() == reflect.Interface {
			r = r.Elem()
			elem = true
		}

		if elem {
			return encode(buf, r.Interface())
		}

		switch r.Kind() {
		case reflect.Float32:
			b := make([]byte, 8)
			order.PutUint64(b, math.Float64bits(r.Float()))
			return appendItem(buf, 0x07, b), nil
		case reflect.Float64:
			b := make([]byte, 8)
			order.PutUint64(b, math.Float64bits(r.Float()))
			return appendItem(buf, 0x07, b), nil
		case reflect.Slice:
			b, err := encode(nil, r.Len())
			if err != nil {
				return nil, err
			}
			for i := 0; i < r.Len(); i++ {
				if b, err = encode(b, r.Index(i).Interface()); err != nil {
					return nil, err
				}
			}
			return appendItem(buf, 0x00, b), nil
		case reflect.Struct:
			num := r.NumField()
			b, err := encode(nil, num)
			if err != nil {
				return nil, err
			}

			for i := 0; i < num; i++ {
				f := r.Field(i)
				b, err = encode(b, r.Type().Field(i).Name)
				if err != nil {
					return nil, err
				}
				b, err = encode(b, f)
				if err != nil {
					return nil, err
				}
			}
			return appendItem(buf, 0x01, b), nil
		default:
			return nil, fmt.Errorf("Unsupported type: %s", r.Type().Kind())
		}
	}
}

func isNillable(k reflect.Kind) bool {
	switch k {
	case reflect.Map, reflect.Ptr, reflect.Interface, reflect.Slice:
		return true
	default:
		return false
	}
}

func fitInt(v int) interface{} {
	switch {
	case v < math.MaxInt8:
		return int8(v)
	case v < math.MaxInt16:
		return int16(v)
	case v < math.MaxInt32:
		return int32(v)
	default:
		return int64(v)
	}
}
