# GO-GETIT
This is a key-value database I built from scratch (for the purpose of sharpening up my Golang & to play around with binary protocols)

## Build
The build script contains an example of building the binary from source (in my example I have 2 copies of the executable but this was to run a copy at startup without a console window appearing; it runs perfectly fine from a single binary)

## Use
### Server
To get started run the binary with the flag `-runtime=server` to create a server that can start serving requests

### Client
- `store X Y` to store value Y in X (value can be a string or 32 bit number, strings are limited to 31 ASCII chars)
- `load X` to get value associated with key X (or empty return if not found)
- `clear {X}` to delete key X (or omit to clear all)
- `count` to get number of values in database
- `exit` shuts down the server
