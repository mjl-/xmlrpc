// Package xmlrpc is an XMLRPC client library that supports extensions
// such as the i8 type, and allowing the use of system.multicall.
package xmlrpc

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
)

// Client holds all information about an XMLRPC server needed to perform function calls.
type Client struct {
	Url string
}

// Call lets you call a remote XMLRPC function named method with parameters params, with the results return in r.
//
// A ProtocolError is returned on invalid XMLRPC message exchange.
// A Fault is returned for an explicit error from the server.
//
// r should be a pointer and is allowed to be a specific type. E.g. a string if that is the type of the expected response.
func (c *Client) Call(r interface{}, method string, params ...interface{}) error {
	out := new(bytes.Buffer)
	err := writeRequest(out, method, params...)
	if err != nil {
		return err
	}
	body := strings.NewReader(out.String())
	resp, err := http.Post(c.Url, "text/xml", body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("http response not successful (status %d)", resp.StatusCode)
	}
	return parseResponse(resp.Body, r)
}
