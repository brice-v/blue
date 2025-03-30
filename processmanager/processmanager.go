package processmanager

import (
	"log"

	"github.com/lxzan/gws"
	kcp "github.com/xtaci/kcp-go"
)

// After Node Connect, assume dialer started

// spawn, will take node name and function (kind of like regular spawn but with node name)

// In listener handler, on a message check if its from spawn or send

func StartListener() {
	listener, err := kcp.Listen(":6666")
	if err != nil {
		log.Println(err.Error())
		return
	}
	app := gws.NewServer(&gws.BuiltinEventHandler{}, nil)
	app.RunListener(listener)
}

type ListenerEventHandler struct{}

func (leh ListenerEventHandler) OnOpen(socket *gws.Conn) {}

func (leh ListenerEventHandler) OnClose(socket *gws.Conn, err error) {}

func (leh ListenerEventHandler) OnPing(socket *gws.Conn, payload []byte) { _ = socket.WritePong(nil) }

func (leh ListenerEventHandler) OnPong(socket *gws.Conn, payload []byte) {}

func (leh ListenerEventHandler) OnMessage(socket *gws.Conn, message *gws.Message) {}

func StartDialer() {
	conn, err := kcp.Dial("127.0.0.1:6666")
	if err != nil {
		log.Println(err.Error())
		return
	}
	app, _, err := gws.NewClientFromConn(&gws.BuiltinEventHandler{}, nil, conn)
	if err != nil {
		log.Println(err.Error())
		return
	}
	app.ReadLoop()
}

type DialerEventHandler struct{}

func (deh DialerEventHandler) OnOpen(socket *gws.Conn) {}

func (deh DialerEventHandler) OnClose(socket *gws.Conn, err error) {}

func (deh DialerEventHandler) OnPing(socket *gws.Conn, payload []byte) { _ = socket.WritePong(nil) }

func (deh DialerEventHandler) OnPong(socket *gws.Conn, payload []byte) {}

func (deh DialerEventHandler) OnMessage(socket *gws.Conn, message *gws.Message) {}
