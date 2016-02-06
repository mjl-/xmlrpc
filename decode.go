package xmlrpc

import (
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"reflect"
	"strconv"
	"time"
)

type decoder struct {
	d *xml.Decoder
	t xml.Token
}

func (d *decoder) error(s string) {
	panic(ProtocolError(errors.New(s)))
}

func (d *decoder) Token() xml.Token {
	if d.t != nil {
		t := d.t
		d.t = nil
		return t
	}
	t, err := d.d.Token()
	if err != nil {
		// eof is intentionally triggered when done parsing, pass it through as is
		if err == io.EOF {
			panic(err)
		}
		panic(ProtocolError(err))
	}
	return t
}

func (d *decoder) Peek() xml.Token {
	if d.t == nil {
		d.t = d.Token()
	}
	return d.t
}

func (d *decoder) PeekElem() xml.Token {
	for {
		t := d.Peek()
		switch t.(type) {
		case xml.StartElement:
			return t
		case xml.EndElement:
			return t
		}
		d.Token()
	}
}

func (d *decoder) StartOpt(name string, eat bool) {
	for {
		t := d.Token()
		switch tt := t.(type) {
		case xml.StartElement:
			if tt.Name.Local != name {
				d.error(fmt.Sprintf("expected start element %v, saw start element %v", name, tt.Name.Local))
			}
			return

		case xml.EndElement:
			d.error(fmt.Sprintf("expected start element %v, saw end element %+v", name, tt))
		}
		if !eat {
			d.error(fmt.Sprintf("expected start element %v, saw %+v", name, t))
		}
	}
}

func (d *decoder) Start(name string) {
	d.StartOpt(name, true)
}

func (d *decoder) EndOpt(name string, eat bool) {
	for {
		t := d.Token()
		switch tt := t.(type) {
		case xml.EndElement:
			if tt.Name.Local != name {
				d.error(fmt.Sprintf("expected end element %v, saw end element %v", name, tt.Name.Local))
			}
			return

		case xml.StartElement:
			d.error(fmt.Sprintf("expected end element %v, saw start element %+v", name, tt))
		}
		if !eat {
			d.error(fmt.Sprintf("expected end element %v, saw %+v", name, t))
		}
	}
}

func (d *decoder) End(name string) {
	d.EndOpt(name, true)
}

func (d *decoder) CharData() []byte {
	t := d.Token()
	tt, ok := t.(xml.CharData)
	if !ok {
		d.error(fmt.Sprintf("expected chardata, saw %+v", t))
	}
	return []byte(tt.Copy())
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func isStart(t xml.Token, name string) bool {
	tt, ok := t.(xml.StartElement)
	return ok && tt.Name.Local == name
}

func isEnd(t xml.Token, name string) bool {
	tt, ok := t.(xml.EndElement)
	return ok && tt.Name.Local == name
}

// <value> has already been read, we read until the </value>, but don't consume it
func (d *decoder) parseValue(rv reflect.Value) {
	typeName := ""
	var charData []byte
	for typeName == "" {
		t := d.Peek()
		switch tt := t.(type) {
		case xml.StartElement:
			typeName = tt.Name.Local
			if Debug {
				logger.Println("parseValue, new typeName", typeName)
			}

		case xml.CharData:
			charData = []byte(tt.Copy())
			d.Token()

		case xml.EndElement:
			// it's an implicit string
			rv.Set(reflect.ValueOf(string(charData)))
			return

		case xml.ProcInst:
			continue
		case xml.Directive:
			continue
		}
	}

	d.Start(typeName)
	switch typeName {
	case "string", "base64":
		// test for empty string
		t := d.Peek()
		tt, ok := t.(xml.EndElement)
		if ok && tt.Name.Local == typeName {
			d.Token()
			rv.Set(reflect.ValueOf(string("")))
			return
		}
		fallthrough

	case "boolean", "int", "i2", "i4", "i8", "double", "dateTime.iso8601":
		buf := d.CharData()

		switch typeName {
		case "boolean":
			switch string(buf) {
			case "0":
				rv.Set(reflect.ValueOf(false))
			case "1":
				rv.Set(reflect.ValueOf(true))
			default:
				d.error(fmt.Sprintf("invalid value %+v for boolean", string(buf)))
			}
		case "int", "i2", "i4":
			r, err := strconv.ParseInt(string(buf), 10, 32)
			check(err)
			rv.Set(reflect.ValueOf(r))
		case "i8":
			r, err := strconv.ParseInt(string(buf), 10, 64)
			check(err)
			rv.Set(reflect.ValueOf(r))
		case "string":
			r := string(buf)
			rv.Set(reflect.ValueOf(r))
		case "base64":
			r, err := base64.StdEncoding.DecodeString(string(buf))
			check(err)
			rv.Set(reflect.ValueOf(r))
		case "double":
			r, err := strconv.ParseFloat(string(buf), 64)
			check(err)
			rv.Set(reflect.ValueOf(r))
		case "dateTime.iso8601":
			const iso8601 = "20060102T15:04:05"
			r, err := time.Parse(iso8601, string(buf))
			check(err)
			rv.Set(reflect.ValueOf(r))
		}

	case "array":
		nv := rv
		if nv.IsNil() {
			var t []interface{}
			tt := reflect.TypeOf(t)
			xv := reflect.MakeSlice(tt, 0, 0)
			nv = reflect.New(xv.Type()).Elem()
			nv.Set(xv)
		}
		if nv.Kind() != reflect.Slice {
			d.error(fmt.Sprintf("decoding xmlrpc array, but value not a slice, but %+v", nv.Kind()))
		}
		d.Start("data")
	array:
		for i := 0; ; i += 1 {
			t := d.PeekElem()
			switch t.(type) {
			case xml.StartElement:
				d.Start("value")

				if i+1 >= nv.Cap() {
					newcap := nv.Cap() + nv.Cap()/2
					if newcap < 4 {
						newcap = 4
					}
					nnv := reflect.MakeSlice(nv.Type(), nv.Len(), newcap)
					reflect.Copy(nnv, nv)
					nv.Set(nnv)
				}
				if i+1 >= nv.Len() {
					nv.SetLen(i + 1)
				}

				ev := nv.Index(i)
				d.parseValue(ev)
				d.End("value")
			case xml.EndElement:
				d.End("data")
				rv.Set(nv)
				break array
			}
		}

	case "struct":
		nv := rv

		structfields := map[string]reflect.Value{}

		switch nv.Kind() {
		case reflect.Interface:
			if nv.NumMethod() != 0 {
				d.error("can not decode xmlrpc struct into non-empty interface")
			}
		case reflect.Map:
			if nv.Type().Key().Kind() != reflect.String {
				d.error(fmt.Sprintf("can only decode xmlrpc struct into a map[string], not %+v", nv.Type()))
			}

		case reflect.Struct:
			n := nv.Type().NumField()
			for i := 0; i < n; i++ {
				sf := nv.Type().Field(i)
				k := sf.Tag.Get("xmlrpc")
				if k == "" {
					k = sf.Name
				}
				structfields[k] = nv.Field(i)
			}
		default:
			d.error(fmt.Sprintf("cannot decode xmlrpc struct into %+v", nv.Type()))
		}

		if (nv.Kind() == reflect.Interface || nv.Kind() == reflect.Map) && nv.IsNil() {
			m := map[string]interface{}{}
			t := reflect.TypeOf(m)
			tv := reflect.MakeMap(t)
			nv = reflect.New(t).Elem()
			nv.Set(tv)
		}

	xstruct:
		for {
			t := d.PeekElem()
			switch t.(type) {
			case xml.StartElement:
				d.Start("member")
				d.Start("name")
				k := d.CharData()
				d.End("name")
				d.Start("value")

				switch nv.Kind() {
				case reflect.Map:
					var tv interface{}
					xv := reflect.ValueOf(&tv).Elem()
					d.parseValue(xv)
					nv.SetMapIndex(reflect.ValueOf(string(k)), xv)

				case reflect.Struct:
					f := structfields[string(k)]
					log.Println("struct field, key ", string(k), "field is", f)
					if !f.IsValid() {
						d.error(fmt.Sprintf("cannot find struct field named %+v", string(k)))
					}
					d.parseValue(f)

				default:
					panic("cannot happen")
				}
				d.End("value")
				d.End("member")
			case xml.EndElement:
				rv.Set(nv)
				break xstruct
			}
		}

	default:
		d.error(fmt.Sprintf("unexpected type %+v", typeName))
	}
	d.End(typeName)
}

func parseResponse(r io.Reader, v interface{}) (err error) {
	needeof := false
	defer func() {
		if e := recover(); e != nil {
			if e == io.EOF {
				if needeof {
					return
				}
				e = ProtocolError(errors.New("eof while parsing xmlrpc response"))
			}
			panic(e)
		}
	}()

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		panic(ProtocolError(errors.New("can only decode into non-nil ptr")))
	}
	rv = rv.Elem()

	xmldec := xml.NewDecoder(r)
	d := &decoder{d: xmldec}

	d.Start("methodResponse")
	t := d.PeekElem()
	if isStart(t, "fault") {
		d.Start("fault")
		d.Start("value")
		var f Fault
		d.parseValue(reflect.ValueOf(&f).Elem())
		d.End("value")
		d.End("fault")
		return f
	}

	d.Start("params")
	for {
		if Debug {
			logger.Println("parsing param")
		}
		t := d.PeekElem()
		if isEnd(t, "params") {
			if Debug {
				logger.Println("end of params")
			}
			d.Token()
			break
		} else if isStart(t, "param") {
			if Debug {
				logger.Println("new param")
			}
			d.Token()
			d.Start("value")
			d.parseValue(rv)
			d.End("value")
			d.End("param")
			if Debug {
				logger.Println("have param", v)
			}
		} else {
			d.error(fmt.Sprintf("unexpected token: %+v", t))
		}
	}
	d.End("methodResponse")
	if Debug {
		logger.Println("have end of methodResponse")
	}
	needeof = true
	d.End("") // raises eof
	return nil
}
