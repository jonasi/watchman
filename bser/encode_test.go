package bser

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
	"testing"
)

type person struct {
	Name string
	Age  int
}

type student struct {
	Name  string
	Age   int
	grade int
}

type Person person

type superperson struct {
	Person
	Power string
}

type superperson2 struct {
	*Person
	Power string
}

type unencodable struct {
	C chan string
}

type stringAlias string
type int8Alias int8
type int16Alias int16
type int32Alias int32
type int64Alias int64
type intAlias int
type boolAlias bool

type customEncoding string

func (c customEncoding) MarshalBSER() ([]byte, error) {
	i, err := strconv.Atoi(string(c))
	if err != nil {
		return nil, err
	}
	return MarshalValue(i)
}

func (c *customEncoding) UnmarshalBSER(b []byte) error {
	var i int
	if err := UnmarshalValue(b, &i); err != nil {
		return err
	}
	*c = customEncoding(strconv.Itoa(i))
	return nil
}

type encodeTest struct {
	data        interface{}
	expectedEnc []byte
	expectErr   bool
}

/*
expected values generated using python client,
just modifying headers since python implementation always uses int32 - https://github.com/facebook/watchman/blob/master/python/pywatchman/pybser.py#L70
ex:
>>> import pywatchman
>>> obj = {"Name": "fred", "Age": 20}
>>> s = pywatchman.bser.dumps(obj)
>>> print(s)
b'\x00\x01\x05\x19\x00\x00\x00\x01\x03\x02\x02\x03\x04Name\x02\x03\x04fred\x02\x03\x03Age\x03\x14'
>>>
*/

var encodeTests = map[string]encodeTest{
	"single_object": {
		data: person{Name: "fred", Age: 20},
		expectedEnc: []byte(
			"\x00\x01\x03\x19\x01\x03\x02\x02\x03\x04Name\x02\x03\x04fred\x02\x03\x03Age\x03\x14",
		),
	},
	"single_object_unexported_field": {
		data: student{Name: "fred", Age: 20, grade: 13},
		// same as previous case since "grade" field will not be encoded
		expectedEnc: []byte(
			"\x00\x01\x03\x19\x01\x03\x02\x02\x03\x04Name\x02\x03\x04fred\x02\x03\x03Age\x03\x14",
		),
	},
	"pos_int8_array": {
		data: [6]int{1, 2, 3, 4, 5, 6},
		expectedEnc: []byte(
			"\x00\x01\x03\x0f\x00\x03\x06\x03\x01\x03\x02\x03\x03\x03\x04\x03\x05\x03\x06",
		),
	},
	"string_slice": {
		data: []string{"a", "b", "c", "d", "e", "f"},
		expectedEnc: []byte(
			"\x00\x01\x03\x1b\x00\x03\x06\x02\x03\x01a\x02\x03\x01b\x02\x03\x01c\x02\x03\x01d\x02\x03\x01e\x02\x03\x01f",
		),
	},
	"pos_int8_slice": {
		data: []int{1, 2, 3, 4, 5, 6},
		expectedEnc: []byte(
			"\x00\x01\x03\x0f\x00\x03\x06\x03\x01\x03\x02\x03\x03\x03\x04\x03\x05\x03\x06",
		),
	},
	"neg_int8_slice": {
		data: []int{-1, -2, -3, -4, -5, -6},
		expectedEnc: []byte(
			"\x00\x01\x03\x0f\x00\x03\x06\x03\xff\x03\xfe\x03\xfd\x03\xfc\x03\xfb\x03\xfa",
		),
	},
	"pos_int16_slice": {
		data: []int{1001, 1002, 1003, 1004, 1005, 1006},
		expectedEnc: []byte(
			"\x00\x01\x03\x15\x00\x03\x06\x04\xe9\x03\x04\xea\x03\x04\xeb\x03\x04\xec\x03\x04\xed\x03\x04\xee\x03",
		),
	},
	"neg_int16_slice": {
		data: []int{-1001, -1002, -1003, -1004, -1005, -1006},
		expectedEnc: []byte(
			"\x00\x01\x03\x15\x00\x03\x06\x04\x17\xfc\x04\x16\xfc\x04\x15\xfc\x04\x14\xfc\x04\x13\xfc\x04\x12\xfc",
		),
	},
	"pos_int32_slice": {
		data: []int{100001, 100002, 100003, 100004, 100005, 100006},
		expectedEnc: []byte(
			"\x00\x01\x03!\x00\x03\x06\x05\xa1\x86\x01\x00\x05\xa2\x86\x01\x00\x05\xa3\x86\x01\x00\x05\xa4\x86\x01\x00\x05\xa5\x86\x01\x00\x05\xa6\x86\x01\x00",
		),
	},
	"neg_int32_slice": {
		data: []int{-100001, -100002, -100003, -100004, -100005, -100006},
		expectedEnc: []byte(
			"\x00\x01\x03!\x00\x03\x06\x05_y\xfe\xff\x05^y\xfe\xff\x05]y\xfe\xff\x05\\y\xfe\xff\x05[y\xfe\xff\x05Zy\xfe\xff",
		),
	},
	"pos_int64_slice": {
		data: []int{3_000_000_000, 3_000_000_001, 3_000_000_002, 3_000_000_003, 3_000_000_004, 3_000_000_005},
		expectedEnc: []byte(
			"\x00\x01\x039\x00\x03\x06\x06\x00^\xd0\xb2\x00\x00\x00\x00\x06\x01^\xd0\xb2\x00\x00\x00\x00\x06\x02^\xd0\xb2\x00\x00\x00\x00\x06\x03^\xd0\xb2\x00\x00\x00\x00\x06\x04^\xd0\xb2\x00\x00\x00\x00\x06\x05^\xd0\xb2\x00\x00\x00\x00",
		),
	},
	"neg_int64_slice": {
		data: []int{-3_000_000_000, -3_000_000_001, -3_000_000_002, -3_000_000_003, -3_000_000_004, -3_000_000_005},
		expectedEnc: []byte(
			"\x00\x01\x039\x00\x03\x06\x06\x00\xa2/M\xff\xff\xff\xff\x06\xff\xa1/M\xff\xff\xff\xff\x06\xfe\xa1/M\xff\xff\xff\xff\x06\xfd\xa1/M\xff\xff\xff\xff\x06\xfc\xa1/M\xff\xff\xff\xff\x06\xfb\xa1/M\xff\xff\xff\xff",
		),
	},
	"float64_slice": {
		data: []float64{0.01, 0.02, 0.03, 0.04, 0.05, 0.06},
		expectedEnc: []byte(
			"\x00\x01\x039\x00\x03\x06\x07{\x14\xaeG\xe1z\x84?\x07{\x14\xaeG\xe1z\x94?\x07\xb8\x1e\x85\xebQ\xb8\x9e?\x07{\x14\xaeG\xe1z\xa4?\x07\x9a\x99\x99\x99\x99\x99\xa9?\x07\xb8\x1e\x85\xebQ\xb8\xae?",
		),
	},
	"map_str_bool_true_only": {
		// only using one entry since map iteration order is not guaranteed
		data: map[string]bool{
			"a": true,
		},
		expectedEnc: []byte(
			"\x00\x01\x03\x08\x01\x03\x01\x02\x03\x01a\x08",
		),
	},
	"map_str_bool_false_only": {
		// only using one entry since map iteration order is not guaranteed
		data: map[string]bool{
			"a": false,
		},
		expectedEnc: []byte(
			"\x00\x01\x03\x08\x01\x03\x01\x02\x03\x01a\x09",
		),
	},
	"string_alias": {
		data: stringAlias("hello"),
		expectedEnc: []byte(
			"\x00\x01\x03\x08\x02\x03\x05hello",
		),
	},
	"int8_alias": {
		data: int8Alias(2),
		expectedEnc: []byte(
			"\x00\x01\x03\x02\x03\x02",
		),
	},
	"int16_alias": {
		data: int16Alias(1001),
		expectedEnc: []byte(
			"\x00\x01\x03\x03\x04\xe9\x03",
		),
	},
	"int32_alias": {
		data: int32Alias(100001),
		expectedEnc: []byte(
			"\x00\x01\x03\x05\x05\xa1\x86\x01\x00",
		),
	},
	"int64_alias": {
		data: int64Alias(3_000_000_001),
		expectedEnc: []byte(
			"\x00\x01\x03\x09\x06\x01^\xd0\xb2\x00\x00\x00\x00",
		),
	},
	"int_alias_int32_value": {
		data: intAlias(100001),
		expectedEnc: []byte(
			"\x00\x01\x03\x05\x05\xa1\x86\x01\x00",
		),
	},
	"bool_alias": {
		data: boolAlias(false),
		expectedEnc: []byte(
			"\x00\x01\x03\x01\x09",
		),
	},
	"template_slice": {
		data: []person{
			{Name: "fred", Age: 20},
			{Name: "pete", Age: 30},
			{Age: 25},
		},
		// copied from https://facebook.github.io/watchman/docs/bser.html#array-of-templated-objects
		expectedEnc: []byte{
			0x00, 0x01, 0x03, 0x2a,
			0x0b,
			0x00,
			0x03, 0x02,
			0x02,
			0x03, 0x04,
			0x4e, 0x61, 0x6d, 0x65, // "Name" instead of "name"
			0x02,
			0x03, 0x03,
			0x41, 0x67, 0x65, // "Age" instead of "age"
			0x03, 0x03,
			0x02,
			0x03, 0x04,
			0x66, 0x72, 0x65, 0x64,
			0x03, 0x14,
			0x02,
			0x03, 0x04,
			0x70, 0x65, 0x74, 0x65,
			0x03, 0x1e,
			0x02, 0x03, 0x00, // empty string instead of 0x0c
			0x03, 0x19,
		},
	},
	"template_slice_with_unexported_field": {
		data: []student{
			{Name: "fred", Age: 20, grade: 10},
			{Name: "pete", Age: 30, grade: 11},
			{Age: 25, grade: 12},
		},
		// copied from https://facebook.github.io/watchman/docs/bser.html#array-of-templated-objects
		// same as previous case since "grade" field will not be encoded
		expectedEnc: []byte{
			0x00, 0x01, 0x03, 0x2a,
			0x0b,
			0x00,
			0x03, 0x02,
			0x02,
			0x03, 0x04,
			0x4e, 0x61, 0x6d, 0x65, // "Name" instead of "name"
			0x02,
			0x03, 0x03,
			0x41, 0x67, 0x65, // "Age" instead of "age"
			0x03, 0x03,
			0x02,
			0x03, 0x04,
			0x66, 0x72, 0x65, 0x64,
			0x03, 0x14,
			0x02,
			0x03, 0x04,
			0x70, 0x65, 0x74, 0x65,
			0x03, 0x1e,
			0x02, 0x03, 0x00, // empty string instead of 0x0c
			0x03, 0x19,
		},
	},
	"template_non-nil-pointer_slice": {
		data: []*person{
			&person{Name: "fred", Age: 20},
			&person{Name: "pete", Age: 30},
			&person{Age: 25},
		},
		// copied from https://facebook.github.io/watchman/docs/bser.html#array-of-templated-objects
		expectedEnc: []byte{
			0x00, 0x01, 0x03, 0x2a,
			0x0b,
			0x00,
			0x03, 0x02,
			0x02,
			0x03, 0x04,
			0x4e, 0x61, 0x6d, 0x65, // "Name" instead of "name"
			0x02,
			0x03, 0x03,
			0x41, 0x67, 0x65, // "Age" instead of "age"
			0x03, 0x03,
			0x02,
			0x03, 0x04,
			0x66, 0x72, 0x65, 0x64,
			0x03, 0x14,
			0x02,
			0x03, 0x04,
			0x70, 0x65, 0x74, 0x65,
			0x03, 0x1e,
			0x02, 0x03, 0x00, // empty string instead of 0x0c
			0x03, 0x19,
		},
	},
	"pointer_slice_with_nil": {
		data: []*person{
			&person{Name: "fred", Age: 20},
			nil,
		},
		expectedEnc: []byte(
			"\x00\x01\x03\x1d\x00\x03\x02\x01\x03\x02\x02\x03\x04Name\x02\x03\x04fred\x02\x03\x03Age\x03\x14\n",
		),
	},
	"template_arr": {
		data: [3]person{
			{Name: "fred", Age: 20},
			{Name: "pete", Age: 30},
			{Age: 25},
		},
		// copied from https://facebook.github.io/watchman/docs/bser.html#array-of-templated-objects
		expectedEnc: []byte{
			0x00, 0x01, 0x03, 0x2a,
			0x0b,
			0x00,
			0x03, 0x02,
			0x02,
			0x03, 0x04,
			0x4e, 0x61, 0x6d, 0x65, // "Name" instead of "name"
			0x02,
			0x03, 0x03,
			0x41, 0x67, 0x65, // "Age" instead of "age"
			0x03, 0x03,
			0x02,
			0x03, 0x04,
			0x66, 0x72, 0x65, 0x64,
			0x03, 0x14,
			0x02,
			0x03, 0x04,
			0x70, 0x65, 0x74, 0x65,
			0x03, 0x1e,
			0x02, 0x03, 0x00, // empty string instead of 0x0c
			0x03, 0x19,
		},
	},
	"raw_message_slice": {
		// this is just an array of strings: ["ok", "there"]
		data: RawMessage([]byte(
			"\x01\x03\x01\x02\x03\x06Values\x00\x03\x02\x02\x03\x02ok\x02\x03\x05there",
		)),
		expectedEnc: []byte(
			"\x00\x01\x03\x1c\x01\x03\x01\x02\x03\x06Values\x00\x03\x02\x02\x03\x02ok\x02\x03\x05there",
		),
	},
	"custom_marshaller": {
		data:        customEncoding("123"),
		expectedEnc: []byte("\x00\x01\x03\x02\x03\x7b"), // int instead of string due to custom marshal function
	},
	"map_string_custom_marshaller": {
		data: map[string]customEncoding{
			"a": customEncoding("123"),
			"b": customEncoding("456"),
		},
		expectedEnc: []byte("\x00\x01\x03\x10\x01\x03\x02\x02\x03\x01a\x03{\x02\x03\x01b\x04\xc8\x01"),
	},
	"custom_marshaller_fails": {
		data:      customEncoding("abc"), // custom marshal returns error if can't convert string to int
		expectErr: true,
	},
	"map_str_chan": {
		data: map[string]chan string{
			"a": make(chan string),
			"b": make(chan string),
		},
		expectErr: true,
	},
	"chan_slice": {
		data: []chan string{
			make(chan string),
			make(chan string),
		},
		expectErr: true,
	},
	"object_with_chan_field": {
		data: unencodable{
			C: make(chan string),
		},
		expectErr: true,
	},
	"func": {
		data: func() int {
			return 0
		},
		expectErr: true,
	},
}

func TestEncode(t *testing.T) {
	for testName, testCase := range encodeTests {
		t.Run(testName, func(t *testing.T) {
			testEncode(t, testCase)
		})
	}
}

func testEncode(t *testing.T, testCase encodeTest) {
	t.Helper()

	b, err := MarshalPDU(testCase.data)
	if testCase.expectErr && err != nil {
		return
	}
	if testCase.expectErr && err == nil {
		t.Fatal("unexpectedly no error")
	}
	if !testCase.expectErr && err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if !bytes.Equal(testCase.expectedEnc, b) {
		t.Fatalf("unexpected encoded data:\n\nexpected = %v\n\nactual = %v", testCase.expectedEnc, b)
	}
}

var encoderBenches = map[string]interface{}{
	"single_2_field_object": person{Name: "fred", Age: 20},
	"custom_marshaler":      customEncoding("123"),
}

var benchEncErr error

func BenchmarkEncoder(b *testing.B) {
	for benchName, data := range encoderBenches {
		var err error
		b.Run(fmt.Sprintf("data=%s", benchName), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err = MarshalPDU(data)
			}
		})
		benchEncErr = err
	}
}

var encodeSliceBenches = map[string]func(numElems int) interface{}{
	"string": func(numElems int) interface{} {
		data := make([]string, numElems)
		for i := 0; i < numElems; i++ {
			data[i] = strconv.Itoa(i)
		}
		return data
	},
	"int8": func(numElems int) interface{} {
		data := make([]int8, numElems)
		for i := 0; i < numElems; i++ {
			data[i] = int8(i)
		}
		return data
	},
	"int16": func(numElems int) interface{} {
		data := make([]int16, numElems)
		for i := 0; i < numElems; i++ {
			data[i] = int16(math.MaxInt8 + i)
		}
		return data
	},
	"int32": func(numElems int) interface{} {
		data := make([]int32, numElems)
		for i := 0; i < numElems; i++ {
			data[i] = int32(math.MaxInt16 + i)
		}
		return data
	},
	"int64": func(numElems int) interface{} {
		data := make([]int64, numElems)
		for i := 0; i < numElems; i++ {
			data[i] = int64(math.MaxInt32 + i)
		}
		return data
	},
	"float64": func(numElems int) interface{} {
		data := make([]float64, numElems)
		for i := 0; i < numElems; i++ {
			data[i] = float64(i)
		}
		return data
	},
}

var benchEncSliceErr error

func BenchmarkEncodeSlice(b *testing.B) {
	allNumElems := []int{1, 3, 5, 10, 20, 50, 100}
	for dtype, dataFn := range encodeSliceBenches {
		b.Run(fmt.Sprintf("dtype=%s", dtype), func(b *testing.B) {
			for _, numElems := range allNumElems {
				var err error
				b.Run(fmt.Sprintf("num_elems=%d", numElems), func(b *testing.B) {
					b.StopTimer()
					data := dataFn(numElems)
					b.StartTimer()
					for i := 0; i < b.N; i++ {
						_, err = MarshalPDU(data)
					}
				})
				benchEncSliceErr = err
			}
		})
	}
}

var encodeMapBenches = map[string]func(numElems int) interface{}{
	"string": func(numElems int) interface{} {
		data := map[string]string{}
		for i := 0; i < numElems; i++ {
			data[strconv.Itoa(i)] = strconv.Itoa(i)
		}
		return data
	},
	"int8": func(numElems int) interface{} {
		data := map[int8]int8{}
		for i := 0; i < numElems; i++ {
			data[int8(i)] = int8(i)
		}
		return data
	},
	"int16": func(numElems int) interface{} {
		data := map[int16]int16{}
		for i := 0; i < numElems; i++ {
			data[int16(i)] = int16(math.MaxInt8 + i)
		}
		return data
	},
	"int32": func(numElems int) interface{} {
		data := map[int32]int32{}
		for i := 0; i < numElems; i++ {
			data[int32(i)] = int32(math.MaxInt16 + i)
		}
		return data
	},
	"int64": func(numElems int) interface{} {
		data := map[int64]int64{}
		for i := 0; i < numElems; i++ {
			data[int64(i)] = int64(math.MaxInt32 + i)
		}
		return data
	},
	"float64": func(numElems int) interface{} {
		data := map[float64]float64{}
		for i := 0; i < numElems; i++ {
			data[float64(i)] = float64(i)
		}
		return data
	},
}

var benchEncMapErr error

func BenchmarkEncodeMap(b *testing.B) {
	allNumElems := []int{1, 3, 5, 10, 20, 50, 100}
	for dtype, dataFn := range encodeMapBenches {
		b.Run(fmt.Sprintf("dtype=%s", dtype), func(b *testing.B) {
			for _, numElems := range allNumElems {
				var err error
				b.Run(fmt.Sprintf("num_elems=%d", numElems), func(b *testing.B) {
					b.StopTimer()
					data := dataFn(numElems)
					b.StartTimer()
					for i := 0; i < b.N; i++ {
						_, err = MarshalPDU(data)
					}
				})
				benchEncMapErr = err
			}
		})
	}
}

var encodeStructArrBenches = map[string]func(numElems int) interface{}{
	"num_fields=1": func(numElems int) interface{} {
		type tmp struct {
			A int
		}
		s := make([]tmp, numElems)
		for i := 0; i < numElems; i++ {
			s[i] = tmp{i}
		}
		return s
	},
	"num_fields=3": func(numElems int) interface{} {
		type tmp struct {
			A int
			B int
			C int
		}
		s := make([]tmp, numElems)
		for i := 0; i < numElems; i++ {
			s[i] = tmp{A: i, B: i, C: i}
		}
		return s
	},
	"num_fields=5": func(numElems int) interface{} {
		type tmp struct {
			A int
			B int
			C int
			D int
			E int
		}
		s := make([]tmp, numElems)
		for i := 0; i < numElems; i++ {
			s[i] = tmp{A: i, B: i, C: i, D: i, E: i}
		}
		return s
	},
	"num_fields=10": func(numElems int) interface{} {
		type tmp struct {
			A int
			B int
			C int
			D int
			E int
			F int
			G int
			H int
			I int
			J int
		}
		s := make([]tmp, numElems)
		for i := 0; i < numElems; i++ {
			s[i] = tmp{
				A: i, B: i, C: i, D: i, E: i,
				F: i, G: i, H: i, I: i, J: i,
			}
		}
		return s
	},
}

var benchEncStructArrErr error

func BenchmarkEncodeStructArr(b *testing.B) {
	allNumElems := []int{1, 3, 5, 10, 20, 50, 100}
	for benchName, newDataFn := range encodeStructArrBenches {
		b.Run(benchName, func(b *testing.B) {
			for _, numElems := range allNumElems {
				var err error
				b.Run(fmt.Sprintf("num_elems=%d", numElems), func(b *testing.B) {
					b.StopTimer()
					data := newDataFn(numElems)
					b.StartTimer()
					for i := 0; i < b.N; i++ {
						_, err = MarshalPDU(data)
					}
				})
				benchEncStructArrErr = err
			}
		})
	}
}
