package main

/*
references:
- https://code.google.com/p/whispering-gophers/source/browse/master/main.go
- https://gist.github.com/spikebike/2232102
*/

import (
	// "container/list"
	"bufio"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"helper"
	"log"
	"net"
	"net/http"
	"os"
	// "strings"
	// "time"
	"encoding/json"
	"flag"
	"sync"
)

var (
	initialize = flag.Bool("init", false, "is this the initial node?")
	myAddr     = flag.String("ip", "", "your public ip address")
	permission rune
	self       string
)

type Hub struct {
	messages map[string]chan<- Message
	mu       sync.RWMutex
}

type Message struct {
	Sender string
	Body   string
}

var hub = &Hub{messages: make(map[string]chan<- Message)}

func main() {
	//specify initialization with cmdline arg for now
	flag.Parse()
	//configure TLS
	cert, err := tls.LoadX509KeyPair("certs/server.pem", "certs/server.key")
	if err != nil {
		log.Fatal(err)
	}
	config := tls.Config{Certificates: []tls.Certificate{cert}}
	config.Rand = rand.Reader
	// listener, err := net.Listen("tcp", ":3000")
	listener, err := tls.Listen("tcp", ":3000", &config)
	if err != nil {
		log.Fatal(err)
	}
	self = listener.Addr().String()
	log.Printf("listening on self=%s\n", self)
	if *initialize {
		knownAddrs := []string{helper.ELNUX2, helper.ME, helper.EC2}
		for _, addr := range knownAddrs {
			go dial(addr)
		}
	}
	go readInput()
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Fatal(err)
			}
			go serve(conn)
		}
	}()
	http.HandleFunc("/jsclient", func(w http.ResponseWriter, r *http.Request) {
		msg := r.URL.Query()["msg"][0]
		log.Printf("got jsclient message: %s\n", msg)
		m := Message{Sender: self, Body: msg}
		broadcast(m)
	})
	err2 := http.ListenAndServe("localhost:3001", nil)
	if err2 != nil {
		panic("browserListener(): error")
	}
	log.Println("listening for jsclient on port 3001")
}

func dial(addr string) {
	if addr == self {
		return //dont dial self
	}
	ch := hub.Add(addr)
	if ch == nil {
		return //peer already connected
	}
	defer hub.Remove(addr)

	//configure tls
	cert, err := tls.LoadX509KeyPair("certs/client.pem", "certs/client.key")
	if err != nil {
		log.Fatal(err)
	}
	config := tls.Config{Certificates: []tls.Certificate{cert}}
	config.Rand = rand.Reader

	//try to connect
	log.Printf("dialing peer@%s\n", addr)
	conn, err := tls.Dial("tcp", addr, &config)
	if err != nil {
		log.Printf("dial error: %s", err)
		//TODO add retries
		return
	}
	log.Printf("connected to peer@%s\n", addr)
	defer func() {
		conn.Close()
		log.Printf("close connection to peer@%s\n", addr)
	}()
	enc := json.NewEncoder(conn)
	for m := range ch {
		err := enc.Encode(m)
		if err != nil {
			log.Printf("some err peer@%s\n", addr)
			return
		}
	}
}

func readInput() {
	reader := bufio.NewReader(os.Stdin)
	for {
		s, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		m := Message{Sender: self, Body: s[:len(s)-1]}
		broadcast(m)
	}
}

func broadcast(msg Message) {
	for _, ch := range hub.List() {
		select {
		case ch <- msg:
		default:
			//okay to drop messages sometimes?
		}
	}
}

func serve(c net.Conn) {
	log.Println("<", c.RemoteAddr(), "accepted connection")
	d := json.NewDecoder(c)
	for {
		var m Message
		err := d.Decode(&m)
		if err != nil {
			log.Println("<", c.RemoteAddr(), "error:", err)
			break
		}
		log.Printf("< %v received: %v", c.RemoteAddr(), m)
		fmt.Println(m.Body)
		broadcast(m)
		go dial(m.Sender)
	}
	c.Close()
	log.Println("<", c.RemoteAddr(), "close")

}

func (h *Hub) Add(addr string) <-chan Message {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.messages[addr]; ok {
		return nil
	}
	ch := make(chan Message)
	h.messages[addr] = ch
	return ch
}

func (h *Hub) Remove(addr string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.messages, addr)
}

func (h *Hub) List() []chan<- Message {
	h.mu.RLock()
	defer h.mu.RUnlock()
	l := make([]chan<- Message, 0, len(h.messages))
	for _, ch := range h.messages {
		l = append(l, ch)
	}
	return l
}
