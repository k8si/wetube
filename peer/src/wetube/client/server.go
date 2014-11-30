package main

//assume this is the initial node in the network

import (
	// "bufio"
	"code.google.com/p/go.net/websocket"
	"fmt"
	"log"
	// "net"
	"net/http"
	"os"
)

type hub struct {
	connections map[*connection]bool
	broadcast   chan []byte
	register    chan *connection
	unregister  chan *connection
}

type connection struct {
	ws   *websocket.Conn
	send chan []byte
}

var h = hub{
	broadcast:   make(chan []byte),
	register:    make(chan *connection),
	unregister:  make(chan *connection),
	connections: make(map[*connection]bool),
}

const (
	DIRECTOR = 0
	EDITOR   = 1
	VIEWER   = 2
)

type Peer struct {
	ipaddr string
	port   string
	wid    rune
}

var (
	peers        = []Peer{}
	knownAddrs   = []string{"54.149.39.226", "24.128.54.88"}
	myself       Peer
	permission   = DIRECTOR
	numConnected = 0
)

func main() {
	service := ":3000"
	fmt.Println("listening on ", service)
	go h.run()
	http.Handle("/ws", websocket.Handler(handleConnect))
	err := http.ListenAndServe(service, nil)
	checkError(err)
}

func checkError(err error) {
	if err != nil {
		fmt.Println("fatal error: ", err.Error())
		os.Exit(1)
	}
}

func handleConnect(ws *websocket.Conn) {
	fmt.Println("handleConnect...")
	var myIP string
	err := websocket.Message.Receive(ws, &myIP)
	if err != nil {
		_ = websocket.Message.Send(ws, "FAIL:"+err.Error())
		log.Fatal(err)
	}
	fmt.Println("success! my ip addr:", myIP)
	myself = Peer{ipaddr: myIP, port: "3000", wid: rune(numConnected)}
	peers = append(peers, myself)
	websocket.Message.Send(ws, "1")
	numConnected += 1

	var otherIP string
	err = websocket.Message.Receive(ws, &otherIP)
	if err != nil {
		_ = websocket.Message.Send(ws, "FAIL:"+err.Error())
		log.Fatal(err)
	}
	fmt.Println("success! other ip addr: ", otherIP)
	other := Peer{ipaddr: otherIP, port: "3000", wid: rune(numConnected)}
	peers = append(peers, other)
	websocket.Message.Send(ws, "1")

	fmt.Println("connected peers:")
	for _, p := range peers {
		fmt.Println(p.ipaddr)
	}

	c := &connection{send: make(chan []byte, 256), ws: ws}
	h.register <- c
	defer func() { h.unregister <- c }()
	go listen(c.ws)

	// //now wait for new messages
	// for {
	// 	var m string
	// 	err = websocket.Message.Receive(ws, &m)
	// 	if err != nil {
	// 		break
	// 	}
	// 	fmt.Println("got new message: ", m)
	// 	// h.broadcast <- []byte(m)
	// }
}

func listen(ws *websocket.Conn) {
	//now wait for new messages
	for {
		var m string
		err := websocket.Message.Receive(ws, &m)
		if err != nil {
			break
		}
		fmt.Println("got new message: ", m)
		h.broadcast <- []byte(m)
	}
}

func (h *hub) run() {
	for {
		select {
		case c := <-h.register:
			fmt.Println("registering new connection.")
			h.connections[c] = true
		case c := <-h.unregister:
			fmt.Println("unregistering connection")
			if _, ok := h.connections[c]; ok {
				delete(h.connections, c)
				close(c.send)
			}
		case m := <-h.broadcast:
			npeers := len(h.connections)
			fmt.Println("rcvd broadcast. sending to ", npeers, "peers...")
			for c := range h.connections {
				select {
				case c.send <- m:
				default:
					delete(h.connections, c)
					close(c.send)
				}
			}
		}
	}
}

// func foo(dest string, conn *net.TCPConn) {
// 	fmt.Println("foo")
// 	origin := myself.ipaddr
// 	fmt.Println("want to dial remote addr:", dest, " from origin:", origin)
// 	ws, err := websocket.Dial(dest, "", origin)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	if _, err := ws.Write([]byte("hello world\n")); err != nil {
// 		log.Fatal(err)
// 	}
// 	var msg = make([]byte, 512)
// 	var n int
// 	if n, err = ws.Read(msg); err != nil {
// 		log.Fatal(err)
// 	}
// 	fmt.Printf("Received %s.\n", msg[:n])
// }

// func (h *hub) run() {
// 	for {
// 		select {
// 		case c := <-h.register:
// 			h.connections[c] = true
// 		case c := <-h.unregister:
// 			if _, ok := h.connections[c]; ok {
// 				delete(h.connections, c)
// 				close(c.send)
// 			}
// 		case m := <-h.broadcast:
// 			for c := range h.connections {
// 				select {
// 				case c.send <- m:
// 				default:
// 					delete(h.connections, c)
// 					close(c.send)
// 				}
// 			}
// 		}
// 	}

// }

// package main

// import (
// 	"bytes"
// 	"fmt"
// 	"net"
// 	"os"
// 	"strconv"
// )

// const (
// 	CONN_HOST = ""
// 	CONN_PORT = "3000"
// 	CONN_TYPE = "tcp"
// )

// func main() {
// 	// a := net.LocalAddr()
// 	// fmt.Println(a)

// 	l, err := net.Listen(CONN_TYPE, ":"+CONN_PORT)

// 	if err != nil {
// 		fmt.Println("error listening ", err.Error())
// 		os.Exit(-1)
// 	}
// 	a := l.Addr()
// 	fmt.Println(a)
// 	defer l.Close()
// 	fmt.Println("listening on " + CONN_HOST + ":" + CONN_PORT)
// 	for {
// 		conn, err := l.Accept()
// 		if err != nil {
// 			fmt.Println("error accepting: ", err.Error())
// 			os.Exit(-1)
// 		}
// 		fmt.Printf("Received message %s -> %s \n", conn.RemoteAddr(), conn.LocalAddr())
// 		go handleRequest(conn)
// 	}
// }

// func handleRequest(conn net.Conn) {
// 	buf := make([]byte, 1024)
// 	reqlen, err := conn.Read(buf)
// 	if err != nil {
// 		fmt.Println("error reading:", err.Error())
// 	}
// 	message := "receved msg: " + strconv.Itoa(reqlen) + "bytes"
// 	n := bytes.Index(buf, []byte{0})
// 	message += string(buf[:n-1])
// 	conn.Write([]byte(message))
// 	conn.Close()
// }
