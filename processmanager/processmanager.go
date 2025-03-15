package processmanager

// import (
// 	"fmt"
// 	"log"
// 	"net"

// 	"github.com/lxzan/gws"
// 	kcp "github.com/xtaci/kcp-go"
// )

// // Listener

// // Listen for connections coming from other nodes

// // Connector

// // Allow a node (this blue process) to connect to another node (another blue process)

// // TODO: fix up logging, cant use fatal as it will exit the process
// type Command int

// const (
// 	Connect Command = iota
// 	AddConnection
// 	DataRead
// 	DataWrite
// 	ListConnections
// )

// type ConnectorCommand struct {
// 	Command Command
// 	Value   any
// }

// type DataWriteValue struct {
// 	Identifier string
// 	Data       []byte
// }

// func (cc ConnectorCommand) String() string {
// 	return fmt.Sprintf("ConnectorCommand{Command: %d, Value: %#+v}", cc.Command, cc.Value)
// }

// var connectorCommand = make(chan ConnectorCommand, 1)

// func Start(address string) {
// 	go startListener2(address)
// 	// go startListener(address)
// 	// go startConnector()
// }

// func SendCommand(cmd ConnectorCommand) {
// 	// TODO: Need to make sure this is a non-blocking send
// 	connectorCommand <- cmd
// }

// func SendListConnectionsCommand() {
// 	SendCommand(ConnectorCommand{Command: ListConnections, Value: struct{}{}})
// }

// func startListener2(address string) {
// 	listener, err := kcp.Listen(address)
// 	if err != nil {
// 		log.Println(err.Error())
// 		return
// 	}
// 	app := gws.NewServer(&gws.BuiltinEventHandler{}, nil)
// 	app.RunListener(listener)
// }

// func startClient2(address string) {
// 	conn, err := kcp.Dial(address)
// 	if err != nil {
// 		log.Println(err.Error())
// 		return
// 	}
// 	app, _, err := gws.NewClientFromConn(&gws.BuiltinEventHandler{}, nil, conn)
// 	if err != nil {
// 		log.Println(err.Error())
// 		return
// 	}
// 	app.ReadLoop()
// }

// func startListener(address string) {
// 	listener, err := net.Listen("tcp", address)
// 	if err != nil {
// 		log.Fatalf("Failed to start Node Listener: %s", err.Error())
// 	}
// 	log.Printf("Node Listener Listening on: %s", listener.Addr().String())
// 	defer listener.Close()
// 	// if we get a connection from someone
// 	// this means they are connecting to our node
// 	// then we want to add this connection to a somewhere so that we can try to read or write to it
// 	for {
// 		c, err := listener.Accept()
// 		if err != nil {
// 			log.Printf("Failed to accept connection here: %s", err.Error())
// 			continue
// 		}
// 		go handleConnection(c)
// 	}
// }

// func handleConnection(c net.Conn) {
// 	for {
// 		// _, err := c.Read()
// 	}
// }

// // func startConnector() {
// // 	for {
// // 		select {
// // 		case cmd, ok := <-connectorCommand:
// // 			if !ok {
// // 				// Channel is closed, exit the goroutine
// // 				log.Println("Channel closed, exiting connector.")
// // 				return
// // 			}
// // 			// Process the received message
// // 			go executeConnectorCommand(cmd)
// // 		default:
// // 			// Do other work or just sleep to avoid busy waiting
// // 			time.Sleep(10 * time.Microsecond)
// // 		}
// // 	}
// // }

// // func executeConnectorCommand(cmd ConnectorCommand) {
// // 	log.Printf("executeConnectorCommand: " + cmd.String())
// // 	switch cmd.Command {
// // 	case Connect:
// // 		handleConnectCommand(cmd.Value)
// // 	case AddConnection:
// // 		handleAddConnection(cmd.Value)
// // 	case DataRead:
// // 		handleDataRead(cmd.Value)
// // 	case DataWrite:
// // 		handleDataWrite(cmd.Value)
// // 	case ListConnections:
// // 		handleListConnections(cmd.Value)
// // 	default:
// // 		log.Printf("Unhandled connector command: %s", cmd.Command)
// // 	}
// // }

// // func handleConnectCommand(cmdValue any) {
// // 	c, err := net.Dial("tcp", cmdValue.(string))
// // 	if err != nil {
// // 		log.Fatalf("failed to handle connect command: %s", err.Error())
// // 	}
// // 	// defer c.Close()
// // 	SendCommand(ConnectorCommand{Command: AddConnection, Value: c})
// // 	// What im thinking is that we want to add this to some store here
// // 	// then support a command for sending and a command for receiving
// // }

// // type connectionsState struct {
// // 	cs   []net.Conn
// // 	lock sync.RWMutex
// // }

// // var connections = connectionsState{
// // 	cs:   make([]net.Conn, 0),
// // 	lock: sync.RWMutex{},
// // }

// // func handleAddConnection(cmdValue any) {
// // 	c, ok := cmdValue.(net.Conn)
// // 	if !ok {
// // 		log.Printf("cmdValue was not net.Conn, got=%T", cmdValue)
// // 		return
// // 	}
// // 	connections.lock.Lock()
// // 	defer connections.lock.Unlock()
// // 	connections.cs = append(connections.cs, c)
// // 	// TODO: Want a read and writer of connection (and need to handle commands to send to particular commands)
// // 	reader := func(c net.Conn) {
// // 		var buf = make([]byte, 1024)
// // 		for {
// // 			_, err := c.Read(buf)
// // 			if err != nil {
// // 				log.Printf("closing connection on error: %s", err.Error())
// // 				c.Close()
// // 				break
// // 			}
// // 			SendCommand(ConnectorCommand{Command: DataRead, Value: buf})
// // 		}
// // 	}
// // 	go reader(c)
// // 	// writer := func(c net.Conn) {}
// // }

// // func handleDataRead(cmdValue any) {
// // 	buf := cmdValue.([]byte)
// // 	log.Printf("received: %s", string(buf))
// // }

// // func handleDataWrite(cmdValue any) {
// // 	dwv := cmdValue.(DataWriteValue)
// // 	connections.lock.Lock()
// // 	defer connections.lock.Unlock()
// // 	for _, c := range connections.cs {
// // 		if c.RemoteAddr().String() == dwv.Identifier {
// // 			log.Printf("Found Connection")
// // 			_, err := c.Write(dwv.Data)
// // 			if err != nil {
// // 				log.Fatalf("failed to write: %s", err.Error())
// // 			}
// // 		}
// // 	}
// // }

// // func handleListConnections(_ any) {
// // 	connections.lock.Lock()
// // 	defer connections.lock.Unlock()
// // 	for _, c := range connections.cs {
// // 		fmt.Printf("c (local) = %s, (remote) = %s", c.LocalAddr().String(), c.RemoteAddr().String())
// // 	}
// // }
