package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	// "net/url"
	"os"
	// "strconv"
	// "time"
)

const (
	TCP_PORT = "3000"
	ACK      = "1"
	DIRECTOR = 0
	EDITOR   = 1
	WATCHER  = 2
)

var (
	knownAddrs = []string{"54.149.51.58", "174.62.219.8"} //"54.149.39.226",
	myipaddr   string
	permission rune
	initialize = false
)

// func RunClient() {
func main() {
	//specify initialization with cmdline arg
	args := os.Args
	if len(args) == 2 {
		myipaddr = args[1]
	} else if len(args) == 3 {
		myipaddr = args[1]
		initialize = true
		permission = 0
	} else {
		panic("not enough args. usage: go client.go ip-addr [init]")
	}

	fmt.Println("starting client. permission lvl=", permission, "; ipaddr=", myipaddr)
	go h.run()
	//invite some peers to get the stew going
	if initialize {
		for _, addr := range knownAddrs {
			if addr == myipaddr {
				continue
			}
			go invite(addr)
		}
	}
	// for c, _ := range h.connections {
	// 	go c.writer()
	// 	go c.reader()
	// 	// defer func() { h.unregister <- c }()
	// }

	service := ":" + TCP_PORT
	fmt.Println("client listening on ", TCP_PORT)
	//doing this via http because i'm lazy / the routes are convenient
	http.HandleFunc("/jsclient", connectToBrowser) //handle incoming messages from js-client
	http.HandleFunc("/invite", handleInvite)       //handle invitations to connect
	s := &http.Server{
		Addr: service,
	}
	// err := http.ListenAndServe(service, nil)
	err := s.ListenAndServe()
	if err != nil {
		panic(err.Error())
	}
}

// func invite(addr string, conn chan *connection) {
func invite(addr string) {
	fmt.Println("trying to invite peer @ ", addr)
	req := "http://" + addr + ":3000/invite?perm=1"
	res, err := http.Get(req)
	if err != nil {
		//TODO wait for awhile then retry
		panic(err.Error())
	}
	if res.StatusCode != 200 {
		panic("bad status code: " + string(res.StatusCode))
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp", addr+":"+TCP_PORT)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("tcpaddr ", tcpAddr, "resolved.")
	tcpConn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("connection established.")
	tcpConn.Write([]byte("welcome"))
	s := make(chan []byte)
	c := &connection{socket: tcpConn, send: s}
	h.register <- c
	go c.reader()
	go c.writer()
}

func handleRequest(conn net.Conn) {
	buf := make([]byte, 1024)
	_, err := conn.Read(buf)
	if err != nil {
		log.Fatal(err)
	}
	m := string(buf)
	fmt.Println("rcv'ed message: ", m)
	conn.Write([]byte(ACK))
	conn.Close()
}

func handleInvite(w http.ResponseWriter, r *http.Request) {
	fmt.Println("rcv'd invite req: ", r)

}

//TODO need to do something with the msg's
func connectToBrowser(w http.ResponseWriter, r *http.Request) {
	fmt.Println("rcv'd jsclient req: ", r)
	q := r.URL.Query()
	fmt.Println("msg = ", q["msg"])
	s := q["msg"][0]
	fmt.Println("want to broadcast: ", s)
	b := []byte(s)
	// ms := []byte(m)
	h.broadcast <- b
}

type hub struct {
	connections map[*connection]bool
	broadcast   chan []byte
	register    chan *connection
	unregister  chan *connection
}

type connection struct {
	socket *net.TCPConn
	send   chan []byte
}

var h = hub{
	broadcast:   make(chan []byte),
	register:    make(chan *connection),
	unregister:  make(chan *connection),
	connections: make(map[*connection]bool),
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
	fmt.Println("starting writer() for ", c.socket.RemoteAddr())
	for message := range c.send {
		fmt.Println("writing message: ", string(message))
		_, err := c.socket.Write(message)
		if err != nil {
			break
		}
	}
	c.socket.Close()
	fmt.Println("closed writer() for ", c.socket.RemoteAddr())

}

func (c *connection) reader() {
	fmt.Println("starting reader() for ", c.socket.RemoteAddr())
	for {
		msg := make([]byte, 1024)
		_, err := c.socket.Read(msg)
		if err != nil {
			panic(err.Error())
		}
		fmt.Println("got msg: ", string(msg))
		h.broadcast <- msg
	}
	c.socket.Close()
	fmt.Println("closed reader() for ", c.socket.RemoteAddr())

}
