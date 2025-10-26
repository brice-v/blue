package object

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"strings"
)

var NetBuiltins = NewBuiltinSliceType{
	{Name: "_connect", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 3 {
				return newInvalidArgCountError("connect", len(args), 3, "")
			}
			if args[0].Type() != STRING_OBJ {
				return newPositionalTypeError("connect", 1, STRING_OBJ, args[0].Type())
			}
			if args[1].Type() != STRING_OBJ {
				return newPositionalTypeError("connect", 2, STRING_OBJ, args[1].Type())
			}
			if args[2].Type() != STRING_OBJ {
				return newPositionalTypeError("connect", 3, STRING_OBJ, args[2].Type())
			}
			transport := strings.ToLower(args[0].(*Stringo).Value)
			addr := args[1].(*Stringo).Value
			port := args[2].(*Stringo).Value
			addrStr := net.JoinHostPort(addr, port)
			conn, err := net.Dial(transport, addrStr)
			if err != nil {
				return newError("`connect` error: %s", err.Error())
			}
			return CreateBasicMapObjectForGoObj("net", NewGoObj(conn))
		},
		HelpStr: helpStrArgs{
			explanation: "`connect` connects to the given transport://addr:port",
			signature: `connect(transport: str('tcp'|'tcp4'|'tcp6'|'udp'|'udp4'|'udp6'|'ip'|'ip4'|'ip6'|'unix'|'unixgram'|'unixpacket')='tcp',
		addr: str='localhost', port: str='18650') -> {t: 'net', v: GoObj[net.Conn]}`,
			errors:  "InvalidArgCount,PositionalType,CustomError",
			example: "connect() => {t: 'net', v: GoObj[net.Conn]}",
		}.String(),
	}},
	{Name: "_listen", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 3 {
				return newInvalidArgCountError("listen", len(args), 3, "")
			}
			if args[0].Type() != STRING_OBJ {
				return newPositionalTypeError("listen", 1, STRING_OBJ, args[0].Type())
			}
			if args[1].Type() != STRING_OBJ {
				return newPositionalTypeError("listen", 2, STRING_OBJ, args[1].Type())
			}
			if args[2].Type() != STRING_OBJ {
				return newPositionalTypeError("listen", 3, STRING_OBJ, args[2].Type())
			}
			transport := strings.ToLower(args[0].(*Stringo).Value)
			addr := args[1].(*Stringo).Value
			port := args[2].(*Stringo).Value
			addrStr := fmt.Sprintf("%s:%s", addr, port)
			if strings.Contains(transport, "udp") {
				s, err := net.ResolveUDPAddr(transport, ":"+port)
				if err != nil {
					return newError("`listen` udp error: %s", err.Error())
				}
				l, err := net.ListenUDP(transport, s)
				if err != nil {
					return newError("`listen` udp error: %s", err.Error())
				}
				return CreateBasicMapObjectForGoObj("net/udp", NewGoObj(l))
			}
			l, err := net.Listen(transport, addrStr)
			if err != nil {
				return newError("`listen` error: %s", err.Error())
			}
			return CreateBasicMapObjectForGoObj("net/tcp", NewGoObj(l))
		},
		HelpStr: helpStrArgs{
			explanation: "`listen` listens for connections on the given transport://addr:port",
			signature: `listen(transport: str('tcp'|'tcp4'|'tcp6'|'udp'|'udp4'|'udp6'|'ip'|'ip4'|'ip6'|'unix'|'unixgram'|'unixpacket')='tcp',
		addr: str='localhost', port: str='18650') -> {t: 'net/tcp'|'net/udp', v: GoObj[net.Listener]|GoObj[*net.UDPConn]}`,
			errors:  "InvalidArgCount,PositionalType,CustomError",
			example: "listen() => {t: 'net/tcp', v: GoObj[net.Listener]}",
		}.String(),
	}},
	{Name: "_accept", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("accept", len(args), 1, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("accept", 1, GO_OBJ, args[0].Type())
			}
			l, ok := args[0].(*GoObj[net.Listener])
			if !ok {
				return newPositionalTypeErrorForGoObj("accept", 1, "net.Listener", args[0])
			}
			conn, err := l.Value.Accept()
			if err != nil {
				return newError("`accept` error: %s", err.Error())
			}
			return CreateBasicMapObjectForGoObj("net", NewGoObj(conn))
		},
		HelpStr: helpStrArgs{
			explanation: "`accept` accepts connections on the given listener",
			signature:   "accept(l: {t: 'net/tcp', v: GoObj[net.Listener]}) -> {t: 'net', v: GoObj[net.Conn]}",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "accept(l) => {t: 'net', v: GoObj[net.Conn]}",
		}.String(),
	}},
	{Name: "_net_close", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("net_close", len(args), 2, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("net_close", 1, GO_OBJ, args[0].Type())
			}
			if args[1].Type() != STRING_OBJ {
				return newPositionalTypeError("net_close", 2, STRING_OBJ, args[1].Type())
			}
			t := args[1].(*Stringo).Value
			switch t {
			case "net/udp":
				c, ok := args[0].(*GoObj[*net.UDPConn])
				if !ok {
					return newPositionalTypeErrorForGoObj("net_close", 1, "*net.UDPConn", args[0])
				}
				c.Value.Close()
			case "net/tcp":
				listener, ok := args[0].(*GoObj[net.Listener])
				if !ok {
					return newPositionalTypeErrorForGoObj("net_close", 1, "net.Listener", args[0])
				}
				listener.Value.Close()
			case "net":
				conn, ok := args[0].(*GoObj[net.Conn])
				if !ok {
					return newPositionalTypeErrorForGoObj("net_close", 1, "net.Conn", args[0])
				}
				conn.Value.Close()
			default:
				return newError("`net_close` expects type of 'net/tcp', 'net/udp', or 'net'")
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`net_close` closes the given connection/listener",
			signature:   "net_close(c: {t: 'net/tcp', v: GoObj[net.Listener]}|{t: 'net/udp', v: GoObj[*net.UDPConn]}|{t: 'net', v: GoObj[net.Conn]}) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "net_close(c) => null",
		}.String(),
	}},
	{Name: "_net_read", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 4 {
				return newInvalidArgCountError("net_read", len(args), 4, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("net_read", 1, GO_OBJ, args[0].Type())
			}
			if args[1].Type() != STRING_OBJ {
				return newPositionalTypeError("net_read", 2, STRING_OBJ, args[1].Type())
			}
			if args[2].Type() != NULL_OBJ && args[2].Type() != STRING_OBJ && args[2].Type() != INTEGER_OBJ {
				return newPositionalTypeError("net_read", 3, "NULL or STRING or INTEGER", args[2].Type())
			}
			if args[3].Type() != BOOLEAN_OBJ {
				return newPositionalTypeError("net_read", 4, BOOLEAN_OBJ, args[3].Type())
			}
			t := args[1].(*Stringo).Value
			var conn net.Conn
			if t == "net/udp" {
				c, ok := args[0].(*GoObj[*net.UDPConn])
				if !ok {
					return newPositionalTypeErrorForGoObj("net_read", 1, "*net.UDPConn", args[0])
				}
				conn = c.Value
			} else {
				c, ok := args[0].(*GoObj[net.Conn])
				if !ok {
					return newPositionalTypeErrorForGoObj("net_read", 1, "net.Conn", args[0])
				}
				conn = c.Value
			}
			if args[2].Type() == INTEGER_OBJ {
				asBytes := args[3].(*Boolean).Value
				bufLen := args[2].(*Integer).Value
				if bufLen == 0 {
					return newError("`net_read` error: len must not be 0")
				}
				buf := make([]byte, bufLen)
				readLen, err := bufio.NewReader(conn).Read(buf)
				if err != nil {
					return newError("`net_read` error: %s", err.Error())
				}
				if readLen != int(bufLen) {
					return newError("`net_read` error: read length (%d) does not match buffer length (%d)", readLen, bufLen)
				}
				if asBytes {
					return &Bytes{Value: buf}
				} else {
					return &Stringo{Value: string(buf)}
				}
			}
			var endByte byte
			if args[2].Type() == NULL_OBJ {
				endByte = '\n'
			} else {
				endByteString := args[2].(*Stringo).Value
				if len(endByteString) != 1 {
					return newError("`net_read` error: end byte given was not length 1, got=%d", len(endByteString))
				}
				endByte = []byte(endByteString)[0]
			}
			s, err := bufio.NewReader(conn).ReadString(endByte)
			if err != nil {
				return newError("`net_read` error: %s", err.Error())
			}
			return &Stringo{Value: s[:len(s)-1]}
		},
		HelpStr: helpStrArgs{
			explanation: "`net_read` reads on the given connection to end_byte (default '\\n') or the buffer length, returning a string or bytes if as_bytes is true",
			signature:   "net_read(conn_v: GoObj[*net.UDPConn]|GoObj[net.Conn], conn_t: 'net/tcp'|'net/udp'|'net', end_byte_or_len: str|int|null=null, as_bytes: bool=false) -> str|bytes",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "net_read(c.v, c.t) => 'test'",
		}.String(),
	}},
	{Name: "_net_write", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 3 {
				return newInvalidArgCountError("net_write", len(args), 3, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("net_write", 1, GO_OBJ, args[0].Type())
			}
			if args[1].Type() != STRING_OBJ {
				return newPositionalTypeError("net_write", 2, STRING_OBJ, args[1].Type())
			}
			if args[2].Type() != STRING_OBJ && args[2].Type() != BYTES_OBJ {
				return newPositionalTypeError("net_write", 3, "STRING or BYTES", args[2].Type())
			}
			if args[3].Type() != NULL_OBJ && args[3].Type() != STRING_OBJ {
				return newPositionalTypeError("net_write", 4, "NULL or STRING", args[3].Type())
			}
			t := args[1].(*Stringo).Value
			var conn net.Conn
			if t == "net/udp" {
				c, ok := args[0].(*GoObj[*net.UDPConn])
				if !ok {
					return newPositionalTypeErrorForGoObj("net_write", 1, "*net.UDPConn", args[0])
				}
				conn = c.Value
			} else {
				c, ok := args[0].(*GoObj[net.Conn])
				if !ok {
					return newPositionalTypeErrorForGoObj("net_write", 1, "net.Conn", args[0])
				}
				conn = c.Value
			}
			var appendByte *byte = nil
			if args[3].Type() == STRING_OBJ {
				endByteString := args[3].(*Stringo).Value
				if len(endByteString) != 1 {
					return newError("`net_read` error: end byte given was not length 1, got=%d", len(endByteString))
				}
				appendByte = &[]byte(endByteString)[0]
			}
			buf := bytes.Buffer{}
			if args[2].Type() == STRING_OBJ {
				s := args[2].(*Stringo).Value
				n, err := buf.WriteString(s)
				if err != nil {
					return newError("`net_write` error: failed writing to internal buffer. %s", err.Error())
				}
				if n != len(s) {
					return newError("`net_write` error: failed writing string of len %d to internal buffer, wrote=%d", len(s), n)
				}
			} else {
				bs := args[2].(*Bytes).Value
				n, err := buf.Write(bs)
				if err != nil {
					return newError("`net_write` error: failed writing to internal buffer. %s", err.Error())
				}
				if n != len(bs) {
					return newError("`net_write` error: failed writing bytes of len %d to internal buffer, wrote=%d", len(bs), n)
				}
			}
			if appendByte != nil {
				err := buf.WriteByte(*appendByte)
				if err != nil {
					return newError("`net_write` error: failed writing end byte %#+v to internal buffer. %s", *appendByte, err.Error())
				}
			}
			bs := buf.Bytes()
			n, err := conn.Write(bs)
			if err != nil {
				return newError("`net_write` error: %s", err.Error())
			}
			if n != len(bs) {
				return newError("`net_write` error: did not write the entire length got=%d, want=%d", n, len(bs))
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`net_write` writes the string/bytes on the given connection in full or to the end_byte (default null)",
			signature:   "net_write(conn_v: GoObj[*net.UDPConn]|GoObj[net.Conn], conn_t: 'net/tcp'|'net/udp'|'net', value: str|bytes, end_byte: str|null=null) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "net_write(c.v, c.t, 'test') => null",
		}.String(),
	}},
	{Name: "_inspect", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("inspect", len(args), 2, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("inspect", 1, GO_OBJ, args[0].Type())
			}
			if args[1].Type() != STRING_OBJ {
				return newPositionalTypeError("inspect", 2, STRING_OBJ, args[1].Type())
			}
			t := args[1].(*Stringo).Value
			switch t {
			case "net/udp":
				c, ok := args[0].(*GoObj[*net.UDPConn])
				if !ok {
					return newPositionalTypeErrorForGoObj("inspect", 1, "*net.UDPConn", args[0])
				}
				mapObj := NewOrderedMap[string, Object]()
				mapObj.Set("remote_addr", &Stringo{Value: c.Value.RemoteAddr().String()})
				mapObj.Set("local_addr", &Stringo{Value: c.Value.LocalAddr().String()})
				mapObj.Set("remote_addr_network", &Stringo{Value: c.Value.RemoteAddr().Network()})
				mapObj.Set("local_addr_network", &Stringo{Value: c.Value.LocalAddr().Network()})
				return CreateMapObjectForGoMap(*mapObj)
			case "net/tcp":
				l, ok := args[0].(*GoObj[net.Listener])
				if !ok {
					return newPositionalTypeErrorForGoObj("inspect", 1, "net.Listener", args[0])
				}
				mapObj := NewOrderedMap[string, Object]()
				mapObj.Set("addr", &Stringo{Value: l.Value.Addr().String()})
				mapObj.Set("addr_network", &Stringo{Value: l.Value.Addr().Network()})
				return CreateMapObjectForGoMap(*mapObj)
			case "net":
				c, ok := args[0].(*GoObj[net.Conn])
				if !ok {
					return newPositionalTypeErrorForGoObj("inspect", 1, "net.Conn", args[0])
				}
				mapObj := NewOrderedMap[string, Object]()
				mapObj.Set("remote_addr", &Stringo{Value: c.Value.RemoteAddr().String()})
				mapObj.Set("local_addr", &Stringo{Value: c.Value.LocalAddr().String()})
				mapObj.Set("remote_addr_network", &Stringo{Value: c.Value.RemoteAddr().Network()})
				mapObj.Set("local_addr_network", &Stringo{Value: c.Value.LocalAddr().Network()})
				return CreateMapObjectForGoMap(*mapObj)
			default:
				return newError("`inspect` expects type of 'net/tcp', 'net/udp', or 'net'")
			}
		},
		HelpStr: helpStrArgs{
			explanation: "`inspect` returns a map of info for the given net object",
			signature:   "inspect(conn_v: GoObj[*net.UDPConn]|GoObj[net.Conn]|GoObj[net.Listener], conn_t: 'net/tcp'|'net/udp'|'net') -> map[str:str]",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "inspect(c.v, c.t) => {'addr': '127.0.0.1', 'addr_network': 'tcp'}",
		}.String(),
	}},
}
