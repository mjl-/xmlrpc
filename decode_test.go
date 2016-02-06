package xmlrpc

import (
	"os"
	"testing"
)

func testResponseValid(filename string, v interface{}, t *testing.T) {
	f, err := os.Open(filename)
	if err != nil {
		t.Errorf("open %s", filename)
	}
	defer f.Close()

	err = parseResponse(f, v)
	if err != nil {
		t.Errorf("parseResponse %s: %+v", filename, err)
	}
}

func TestRespBoolean(t *testing.T) {
	r := false
	var i interface{}
	testResponseValid("tests/xml/resp-boolean.xml", &r, t)
	testResponseValid("tests/xml/resp-boolean.xml", &i, t)
}

func TestRespDouble(t *testing.T) {
	r := 1.1
	var i interface{}
	testResponseValid("tests/xml/resp-double.xml", &r, t)
	testResponseValid("tests/xml/resp-double.xml", &i, t)
}

func TestRespEmptyString(t *testing.T) {
	s := ""
	var i interface{}
	testResponseValid("tests/xml/resp-empty-string.xml", &s, t)
	testResponseValid("tests/xml/resp-empty-string.xml", &i, t)
}

func TestRespI4(t *testing.T) {
	var v int64
	var i interface{}
	testResponseValid("tests/xml/resp-i4.xml", &v, t)
	testResponseValid("tests/xml/resp-i4.xml", &i, t)
}

func TestRespI8(t *testing.T) {
	var v int64
	var i interface{}
	testResponseValid("tests/xml/resp-i8.xml", &v, t)
	testResponseValid("tests/xml/resp-i8.xml", &i, t)
}

func TestRespImplicitEmptyString(t *testing.T) {
	var s string
	var i interface{}
	testResponseValid("tests/xml/resp-implicit-empty-string.xml", &s, t)
	testResponseValid("tests/xml/resp-implicit-empty-string.xml", &i, t)
}

func TestRespImplicitString(t *testing.T) {
	var s string
	var i interface{}
	testResponseValid("tests/xml/resp-implicit-string.xml", &s, t)
	testResponseValid("tests/xml/resp-implicit-string.xml", &i, t)
}

func TestRespInt(t *testing.T) {
	var v int64
	var i interface{}
	testResponseValid("tests/xml/resp-int.xml", &v, t)
	testResponseValid("tests/xml/resp-int.xml", &i, t)
}

func TestRespArray(t *testing.T) {
	var v []interface{}
	var i interface{}
	testResponseValid("tests/xml/resp-array.xml", &v, t)
	testResponseValid("tests/xml/resp-array.xml", &i, t)
}

func TestRespStruct(t *testing.T) {
	var v map[string]interface{}
	var i interface{}
	testResponseValid("tests/xml/resp-struct.xml", &v, t)
	testResponseValid("tests/xml/resp-struct.xml", &i, t)
}

func TestRespMulti(t *testing.T) {
	var i interface{}
	testResponseValid("tests/xml/resp-multicall.xml", &i, t)
}

func TestRespFault(t *testing.T) {
	filename := "tests/xml/resp-fault.xml"
	f, err := os.Open(filename)
	if err != nil {
		t.Errorf("open %s", filename)
	}
	defer f.Close()

	var v Fault
	err = parseResponse(f, &v)
	if err == nil {
		t.Errorf("expected Fault as response, so no error")
	}
	if _, ok := err.(Fault); ok {
		return
	}
	t.Errorf("expected Fault, saw %+v", err)

	var i interface{}
	err = parseResponse(f, &i)
	if err == nil {
		t.Errorf("expected Fault as response, so no error")
	}
	if _, ok := err.(Fault); ok {
		return
	}
	t.Errorf("expected Fault, saw %+v", err)
}
