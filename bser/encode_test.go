package bser

import (
	"bytes"
	"fmt"
	"testing"
)

type person struct {
	Name string
	Age  int
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
