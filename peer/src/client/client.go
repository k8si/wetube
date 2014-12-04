package main

import (
	"fmt"
	"io"
	// "log"
	// "crypto/tls"
	"net"
	"newmarch"
	"os"
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

func main() {
	newmarch.GenRSAKeys()
	newmarch.GenX509Cert()
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

	if initialize {
		for _, addr := range knownAddrs {
			if addr == myipaddr {
				continue
			}
			go invite(addr)
		}
	}

	service := ":" + TCP_PORT
	listener, err := net.Listen("tcp", service)
	fmt.Println("listening on port ", TCP_PORT)
	if err != nil {
		panic(err.Error())
	}
	defer listener.Close()
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

func invite(addr string) {
	fmt.Println("inviting peer @ ", addr)
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr+":"+TCP_PORT)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("tcpaddr ", tcpAddr, "resolved.")
	tcpConn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		panic(err.Error())
	}
	// cf := &tls.Config{Rand: rand.Reader}
	// ssl := tls.Client(tcpConn, cf)
	// c.clientConn = httputil.NewClientConn(ssl, nil)
	// req, err := http.NewRequest("GET", c.path.String(), nil)
	// resp, err := c.clientConn.Do(req)

	fmt.Println("connection established.")
	msg := "invite"
	tcpConn.Write([]byte(msg))
	tcpConn.Close()
}
