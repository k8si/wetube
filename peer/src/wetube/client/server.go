package main

//assume this is the initial node in the network

import (
	// "bufio"
	"code.google.com/p/go.net/websocket"
	"fmt"
	"log"
	"net"
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
	fmt.Println("handleConnect....")
	// // fmt.Println("doing initLocalSocket()....")
	var msg string
	err := websocket.Message.Receive(ws, &msg)
	if err != nil {
		fmt.Println("ProcessSocket: got error", err)
		_ = websocket.Message.Send(ws, "FAIL:"+err.Error())
		return
	}
	fmt.Println("success! received message: ", msg)
	fmt.Println("message: remote addr: ", ws.RemoteAddr(), "; local addr: ", ws.LocalAddr())
	myself = Peer{ipaddr: msg, port: "3000", wid: rune(numConnected)}
	websocket.Message.Send(ws, "ACK! --"+myself.ipaddr)
	// //inc numConnected (as I am now connected)
	// numConnected += 1
	// //if len(peers) == 0 (I am the init director) { invitePeers }
	// // else demote my permissions
	// // if len(peers) == 0 {
	// // invitePeers()
	// // }

}

func foo(dest string, conn *net.TCPConn) {
	fmt.Println("foo")
	origin := myself.ipaddr
	fmt.Println("want to dial remote addr:", dest, " from origin:", origin)
	ws, err := websocket.Dial(dest, "", origin)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := ws.Write([]byte("hello world\n")); err != nil {
		log.Fatal(err)
	}
	var msg = make([]byte, 512)
	var n int
	if n, err = ws.Read(msg); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Received %s.\n", msg[:n])
}

func (h *hub) run() {
	for {
		select {
		case c := <-h.register:
			h.connections[c] = true
		case c := <-h.unregister:
			if _, ok := h.connections[c]; ok {
				delete(h.connections, c)
				close(c.send)
			}
		case m := <-h.broadcast:
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
