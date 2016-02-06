package xmlrpc

import (
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"reflect"
	"time"
)

type encoder struct {
	w io.Writer
}

func (enc *encoder) error(s string) {
	panic(ProtocolError(errors.New(s)))
}

func (enc *encoder) Write(s string) {
	io.WriteString(enc.w, s)
}

func (enc *encoder) Text(s string) {
	xml.EscapeText(enc.w, []byte(s))
}

func (enc *encoder) Value(v reflect.Value) {
	enc.Write("<value>")

	rv := v
	if v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		rv = v.Elem()
	}

	switch rv.Kind() {
	case reflect.Bool:
		enc.Write("<boolean>")
		if rv.Bool() {
			enc.Text("1")
		} else {
			enc.Text("0")
		}
		enc.Write("</boolean>")

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		enc.Write("<int>")
		enc.Text(fmt.Sprintf("%d", rv.Int()))
		enc.Write("</int>")

	case reflect.Int64:
		enc.Write("<i8>")
		enc.Text(fmt.Sprintf("%d", rv.Int()))
		enc.Write("</i8>")

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		enc.Write("<int>")
		enc.Text(fmt.Sprintf("%d", rv.Uint()))
		enc.Write("</int>")

	case reflect.Uint64:
		// not very good, remote side is likely to try to fit it in a signed int64...
		enc.Write("<i8>")
		enc.Text(fmt.Sprintf("%d", rv.Uint()))
		enc.Write("</i8>")

	case reflect.Float32, reflect.Float64:
		enc.Write("<double>")
		enc.Text(fmt.Sprintf("%f", rv.Float()))
		enc.Write("</double>")

	case reflect.String:
		enc.Text(rv.String())

	case reflect.Array, reflect.Slice:
		if rv.Type() == reflect.TypeOf([]byte(nil)) {
			enc.Write("<base64>")
			enc.Text(base64.StdEncoding.EncodeToString(rv.Bytes()))
			enc.Write("</base64>")
			break
		}

		enc.Write("<array><data>")
		n := rv.Len()
		for i := 0; i < n; i++ {
			enc.Value(rv.Index(i))
		}
		enc.Write("</data></array>")

	case reflect.Map:
		kt := rv.Type().Key()
		if kt.Kind() != reflect.String {
			enc.error("cannot encode maps with non-strings as keys")
		}
		enc.Write("<struct>")
		for _, k := range rv.MapKeys() {
			mv := rv.MapIndex(k)
			enc.Write("<member>")
			enc.Write("<name>")
			enc.Text(k.String())
			enc.Write("</name>")
			enc.Value(mv)
			enc.Write("</member>")
		}
		enc.Write("</struct>")

	case reflect.Struct:
		if t, ok := rv.Interface().(time.Time); ok {
			const iso8601 = "20060102T15:04:05"
			enc.Write("<dateTime.iso8601>")
			enc.Text(t.Format(iso8601))
			enc.Write("</dateTime.iso8601>")
			break
		}

		rt := rv.Type()
		n := rv.NumField()
		enc.Write("<struct>")
		for i := 0; i < n; i++ {
			ft := rt.Field(i)
			fv := rv.Field(i)

			enc.Write("<member>")
			enc.Write("<name>")
			enc.Text(ft.Name)
			enc.Write("</name>")
			enc.Value(fv)
			enc.Write("</member>")
		}
		enc.Write("</struct>")

	case reflect.Interface, reflect.Complex64, reflect.Complex128, reflect.Uintptr, reflect.Chan, reflect.Func, reflect.Ptr:
		enc.error(fmt.Sprintf("don't know how to encode %+v of kind %+v", v, rv.Kind()))

	default:
		panic("should not happen")
	}

	enc.Write("</value>")
}

func writeRequest(w io.Writer, method string, args ...interface{}) (err error) {
	defer func() {
		if e := recover(); e != nil {
			if ee, ok := e.(ProtocolError); ok {
				err = ee
				return
			}
			panic(e)
		}
	}()

	enc := encoder{w}
	_, xerr := io.WriteString(w, xml.Header)
	if xerr != nil {
		enc.error(xerr.Error())
	}

	enc.Write("<methodCall>")
	enc.Write("<methodName>")
	enc.Text(method)
	enc.Write("</methodName>")
	enc.Write("<params>")
	for _, v := range args {
		enc.Write("<param>")
		enc.Value(reflect.ValueOf(v))
		enc.Write("</param>")
	}
	enc.Write("</params>")
	enc.Write("</methodCall>")
	return
}
