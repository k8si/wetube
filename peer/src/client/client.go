package main

import (
	"fmt"
	"log"
	"net"
	// "net/http"
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
	knownAddrs = []string{"54.149.39.226", "174.62.219.8"}
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
			conn := make(chan *connection)
			go invite(addr, conn)
		}
	}

	service := ":" + TCP_PORT
	fmt.Println("client listening on ", TCP_PORT)
	l, err := net.Listen("tcp", service)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	fmt.Println("listening on port ", TCP_PORT)
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("got request")
		go handleRequest(conn)
	}
}

func invite(addr string, conn chan *connection) {
	fmt.Println("trying to invite peer @ ", addr)
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr+":"+TCP_PORT)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("tcpaddr: ", tcpAddr)

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

type hub struct {
	connections map[*connection]bool
	broadcast   chan []byte
	register    chan *connection
	unregister  chan *connection
}

type connection struct {
	ws   *net.TCPConn
	send chan []byte
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
		_, err := c.ws.Write(message)
		if err != nil {
			break
		}
	}
	c.ws.Close()
}
