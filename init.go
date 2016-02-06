package xmlrpc

import (
	"log"
	"os"
)

var logger *log.Logger

// Debug determines whether debug messages are printed for reading and writing XMLRPC messages.
var Debug = false

func init() {
	logger = log.New(os.Stdout, "xmlrpc: ", log.LstdFlags)
}
