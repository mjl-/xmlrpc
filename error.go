package xmlrpc

// Fault is an error as returned by the remote XMLRPC server.
type Fault struct {
	Code   int64  `xmlrpc:"faultCode"`
	String string `xmlrpc:"faultString"`
}

// Error returns the String, the error message from this Fault, as returned from the XMLRPC server.
func (f Fault) Error() string {
	return f.String
}

// ProtocolError is an error due to invalid message change between XMLRPC client and server.
type ProtocolError error
