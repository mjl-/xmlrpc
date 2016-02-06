# xmlrpc

XMLRPC client library.


# To do

- see if we can make better use of interfaces
- add a separate function for multicall?  where we turn the responses into Fault's if they are so
- make it work when we write an <int> to a int (maybe in a struct).  now it complains it should be an int64.
- make xmlrpc tags works for writing out structs too
- document how encoding & decoding is done, including struct tags for xmlrpc.

- work with charsets other than utf-8?
- properly do decoding when more specific types are passed in
- more test cases, also for invalid xml
- test with more explicit types passed in
- test if decoding into arrays work (as opposed to slices)

# testing

...
