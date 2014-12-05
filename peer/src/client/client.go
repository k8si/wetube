package main

import (
	"fmt"
	// "io"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"log"
	"net"
	// "net/http/httputil"
	"helper"
	"net/http"
	"os"
	// "time"
)

const (
	TCP_PORT = "3000"
	DIRECTOR = 0
	EDITOR   = 1
	WATCHER  = 2
)

var (
	myipaddr   string
	permission rune
	initialize = false
)

var h = helper.Hub{
	Broadcast:   make(chan []byte),
	Register:    make(chan *helper.Connection),
	Unregister:  make(chan *helper.Connection),
	Connections: make(map[*helper.Connection]bool),
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

	go browserListener()

	go h.Run()

	//load server certs for TLS
	cert, err := tls.LoadX509KeyPair("certs/server.pem", "certs/server.key")
	checkError(err)
	config := tls.Config{Certificates: []tls.Certificate{cert}}
	config.Rand = rand.Reader
	service := ":" + TCP_PORT
	listener, err := tls.Listen("tcp", service, &config)
	checkError(err)
	fmt.Println("server listening on port ", TCP_PORT)
	defer listener.Close()

	if initialize {
		fmt.Println("inviting peers...")
		invitePeers()
	}

	for {
		//wait for a connection
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("server: accept: %s", err)
			break
		}
		// checkError(err)
		// go handleConnection(conn)
		log.Printf("server: accepted from %s", conn.RemoteAddr())
		tlscon, ok := conn.(*tls.Conn)
		if ok {
			log.Print("ok=true")
			state := tlscon.ConnectionState()
			for _, v := range state.PeerCertificates {
				log.Print(x509.MarshalPKIXPublicKey(v.PublicKey))
			}
		}
		//TODO need chan to store connection so we can defer unregister
		go handleClient(conn)
	}

	// for _, c := range h.Connections {
	// 	// fmt.Println("closing connection to ", c.Socket.RemoteAddr())
	// 	h.Unregister <- c
	// }
}

func handleClient(conn net.Conn) {
	s := make(chan []byte)
	tlsconn := conn.(*tls.Conn)
	c := &helper.Connection{Socket: tlsconn, Send: s}
	h.Register <- c
	go c.Writer()
	go c.Reader()
	// c.send <- []byte("ACK")
}

// func handleClient(conn net.Conn) {
// // defer conn.Close()
// buf := make([]byte, 512)
// for {
// 	log.Print("server: conn: waiting")
// 	n, err := conn.Read(buf)
// 	if err != nil {
// 		log.Printf("server: conn: read: %s", err)
// 		break
// 	}
// 	log.Printf("server: conn: got message %q\n", string(buf[:n]))
// 	// n, err = conn.Write(buf[:n])
// 	// log.Printf("sever: conn: wrote %d bytes", n)
// 	// if err != nil {
// 	// 	log.Printf("sever: write: %s", err)
// 	// 	break
// 	// }
// }
// log.Println("server: conn: closed")
// }

func handleConnection(conn net.Conn) {
	fmt.Println("got connection")
	// req := make([]byte, 1024)
	// sent, err := conn.Read(req)
	// checkError(err)
	// fmt.Println("msg: ", sent)
	// // c := make(chan net.Conn)
	// cert, err := tls.LoadX509KeyPair("jan.newmarch.name.pem", "private.pem")
	// checkError(err)
	// cf := &tls.Config{Rand: rand.Reader, Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true} //ServerName: "jan.newmarch.name"}
	// ssl := tls.Server(conn, cf)
	// s := make(chan []byte, 256)
	// // thing := httputil.NewClientConn(ssl, nil)
	// // newconn := &connection{socket: thing, send: s}
	// newconn := &connection{socket: ssl, send: s}
	// h.register <- newconn
	// // msg := "ACK"
	// go newconn.writer()
	// go newconn.reader()
	// // h.broadcast <- []byte(msg)
	// // newconn.send <- []byte(msg)

}

func onInviteReceived() {

}

func browserListener() {
	http.HandleFunc("/jsclient", func(w http.ResponseWriter, r *http.Request) {
		// fmt.Printf(w, "hello")
		fmt.Println("got jsclient message")
		msg := myipaddr + ":jsclientmsg"
		h.Broadcast <- []byte(msg)
	})
	err := http.ListenAndServe(":3001", nil)
	fmt.Println("listening on port 3001")
	checkError(err)
}

func checkError(err error) {
	if err != nil {
		panic(err.Error())
	}
}
