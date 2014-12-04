package main

import (
	"fmt"
	// "io"
	// "log"
	"crypto/rand"
	"crypto/tls"
	"net"
	// "newmarch"
	// "net/http/httputil"
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
	knownAddrs = []string{"128.119.243.175", "54.149.118.210", "128.119.40.193"}
	myipaddr   string
	permission rune
	initialize = false
)

var h = hub{
	broadcast:   make(chan []byte),
	register:    make(chan *connection),
	unregister:  make(chan *connection),
	connections: make(map[*connection]bool),
}

func main() {
	//specify initialization with cmdline arg
	args := os.Args
	fmt.Println("args: ", args)
	if len(args) == 2 {
		fmt.Println("normal peer")
		myipaddr = args[1]
	} else if len(args) == 3 {
		fmt.Println("got initialize, peer = DIR")
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
		fmt.Println("inviting peers...")
		invitePeers()
	}

	for {
		//wait for a connection
		conn, err := listener.Accept()
		if err != nil {
			panic(err.Error())
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	fmt.Println("got connection")
	req := make([]byte, 1024)
	sent, err := conn.Read(req)
	checkError(err)
	fmt.Println("msg: ", sent)

	// c := make(chan net.Conn)
	cf := &tls.Config{Rand: rand.Reader}
	ssl := tls.Client(conn, cf)
	s := make(chan []byte, 256)
	// thing := httputil.NewClientConn(ssl, nil)
	// newconn := &connection{socket: thing, send: s}
	newconn := &connection{socket: ssl, send: s}
	h.register <- newconn
	msg := "ACK"
	go newconn.writer()
	go newconn.reader()
	// h.broadcast <- []byte(msg)
	newconn.send <- []byte(msg)

}

func onInviteReceived() {

}

func checkError(err error) {
	if err != nil {
		panic(err.Error())
	}
}
