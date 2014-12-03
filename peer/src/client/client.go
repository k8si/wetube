package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	// "net/url"
	"os"
	"strconv"
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
)

// func RunClient() {
func main() {
	//for now set permission via cmdline arg
	args := os.Args
	if len(args) < 3 {
		panic("not enough args")
	}
	p, _ := strconv.ParseInt(args[1], 10, 32)
	permission = rune(p)
	myipaddr = args[2]
	fmt.Println("starting client. permission lvl=", permission, "; ipaddr=", myipaddr)
	go h.run()
	if permission == DIRECTOR {
		//initialization; need to invite peers
		for _, addr := range knownAddrs {
			if addr == myipaddr {
				continue
			}
			// conn := make(chan *connection)
			// go invite(addr, conn)
			go invite(addr)
		}
	}
	for c, _ := range h.connections {
		go c.writer()
		go c.reader()
		// defer func() { h.unregister <- c }()
	}

	service := ":" + TCP_PORT
	fmt.Println("client listening on ", TCP_PORT)
	http.HandleFunc("/jsclient", connectToBrowser)
	http.HandleFunc("/invite", handleInvite)
	err := http.ListenAndServe(service, nil)
	if err != nil {
		panic(err.Error())
	}
	// l, err := net.Listen("tcp", service)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer l.Close()
	// fmt.Println("listening on port ", TCP_PORT)
	// for {
	// 	conn, err := l.Accept()
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	go handleRequest(conn)
	// }
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
	s := make(chan []byte)
	c := &connection{socket: tcpConn, send: s}
	h.register <- c
	// inviteMsg := "hello?"
	// tcpConn.Write([]byte(inviteMsg))
	// tcpConn.Close()

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
	fmt.Println("c.writer()...")
	for message := range c.send {
		fmt.Println("writing message: ", string(message))
		_, err := c.socket.Write(message)
		if err != nil {
			break
		}
	}
	c.socket.Close()
}

func (c *connection) reader() {
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
}
