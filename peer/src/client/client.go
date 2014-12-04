package main

import (
	"fmt"
	"io"
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
	//"54.149.39.226", "54.149.51.58", "174.62.219.8"
	knownAddrs = []string{"54.149.39.226", "174.62.219.8"}
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
	// newmarch.GenRSAKeys()
	// newmarch.GenX509Cert()
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
