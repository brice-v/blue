## `net` is the module that deals with tcp and udp connections
## clients and servers can be created with these basic functions


val __connect = _connect;
val __listen = _listen;
val __inspect = _inspect;
val __net_accept = _accept;
val __net_close = _net_close;
val __net_read = _net_read;
val __net_write = _net_write;

fun connect(transport="tcp", addr="localhost", port="18650") {
    ## `connect` creates a net object from the given parameters
    ##
    ## transport can be any valid go transport such as 'tcp' or 'udp'
    ## addr is the address of the network server
    ## port is the port of the network server
    ##
    ## connect(transport: str='tcp', addr: str='localhost', port: str='18650') -> 
    ##         {t: 'net', 'v': uint}
    __connect(transport, addr, port)
}

fun listen(transport="tcp", addr="localhost", port="18650") {
    ## `listen` creates a net object to that represents a net server
    ##
    ## transport can be any valid go transport such as 'tcp' or 'udp'
    ## addr is the address to listen on for the network server
    ## port is the port to listen on for the network server
    ##
    ## listen(transport: str='tcp', addr: str='localhost', port: str='18650') ->
    ##        {t: 'net/udp', v: uint} or {t: 'net/tcp', v: uint}
    __listen(transport, addr, port)
}

fun inspect(obj) {
    ## `inspect` prints out the details of the net object
    ##
    ## inspect(obj: {t: 'net'|'net/tcp'|'net/udp': v: uint}) -> str
    return match obj {
        {t: _, v: _} => {
            __inspect(obj.v, obj.t)
        },
        _ => {
            error("`inspect` expects object")
        },
    };
}

fun net_accept(net_id) {
    ## `net_accept` will accept connections on the listener server created via 'listen'
    ## note: this function should mostly be called with the core 'accept' function
    ##
    ## net_accept(net_id: uint) -> {t: 'net', v: uint}
    __net_accept(net_id)
}

fun net_close(net_id, net_str) {
    ## `net_close` will close the connection based on the net_str passed in
    ## note: this function should mostly be called with the core 'close' function
    ##
    ## net_close(net_id: uint, net_str: 'net/tcp'|'net/udp'|'net') -> null
    __net_close(net_id, net_str)
}

fun net_read(net_id, net_str, end_byte_or_len=null, as_bytes=false) {
    ## `net_read` will read on the connection based on the net_str passed in
    ## note: this function should mostly be called with the core 'read' function
    ##
    ## If end_byte_or_len is null, it will read to '\n'
    ## If end_byte_or_len is a single char string, that byte will be used
    ##    NOTE: In this case it should not be inside the given string
    ## NOTE: In both of these cases a string will be returned
    ## If end_byte_or_len is an int, it will read that # of bytes
    ##    NOTE: as_bytes will return bytes directly if true, otherwise it will be a string
    ##
    ## When reading, if a len is given the entire buffer is returned, otherwise if a 
    ## end byte is given, it will be removed from the output
    ##
    ## net_read(net_id: uint, net_str: 'net/tcp'|'net/udp'|'net', end_byte_or_len: str|int|null=null, as_bytes: bool=false) -> str|bytes
    __net_read(net_id, net_str, end_byte_or_len, as_bytes)
}

fun net_write(net_id, net_str, value, end_byte='\n') {
    ## `net_write` will write the value to the connection based on the net_str passed in
    ## note: this function should mostly be called with the 'write' function
    ##
    ## If end_byte is null, nothing will be appended
    ## If end_byte is a single char string, that byte will be used
    ##    NOTE: In this case it should not be inisde or appended to the end of value as that
    ##          will be done automatically
    ##
    ## net_write(net_id: uint, net_str: 'net/tcp'|'net/udp'|'net', value: str|bytes, end_byte: str|null=null) -> null
    __net_write(net_id, net_str, value, end_byte)
}
