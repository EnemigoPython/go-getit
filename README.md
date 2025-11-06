# GO-GETIT
This is a key-value database I built from scratch (for the purpose of sharpening up my Golang & to play around with binary protocols)

## Build
The build script contains an example of building the binary from source (in my example I have 2 copies of the executable but this was to run a copy at startup without a console window appearing; it runs perfectly fine from a single binary)

## Use
### Server
To get started run the binary with the flag `-runtime=server` to create a server that can start serving requests

### Client
- `store X Y` to store value Y in X (value can be a string or 32 bit number, strings are limited to 31 ASCII chars) -> returns `1` if new entry or `0` if data overwritten
- `load X` to get value associated with key X (or empty return if not found)
- `clear {X}` to delete key X (or omit to clear all) -> returns `0` if success or empty if not found
- `keys` streams all keys set in the store
- `values` streams all values set in the store
- `items` streams all keys & values in the store (space separated)
- `count` to get number of entries in the store
- `size` to get size of file in bytes
- `space {current/empty/max}` to get maximum number of entries possible in current file size -> empty gets unused entry space, max gets maximum possible, default current
- `exit` shuts down the server

### Config Flags
- `--runtime={client/server}` defaults to client
- `--port=X` to set the port
- `--store=X` sets the name of the store
- `--debug` starts in debug mode