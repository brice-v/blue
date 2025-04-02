package processmanager

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/puzpuzpuz/xsync/v3"
)

type Payload struct {
	InitialPacket []byte
	FullPacket    []byte
}

type ConnAndCh struct {
	Conn    net.Conn
	WriteCh chan Payload
	RespCh  chan Payload // TODO: I think what well use this for is to send responses on?
}

var NodeNameToConnection = xsync.NewMapOf[string, *ConnAndCh]()

// After Node Connect, assume dialer started

// spawn, will take node name and function (kind of like regular spawn but with node name)

// In listener handler, on a message check if its from spawn or send

func StartListener(nodeName, addr string) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Println(err.Error())
		return
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("TODO What to do with this error: %s", err.Error())
		}
		go handleListenerConnection(conn)
	}
}

func handleListenerConnection(conn net.Conn) {
	r := bufio.NewReader(conn)
	for {
		// Always read first set of bytes with delimeter of \r
		bs, err := r.ReadBytes('\r')
		if err != nil {
			log.Printf("listener reader error: failed to read bytes: %s", err.Error())
			return
		}
		// first byte will be s, o, or e
		// s for spawn, o for object, e for error
		// after that first byte, the rest is the # of bytes to read next
		numBytes, err := strconv.ParseInt(string(bs[1:]), 10, 64)
		if err != nil {
			log.Printf("listener reader error: failed to read number of bytes: %s", err.Error())
			return
		}
		buf := make([]byte, numBytes)
		readCount, err := conn.Read(buf)
		if err != nil {
			log.Printf("listener read error: failed reading buffer of size %d: %s", numBytes, err.Error())
		}
		if readCount != int(numBytes) {
			log.Printf("listener read error: readCount %d did not match expected number of bytes %d", readCount, numBytes)
			return
		}
		go handleListenerCommand(bs[0], buf)
	}
}

func handleListenerCommand(c byte, buf []byte) {
	switch c {
	case 's':
		handleListenerSpawn(buf)
	case 'o':
		handleListenerObject(buf)
	case 'e':
		handleListenerError(buf)
	}
}

func handleListenerSpawn(buf []byte) {
	// Essentially in here we want to:
	// decode and execute the spawn (make that function public and properly handle)
	// THEN somehow get that pid back to the connection that requested this, maybe by just writing to it
	// and in spawn func we just read to block and handle similar to the object listener handler below
	log.Printf("handleListenerSpawn")
}
func handleListenerObject(buf []byte) {
	log.Printf("handleListenerObject")
}
func handleListenerError(buf []byte) {
	log.Printf("handleListenerError")
}

func StartNodeConnection(nodeNameAndAddr string) error {
	colonIndex := strings.Index(nodeNameAndAddr, ":")
	if colonIndex == -1 {
		return fmt.Errorf("`:` not found in node name and address")
	}
	nodeName := nodeNameAndAddr[:colonIndex]
	addr := nodeNameAndAddr[colonIndex:]
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	cnc := &ConnAndCh{
		Conn:    conn,
		WriteCh: make(chan Payload),
	}
	_, alreadyLoaded := NodeNameToConnection.LoadAndStore(nodeName, cnc)
	if alreadyLoaded {
		return fmt.Errorf("connection already found for node %s", nodeNameAndAddr)
	}
	go handleWritesToConnectionChannel(cnc)
	return nil
}

var delimeter = []byte{'\r'}

func handleWritesToConnectionChannel(cnc *ConnAndCh) {
	for {
		x := <-cnc.WriteCh
		count, err := cnc.Conn.Write(x.InitialPacket)
		if err != nil {
			log.Printf("failed on write for x.InitialPacket: error: %s", err.Error())
			break
		}
		if len(x.InitialPacket) != count {
			log.Printf("failed on write, count for if x.InitialPacket did not match, got=%d", count)
			break
		}
		count, err = cnc.Conn.Write(delimeter)
		if err != nil {
			log.Printf("failed on write for delimeter: error: %s", err.Error())
			break
		}
		if len(delimeter) != count {
			log.Printf("failed on write, count for if delimeter did not match, got=%d", count)
			break
		}
		count, err = cnc.Conn.Write(x.FullPacket)
		if err != nil {
			log.Printf("failed on write for x.FullPacket: error: %s", err.Error())
			break
		}
		if len(x.FullPacket) != count {
			log.Printf("failed on write, count for if x.FullPacket did not match, got=%d", count)
			break
		}
	}
	cnc.WriteCh = nil
}
