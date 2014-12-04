package main

import (
	"fmt"
	"io"
	// "log"
	"crypto/rand"
	"crypto/tls"
	"net"
	// "newmarch"
	"net/http/httputil"
	"os"
	"time"
)

const (
	TCP_PORT = "3000"
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

type hub struct {
	connections map[*connection]bool
	broadcast   chan []byte
	register    chan *connection
	unregister  chan *connection
}

var h = hub{
	broadcast:   make(chan []byte),
	register:    make(chan *connection),
	unregister:  make(chan *connection),
	connections: make(map[*connection]bool),
}

type connection struct {
	// socket *net.TCPConn
	socket *httputil.ClientConn
	send   chan []byte
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
			fmt.Println("rcvd broadcast.")
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
	// fmt.Println("starting writer() for ", c.socket.RemoteAddr())
	// for message := range c.send {
	// 	fmt.Println("writing message: ", string(message))
	// 	_, err := c.socket.Write(message)
	// 	if err != nil {
	// 		break
	// 	}
	// }
	// c.socket.Close()
	// fmt.Println("closed writer() for ", c.socket.RemoteAddr())
}

func (c *connection) reader() {
	// fmt.Println("starting reader() for ", c.socket.RemoteAddr())
	// for {
	// 	msg := make([]byte, 1024)
	// 	_, err := c.socket.Read(msg)
	// 	if err != nil {
	// 		panic(err.Error())
	// 	}
	// 	fmt.Println("got msg: ", string(msg))
	// 	h.broadcast <- msg
	// }
	// c.socket.Close()
	// fmt.Println("closed reader() for ", c.socket.RemoteAddr())
}

func main() {
	// newmarch.GenRSAKeys()
	// newmarch.GenX509Cert()
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

	go h.run()

	cert, err := tls.LoadX509KeyPair("jan.newmarch.name.pem", "private.pem")
	checkError(err)
	config := tls.Config{Certificates: []tls.Certificate{cert}}
	now := time.Now()
	config.Time = func() time.Time { return now }
	config.Rand = rand.Reader
	service := ":" + TCP_PORT
	listener, err := tls.Listen("tcp", service, &config)
	checkError(err)
	fmt.Println("listening on port ", TCP_PORT)
	defer listener.Close()

	if initialize {
		for _, addr := range knownAddrs {
			if addr == myipaddr {
				continue
			}
			c := make(chan net.Conn)
			go invite(addr, c)
			cf := &tls.Config{Rand: rand.Reader}
			ssl := tls.Client(<-c, cf)
			s := make(chan []byte)
			thing := httputil.NewClientConn(ssl, nil)
			conn := &connection{socket: thing, send: s}
			h.register <- conn
		}
	}

	for {
		//wait for a connection
		conn, err := listener.Accept()
		if err != nil {
			panic(err.Error())
		}
		go func(c net.Conn) {
			fmt.Println("got message.")
			io.Copy(c, c)
			c.Close()
		}(conn)
	}
}

func checkError(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func invite(addr string, c chan net.Conn) {
	fmt.Println("inviting peer @ ", addr)
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr+":"+TCP_PORT)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("tcpaddr ", tcpAddr, "resolved.")
	tcpConn := tryInvite(tcpAddr)
	if tcpConn == nil {
		for {
			fmt.Println("retrying...")
			tcpConn := tryInvite(tcpAddr)
			if tcpConn != nil {
				break
			}
			waitFor := 10
			time.Sleep(time.Duration(waitFor) * time.Second)
			continue
		}
	}
	c <- tcpConn
}

func tryInvite(tcpAddr *net.TCPAddr) *net.TCPConn {
	tcpConn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil
	}
	return tcpConn
}
