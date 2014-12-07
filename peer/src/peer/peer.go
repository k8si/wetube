package main

/*
references:
- https://code.google.com/p/whispering-gophers/source/browse/master/main.go
- https://gist.github.com/spikebike/2232102
*/

/*
TODO
- connection endpoints that quit aren't removed until another message is broadcast (more or less) (this is probably a problem for director elections)
- voting for new director has deadlock/race condition bugs

TODO (longterm)
- sending permissions
- digital signatures
- deal with InsecureSkipVerify
- Youtube API
- tests
*/

import (
	"bufio"
	"crypto/rand"
	"crypto/tls"
	"helper"
	"log"
	// "net/http"
	"flag"
	"os"
	"strconv"
	"strings"
	"sync"
)

var (
	initialize   = flag.Bool("init", false, "is this the initial node?") //TODO no longer used
	myAddr       = flag.String("ip", "", "your public ip address")       //TODO this is just "self"
	permission   = flag.Int("perm", 2, "permission [0=DIR|1=EDIT|2=VIEW")
	self         string
	directorAddr string
	nodeID       int
)

var (
	hub       = &Hub{peers: make(map[string]chan<- Message)}
	welcome   Message
	nodeidreg NodeIDRegistry
)

func main() {
	//specify initialization with cmdline arg for now
	flag.Parse()

	//configure TLS
	cert, err := tls.LoadX509KeyPair("cacert.pem", "id_rsa")
	if err != nil {
		log.Fatal(err)
	}
	config := tls.Config{Certificates: []tls.Certificate{cert}}
	config.Rand = rand.Reader
	listener, err := tls.Listen("tcp", ":3000", &config)
	if err != nil {
		log.Fatal(err)
	}

	self = *myAddr
	if *permission == helper.DIRECTOR {
		takeOffice()
	}

	log.Printf("listening on self=%s\n", self)

	// if *initialize {
	// 	knownAddrs := []string{helper.ELNUX2, helper.ME, helper.EC2, helper.ELNUX7}
	// 	for _, addr := range knownAddrs {
	// 		go dial(addr)
	// 	}
	// }

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Fatal(err)
			}
			go serve(conn)
		}
	}()

	// http.HandleFunc("/jsclient", func(w http.ResponseWriter, r *http.Request) {
	// 	msg := r.URL.Query()["msg"][0]
	// 	log.Printf("got jsclient message: %s\n", msg)
	// 	m := Message{Sender: self, Subject: "", Body: msg}
	// 	broadcast(m)
	// })
	// err2 := http.ListenAndServe("localhost:3001", nil)
	// if err2 != nil {
	// 	panic("browserListener(): error")
	// }

	readInput()
}

func readInput() {
	r := bufio.NewReader(os.Stdin)
	for {
		s, err := r.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		//parse the user's input
		input := s[:len(s)-1]
		parts := strings.Split(input, helper.MSG_DELIM)
		if len(parts) < 2 {
			if parts[0] == "list" {
				hub.PrintAll()
			} else {
				log.Println("readInput: invalid input: input must be of form [subject]#[body]")
			}
			continue
		}
		// log.Println("readInput: got client message: subject=", parts[0], ", body=", parts[1])
		m := Message{
			ID:      helper.RandomID(),
			Sender:  self,
			Subject: parts[0],
			Body:    parts[1],
		}
		//only directors/editors can send new instructions
		if *permission < helper.VIEWER {
			broadcast(m)
			if *permission == helper.DIRECTOR && m.Subject == "invite" {
				done := make(chan int)
				go dial(m.Body, done)
				if <-done == 0 {
					newID := nodeidreg.getNewID()
					newIDstr := strconv.Itoa(newID)
					welcome.Body = directorAddr + "," + newIDstr
					broadcast(welcome)
				}
			}
			seen(m.ID)
		} else {
			log.Println("readInput: you dont have permission.")
		}
	}
}

func broadcast(msg Message) {
	// log.Println("broadcasting message from Sender", msg.Sender, ": ", msg.Subject, msg.Body)
	for _, ch := range hub.List() {
		select {
		case ch <- msg:
		default:
			//okay to drop messages sometimes?
		}
	}
}

/* Message Log */
var seenMessages = struct {
	m map[string]bool
	sync.Mutex
}{m: make(map[string]bool)}

func seen(id string) bool {
	seenMessages.Lock()
	ok := seenMessages.m[id]
	seenMessages.m[id] = true
	seenMessages.Unlock()
	return ok
}
