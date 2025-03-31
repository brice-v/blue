package processmanager

import (
	"blue/object"
	"log"
	"net"

	"github.com/lxzan/gws"
	"github.com/puzpuzpuz/xsync/v3"
	kcp "github.com/xtaci/kcp-go"
)

var NodeNameToConnection = xsync.NewMapOf[string, net.Conn]()

// After Node Connect, assume dialer started

// spawn, will take node name and function (kind of like regular spawn but with node name)

// In listener handler, on a message check if its from spawn or send

func StartListener() {
	listener, err := kcp.Listen(":6666")
	if err != nil {
		log.Println(err.Error())
		return
	}
	app := gws.NewServer(&ListenerEventHandler{}, nil)
	app.RunListener(listener)
}

type ListenerEventHandler struct{}

func (leh ListenerEventHandler) OnOpen(socket *gws.Conn) {}

func (leh ListenerEventHandler) OnClose(socket *gws.Conn, err error) {}

func (leh ListenerEventHandler) OnPing(socket *gws.Conn, payload []byte) { _ = socket.WritePong(nil) }

func (leh ListenerEventHandler) OnPong(socket *gws.Conn, payload []byte) {}

func (leh ListenerEventHandler) OnMessage(socket *gws.Conn, message *gws.Message) {
	if message.Data == nil {
		socket.WriteClose(1000, nil)
	}
	bs := message.Data.Bytes()
	object.Decode(bs)
}

func StartNodeConnection(addr string) error {
	conn, err := kcp.Dial("127.0.0.1:6666")
	if err != nil {
		return err
	}
	app, _, err := gws.NewClientFromConn(&gws.BuiltinEventHandler{}, nil, conn)
	if err != nil {
		return err
	}
	go handleNodeDialerWrites(app)
	return nil
}

var ProcessManagerDialerChan = make(chan []byte)

func handleNodeDialerWrites(app *gws.Conn) {
	for {
		if ProcessManagerDialerChan == nil {
			break
		}
		bs := <-ProcessManagerDialerChan
		app.WriteMessage(gws.OpcodeBinary, bs)
	}
	app.WriteClose(1000, nil)
}
