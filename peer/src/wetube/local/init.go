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
	websocket.Message.Send(ws, "1")
	//inc numConnected (as I am now connected)
	numConnected += 1
	//if len(peers) == 0 (I am the init director) { invitePeers }
	// else demote my permissions
	if len(peers) == 0 {
		connections := invitePeers()
		num := len(connections)
		fmt.Println("est. ", num, "connections")
		for _, c := range connections {
			h.register <- &c
			go listen(c.ws)
			go c.writer()
			defer func() { h.unregister <- &c }()
		}
	}
	//now wait for new instructions from js-client
	listen(ws)
}

func invitePeers() []connection {
	connections := []connection{}
	for _, addr := range knownAddrs {
		if addr == myself.ipaddr {
			// fmt.Println(addr, " is me!")
			continue
		}
		fmt.Println("inviting ", addr, "....")
		//send invite and wait for accept response, then establish socket connection
		//assume invitee has socket server listening on port 3000
		tcpAddr, err := net.ResolveTCPAddr("tcp", addr+":3000")
		if err != nil {
			fmt.Println("ResolveTCPAddr error", err.Error())
			log.Fatal(err)
		}
		conn, err := net.DialTCP("tcp", nil, tcpAddr)
		if err != nil {
			fmt.Println("DialTCP error", err.Error())
			log.Fatal(err)
		}
		fmt.Println("got connection for peer @", addr)
		c := handshake(addr, conn)
		connections = append(connections, c)
	}
	return connections
}

//set up the initial connection between this node and the peer
func handshake(dest string, conn *net.TCPConn) connection {
	fmt.Println("trying handshake....")

	origin := myself.ipaddr
	fmt.Println("want to dial remote addr:", dest, " from origin:", origin)

	//try and contact the peer
	ws, err := websocket.Dial("ws://"+dest+":3000/ws", "", "http://"+origin)
	if err != nil {
		log.Fatal(err)
	}

	//send the peer its public ip address
	if _, err := ws.Write([]byte(dest)); err != nil {
		log.Fatal(err)
	}
	var ack1 = make([]byte, 512)
	var n1 int
	if n1, err = ws.Read(ack1); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Received %s.\n", ack1[:n1])

	//send the peer my public ip address
	if _, err := ws.Write([]byte(origin)); err != nil {
		log.Fatal(err)
	}
	var ack2 = make([]byte, 512)
	var n2 int
	if n2, err = ws.Read(ack2); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Received %s.\n", ack2[:n2])
	c := connection{send: make(chan []byte, 256), ws: ws}
	return c
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

func (c *connection) writer() {
	fmt.Println("c.writer()...")
	for message := range c.send {
		fmt.Println("writing message: ", string(message))
		_, err := c.ws.Write(message)
		if err != nil {
			break
		}
	}
	c.ws.Close()
}

// func (c *connection) reader() {
// 	for {
// 		_, message, err := c.ws.ReadMessage()
// 		if err != nil {
// 			break
// 		}
// 		h.broadcast <- message
// 	}
// 	c.ws.Close()
// }
