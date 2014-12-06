package main

import (
	"container/list"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"helper"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	myAddr          string
	permission      rune
	initialize      = false
	registeredAddrs = list.New()
)

// var h = helper.Hub{
// 	Broadcast:   make(chan []byte),
// 	Register:    make(chan *helper.Connection),
// 	Unregister:  make(chan *helper.Connection),
// 	Connections: make(map[*helper.Connection]bool),
// 	Addrs:       list.New(),
// }

var h = hub{
	broadcast:   make(chan []byte),
	register:    make(chan *connection),
	unregister:  make(chan *connection),
	connections: make(map[*connection]bool),
}

func main() {
	//specify initialization with cmdline arg
	args := os.Args
	if len(args) == 2 {
		myAddr = args[1]
	} else if len(args) == 3 {
		myAddr = args[1]
		initialize = true
		permission = helper.DIRECTOR
	} else {
		panic("not enough args. usage: go peer.go ip-addr [init]")
	}

	//start listening for GUI interactions
	go browserListener()

	//start the hub
	go h.run()

	//make TLS configuration
	cert, err := tls.LoadX509KeyPair("certs/server.pem", "certs/server.key")
	checkError(err, "main: loadkeys")
	serverConfig := tls.Config{Certificates: []tls.Certificate{cert}}
	serverConfig.Rand = rand.Reader
	//start listening on port 3000 for incoming connections
	service := ":" + helper.TCP_PORT
	listener, err := tls.Listen("tcp", service, &serverConfig)
	checkError(err, "main: error listening")
	fmt.Println("peer listening for incoming connections on port 3000")
	defer listener.Close()

	if initialize {
		log.Println("inviting peers...")
		knownAddrs := []string{helper.ME, helper.ELNUX2, helper.EC2}
		cert, err := tls.LoadX509KeyPair("certs/client.pem", "certs/client.key")
		checkError(err, "invitePeers: load keys")
		clientConfig := tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
		for _, addr := range knownAddrs {
			if addr == myAddr {
				continue
			}
			// output := make(chan *tls.Conn)
			go sendInvite(addr, clientConfig)
			// newSocket := <-output
			//register the new connection
			// go registerAndListen(newSocket)
		}
		log.Println("done inviting peers.")
	}

	for {
		//wait for an incoming connection
		conn, err := listener.Accept()
		checkError(err, "main: listener: accept")
		log.Printf("accepted incoming connection from [Peer @ %s]", conn.RemoteAddr())
		go handleIncoming(conn)
	}
}

func handleIncoming(conn net.Conn) {
	defer conn.Close()
	// incomingAddr := strings.Split(conn.RemoteAddr().String(), ":")
	//read the message
	// for {
	buf := make([]byte, 512)
	n, err := conn.Read(buf)
	checkError(err, "handleIncoming: error reading")
	response := string(buf[:n])
	log.Printf("handleIncoming(): [Peer @ %s] sent message: %s", conn.RemoteAddr(), response)
	go route(response)
	// log.Printf("handleIncoming(): writing ACK to [Peer @ %s]", conn.RemoteAddr())
	// responseMsg := "ACK"
	// _, err = conn.Write([]byte(responseMsg))
	// checkError(err, "handleIncoming: error writing ACK")
	// break
	// }
	log.Printf("handleIncoming(): closed connection to [Peer @ %s]\n", conn.RemoteAddr())

	// //check if we've already registered an outgoing connection to this peer
	// found := false
	// for e := registeredAddrs.Front(); e != nil; e = e.Next() {
	// 	if e.Value.(string) == incomingAddr {
	// 		found = true
	// 		break
	// 	}
	// }
	// if found {
	// 	log.Printf("already have outgoing connection to [Peer @ %s]", incomingAddr)
	// } else {
	// 	//send "ACK" message
	// 	ack := []byte("ACK")
	// 	_, err := conn.Write(ack)
	// }
}

func listConnections() {
	//print currently registered connections for debugging
	//FIXME this causes a runtime error (invalid memory address/nil pointer dereference) ONLY on the Edlab ????
	log.Printf("route(): currently connected to %d peers\n:", len(h.connections))
	for c, _ := range h.connections {
		if c == nil {
			log.Fatalf("wtf?")
		}
		log.Printf("\t[Peer @ %s]\n", c.socket.RemoteAddr().String())
	}
}

func route(message string) {
	parts := strings.Split(message, "#")
	subject := parts[0]
	contents := parts[1]
	if subject == "HIFROM" {
		//contents should be the IP address of the peer who wants to connect
		log.Printf("route(): got HIFROM %s\n", contents)
		peerAddr := contents + ":" + helper.TCP_PORT

		listConnections()

		//check if we've already registered an outgoing connection to this peer
		found := false
		for c := range h.connections {
			thisAddr := c.socket.RemoteAddr().String()
			if thisAddr == peerAddr {
				log.Printf("route(): already registered address %s\n", peerAddr)
				found = true
				break
			}
		}
		if !found {
			log.Printf("route(): address %s not found, trying to connect...", peerAddr)
			cert, err := tls.LoadX509KeyPair("certs/client.pem", "certs/client.key")
			checkError(err, "route: load keys")
			clientConfig := tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
			// output := make(chan *tls.Conn)
			go sendInvite(contents, clientConfig)
			// newSocket := <-output
			//register the new connection
			// go registerAndListen(newSocket)
		}
	} else if subject == "NEWPEER" {
		log.Printf("route(): got NEWPEER @ %s\n", contents)
		// peerAddr := contents + ":" + helper.TCP_PORT

		listConnections()

		// //check if we've already registered an outgoing connection to this peer
		// found := false
		// for c := range h.Connections {
		// 	thisAddr := c.Socket.RemoteAddr().String()
		// 	if thisAddr == peerAddr {
		// 		log.Printf("route(): already registered address %s\n", peerAddr)
		// 		found = true
		// 		break
		// 	}
		// }
		// if !found {
		// 	log.Printf("route(): address %s not found, trying to connect...", peerAddr)
		// 	cert, err := tls.LoadX509KeyPair("certs/client.pem", "certs/client.key")
		// 	checkError(err, "route: load keys")
		// 	clientConfig := tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
		// 	// output := make(chan *tls.Conn)
		// 	go sendInvite(contents, clientConfig)
		// 	// newSocket := <-output
		// 	//register the new connection
		// 	// go registerAndListen(newSocket)
		// }
	} else {
		log.Fatalf("received invalid message: %s %s", subject, contents)
	}
}

// func invitePeers() {
// 	knownAddrs := []string{helper.ME, helper.ELNUX2, helper.EC2}
// 	cert, err := tls.LoadX509KeyPair("certs/client.pem", "certs/client.key")
// 	checkError(err, "invitePeers: load keys")
// 	clientConfig := tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
// 	for _, addr := range knownAddrs {
// 		if addr == myAddr {
// 			continue
// 		}
// 		// output := make(chan *tls.Conn)
// 		go sendInvite(addr, clientConfig)
// 		// newSocket := <-output
// 		//register the new connection
// 		// go registerAndListen(newSocket)
// 	}
// }

func registerAndListen(socket *tls.Conn) {
	log.Printf("registerAndListen()ing to [Peer @ %s]\n", socket.RemoteAddr())
	send := make(chan []byte, 1024)
	newConn := &connection{socket: socket, send: send}
	h.register <- newConn
	// defer func() { h.Unregister <- newConn }()
	go newConn.writer()
	// output := make(chan *helper.Message)
	// output := make(chan []byte, 1024)
	go newConn.reader()
	// h.Broadcast <- <-output
	// go newConn.Reader()
	//message := <-output
}

func sendInvite(addr string, config tls.Config) {
	log.Printf("sendInvite() to [Peer @ %s]\n", addr)
	peerAddr := addr + ":" + helper.TCP_PORT
	connectOutput := make(chan *tls.Conn)
	go connect(peerAddr, config, connectOutput)
	peerConn := <-connectOutput
	// peerConn := connect(peerAddr, config)
	// defer peerConn.Close()
	inviteMsg := "HIFROM#" + myAddr
	n, err := peerConn.Write([]byte(inviteMsg))
	checkError(err, "sendInvite: write error")
	log.Printf("wrote HIFROM (%d bytes)", n)

	//now register and listen to the new connection
	go registerAndListen(peerConn)

	// output <- peerConn

	// if string(ack) == "ACK" {
	// 	output <- peerConn
	// } else {
	// 	panic("no ACK received")
	// }
}

func listenForAck(peerConn *tls.Conn) {
	for {
		buf := make([]byte, 512)
		n, err := peerConn.Read(buf)
		checkError(err, "sendInvite: error in response")
		response := string(buf[:n])
		log.Printf("received response from [Peer @ %s]: %s\n", peerConn.RemoteAddr(), response)
		if response == "ACK" {
			// registerConnection(peerConn)
			break
		}
	}
}

func connect(addr string, config tls.Config, output chan *tls.Conn) {
	log.Printf("connect(): trying to connect to [Peer @ %s]...", addr)
	//this will end up calling "handleIncoming", i.e. peer will receive incoming connection from myAddr:[some random port]
	conn, err := tls.Dial("tcp", addr, &config)
	if err != nil {
		numRetries := 0
		for {
			log.Printf("connect() to %s: failed. #retries=%d...", addr, numRetries)
			conn, err = tls.Dial("tcp", addr, &config)
			if err == nil && conn != nil {
				break
			}
			numRetries += 1
			waitFor := 10
			time.Sleep(time.Duration(waitFor) * time.Second)
			continue
		}
	}
	log.Printf("success!\n")
	output <- conn
}

func browserListener() {
	http.HandleFunc("/jsclient", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("browserListener(): got new message from GUI")
		q := r.URL.Query()
		msg := q["msg"][0]
		fmt.Println("got jsclient message", msg)
		broadcast := "VID" + helper.DELIM + msg
		h.broadcast <- []byte(broadcast)
	})
	err := http.ListenAndServe(":3001", nil)
	fmt.Println("listening on port 3001")
	checkError(err, "browserListener: listener error")
}

func checkError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s : %s", msg, err)
	}
}

type hub struct {
	connections map[*connection]bool
	broadcast   chan []byte
	register    chan *connection
	unregister  chan *connection
}

type connection struct {
	socket *tls.Conn
	send   chan []byte
}

func (h *hub) run() {
	for {
		select {
		case c := <-h.register:
			h.connections[c] = true
		case c := <-h.unregister:
			if _, ok := h.connections[c]; ok {
				delete(h.connections, c)
				close(c.send)
			}
		case m := <-h.broadcast:
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
	fmt.Println("starting writer()")
	for message := range c.send {
		fmt.Println("writing message: ", string(message))
		_, err := c.socket.Write(message)
		checkError(err, "connection.writer(): write error")
	}
	c.socket.Close()
	fmt.Println("closed writer()")
}

func (c *connection) reader() {
	fmt.Println("starting reader()")
	for {
		msg := make([]byte, 1024)
		_, err := c.socket.Read(msg)
		if err != nil {
			break
		}
		// checkError(err, "connection.reader(): read error")
		// fmt.Println("got msg: ", string(msg))
		h.broadcast <- msg
	}
	c.socket.Close()
	fmt.Println("closed reader()")
}
