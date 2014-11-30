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
	http.Handle("/websocket/", websocket.Handler(initBrowser2ClientSocket))
	err := http.ListenAndServe(service, nil)
	checkError(err)
}

func checkError(err error) {
	if err != nil {
		fmt.Println("fatal error: ", err.Error())
		os.Exit(1)
	}
}

func initBrowser2ClientSocket(ws *websocket.Conn) {
	fmt.Println("doing initLocalSocket()....")
	var msg string
	err := websocket.Message.Receive(ws, &msg)
	if err != nil {
		fmt.Println("ProcessSocket: got error", err)
		_ = websocket.Message.Send(ws, "FAIL:"+err.Error())
		return
	}
	fmt.Println("go message: ", msg)
	myself = Peer{ipaddr: msg, port: "3000", wid: rune(numConnected)}
	websocket.Message.Send(ws, "ACK! --local")
	//inc numConnected (as I am now connected)
	numConnected += 1
	//if len(peers) == 0 (I am the init director) { invitePeers }
	// else demote my permissions
	if len(peers) == 0 {
		invitePeers()
	}

}

func invitePeers() {
	for _, addr := range knownAddrs {
		if addr == myself.ipaddr {
			fmt.Println(addr, " is me!")
			continue
		}
		fmt.Println("inviting ", addr, "....")
		//send invite and wait for accept response, then establish socket connection
		//assume invitee has socket server listening on port 3000
		tcpAddr, err := net.ResolveTCPAddr("tcp", addr+":3000")
		if err != nil {
			fmt.Println("ResolveTCPAddr error", err.Error())
			return
		}
		conn, err := net.DialTCP("tcp", nil, tcpAddr)
		if err != nil {
			fmt.Println("DialTCP error", err.Error())
			return
		}
		fmt.Println("got connection for peer @", addr)
		foo(addr, conn)
	}
}

func foo(dest string, conn *net.TCPConn) {
	fmt.Println("foo")
	origin := myself.ipaddr
	fmt.Println("want to dial remote addr:", dest, " from origin:", origin)
	ws, err := websocket.Dial("ws://"+dest+":3000/ws", "", "http://"+origin)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := ws.Write([]byte(dest)); err != nil {
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
