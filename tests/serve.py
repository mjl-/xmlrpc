#!/usr/bin/env python

import sys
from SimpleXMLRPCServer import SimpleXMLRPCServer
from SimpleXMLRPCServer import SimpleXMLRPCRequestHandler


def main(prog, *args):
	server = SimpleXMLRPCServer(("localhost", 8000), requestHandler=SimpleXMLRPCRequestHandler)
	server.register_introspection_functions()
	server.register_function(lambda a, b: a+b, 'add')
	server.register_function(lambda a, b: a-b, 'subtract')
	server.register_function(lambda a, i: a[i], 'index')
	server.register_function(lambda s: s.upper(), 'upper')
	server.register_function(lambda s: s.title(), 'title')
	print 'http://localhost:8000/'
	server.serve_forever()

if __name__ == '__main__':
	main(*sys.argv)
