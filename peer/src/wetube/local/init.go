package main

/*
TODO
- should be able to just run this code on local, EC2 (instead of separating init.go and server.go)
- how to set node ids? do we need node ids?
- how to set permission lvls?
*/

//assume this is the initial node in the network

import (
	// "bufio"
	"fmt"
	"golang.org/x/net/websocket"
	"log"
	// "net"
	"net/http"
	// "os"
	// "strconv"
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
	ipaddr     string
	diraddr    string
	nodeId     rune
	permission rune
}

var (
	peers        = []Peer{}
	knownAddrs   = []string{"54.149.39.226", "174.62.219.8"}
	myself       Peer
	permission   rune
	numConnected = 0
)

func main() {
	//for now, get permission of cmdline
	// args := os.Args
	// if len(args) < 2 {
	// 	os.Exit(1)
	// }
	// p, _ := strconv.ParseInt(args[1], 10, 32)
	// permission = rune(p)
	// fmt.Println("starting node with permission: ", permission)
	service := ":3000"
	fmt.Println("listening on ", service)
	go h.run()
	// http.Handle("/websocket/", websocket.Handler(initBrowser2ClientSocket))
	// http.Handle("/ws", websocket.Handler(initBrowser2ClientSocket))
	http.Handle("/jscli", websocket.Handler(initBrowser2ClientSocket))
	http.HandleFunc("/invite", inviteHandler)
	err := http.ListenAndServe(service, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func inviteHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Println("rcv'd invite to connection from ", req.RemoteAddr())
}

func initBrowser2ClientSocket(ws *websocket.Conn) {
	fmt.Println("doing initBrowser2ClientSocket()....")
	var msg string
	err := websocket.Message.Receive(ws, &msg)
	if err != nil {
		_ = websocket.Message.Send(ws, "FAIL:"+err.Error())
		log.Fatal(err)
	}
	fmt.Println("got message: ", msg)
	websocket.Message.Send(ws, "1")
	npeers := len(h.connections)
	fmt.Println("currently connected to ", npeers, "peers.")
	// conn := make(chan *connection)
	// if npeers == 0 && permission == DIRECTOR {
	// 	myself := Peer{ipaddr: msg, diraddr: msg, nodeId: 0, permission: DIRECTOR}
	// 	fmt.Println("inviting peers to join...")
	// 	for _, addr := range knownAddrs {
	// 		if addr != myself.ipaddr {
	// 			go invitePeer(addr, conn)
	// 			go registerAndListen(conn)
	// 		}
	// 	}
	// }
	//listen for new messages from browser client
	listen(ws)
}

func invitePeer(addr string, conn chan *connection) {
	fmt.Println("trying to invite peer @ ", addr)
	origin := "http://" + myself.ipaddr
	dest := "ws://" + addr + ":3000/invite"
	fmt.Println("dialing...")
	ws, err := websocket.Dial(dest, "", origin)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("success!")
	conn <- &connection{send: make(chan []byte, 256), ws: ws}
}

func registerAndListen(conn chan *connection) {
	fmt.Println("registerAndListening....")
	c := <-conn
	addr := c.ws.RemoteAddr()
	fmt.Println("connected to ", addr)
	h.register <- c
	go listen(c.ws)
	go c.writer()
	// defer func() { h.unregister <- c }()
}

func respondToInvitation(ws *websocket.Conn) {
	fmt.Println("respond to invitation...")
	if _, err := ws.Write([]byte("HI")); err != nil {
		log.Fatal(err)
	}
	conn := make(chan *connection)
	conn <- &connection{send: make(chan []byte, 256), ws: ws}
	registerAndListen(conn)
	fmt.Println("done.")
}

func listen(ws *websocket.Conn) {
	//now wait for new messages
	fmt.Println("listening for new messages...")
	for {
		var m string
		err := websocket.Message.Receive(ws, &m)
		if err != nil {
			fmt.Println("got err!", err.Error())
			break
		}
		fmt.Println("got new message: ", m)
		h.broadcast <- []byte(m)
	}
	fmt.Println("done listening.")
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

// func invitePeers() []connection {
// 	connections := []connection{}
// 	for _, addr := range knownAddrs {
// 		if addr == myself.ipaddr {
// 			// fmt.Println(addr, " is me!")
// 			continue
// 		}
// 		fmt.Println("inviting ", addr, "....")
// 		//send invite and wait for accept response, then establish socket connection
// 		//assume invitee has socket server listening on port 3000
// 		tcpAddr, err := net.ResolveTCPAddr("tcp", addr+":3000")
// 		if err != nil {
// 			fmt.Println("ResolveTCPAddr error", err.Error())
// 			log.Fatal(err)
// 		}
// 		conn, err := net.DialTCP("tcp", nil, tcpAddr)
// 		if err != nil {
// 			fmt.Println("DialTCP error", err.Error())
// 			log.Fatal(err)
// 		}
// 		fmt.Println("got connection for peer @", addr)
// 		c := handshake(addr, conn)
// 		connections = append(connections, c)
// 	}
// 	return connections
// }

// //set up the initial connection between this node and the peer
// func handshake(dest string, conn *net.TCPConn) connection {
// 	fmt.Println("trying handshake....")

// 	origin := myself.ipaddr

// 	if dest == origin {
// 		c := connection{send: make(chan []byte, 256), ws: nil}
// 		return c
// 	}

// 	fmt.Println("want to dial remote addr:", dest, " from origin:", origin)

// 	//try and contact the peer
// 	ws, err := websocket.Dial("ws://"+dest+":3000/ws", "", "http://"+origin)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	//send the peer its public ip address
// 	if _, err := ws.Write([]byte(dest)); err != nil {
// 		log.Fatal(err)
// 	}
// 	var ack1 = make([]byte, 512)
// 	var n1 int
// 	if n1, err = ws.Read(ack1); err != nil {
// 		log.Fatal(err)
// 	}
// 	fmt.Printf("Received %s.\n", ack1[:n1])

// 	//send the peer my public ip address
// 	if _, err := ws.Write([]byte(origin)); err != nil {
// 		log.Fatal(err)
// 	}
// 	var ack2 = make([]byte, 512)
// 	var n2 int
// 	if n2, err = ws.Read(ack2); err != nil {
// 		log.Fatal(err)
// 	}
// 	fmt.Printf("Received %s.\n", ack2[:n2])

// 	//TODO send the peer their permission lvl?
// 	//TODO send all of this in a single msg?
// 	// or ping --> ACK --> send info --> ACK

// 	c := connection{send: make(chan []byte, 256), ws: ws}
// 	return c
// }
