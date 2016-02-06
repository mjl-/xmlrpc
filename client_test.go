package xmlrpc

import (
	"fmt"
)

// Call function "add" with parameters 1 and 2 on the server running on localhost, storing the return value in the provided int64.
func ExampleClient_Call() {
	c := &Client{"http://localhost/"}
	var i int64
	err := c.Call(&i, "add", 1, 2)
	if err != nil {
		switch err.(type) {
		case Fault:
			// function failed at server
		case ProtocolError:
			// protocol botch between client & server
		default:
			// some other error
		}
	}
	fmt.Println(i)
}
