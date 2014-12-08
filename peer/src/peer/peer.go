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
- election procedure probably wont survive multiple directors (e.g. multiple peers will be assigned ID "1")
- IN GENERAL need to deal with having multiple directors
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
	"flag"
	"fmt"
	"helper"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	// "sync"
)

var (
	// initialize   = flag.Bool("init", false, "is this the initial node?") //TODO no longer used
	myAddr      = flag.String("ip", "", "your public ip address") //TODO this is just "self"
	interactive = flag.Bool("i", false, "interactive mode")
	permission  = flag.Int("perm", 2, "permission [0=DIR|1=EDIT|2=VIEW")
	self        string
	nodeID      int
)

var hub = &Hub{peers: make(map[string]chan<- Message)}

func main() {
	//specify initialization with cmdline arg for now
	flag.Parse()
	self = *myAddr

	//configure TLS
	cert, err := tls.LoadX509KeyPair("cacert.pem", "id_rsa")
	if err != nil {
		log.Fatal(err)
	}
	config := tls.Config{Certificates: []tls.Certificate{cert}}
	config.Rand = rand.Reader

	//listen on port 3000 for incoming connections
	// service := self + ":3000"
	listener, err := tls.Listen("tcp", ":3000", &config)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("listening on self=%s\n", self)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Fatal(err)
			}
			go serve(conn)
		}
	}()

	//if we're the director, invite peers in the file "invitees.txt"
	//TODO there are probably much better way(s)
	if *permission == helper.DIRECTOR {
		takeOffice()
		done := make(chan []string)
		go readInvitees(done)
		<-done
		log.Printf("*** done inviting peers. connected to %d. ***\n", hub.Size())
		// for _, a := range <-done {
		// 	if a != "" {
		// 		fmt.Println("invited ", a)
		// 	}
		// }
	}

	if *interactive {
		readInputStdin()
	} else {
		readInput()
	}
}

func readInvitees(done chan []string) {
	//map of {address: permission} read from "invitees.txt"
	invited := make(map[string]chan int)
	numpeers := 0
	f, err := os.Open("invitees.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), " ")
		if len(parts) != 2 {
			log.Fatal("bad line in invitees.txt")
		}
		addr := parts[0]
		numpeers += 1
		ch := make(chan int)
		invited[addr] = ch
		go invitePeer(addr, parts[1], ch)
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	addrs := make([]string, numpeers)
	for a, v := range invited {
		if <-v == 0 {
			addrs = append(addrs, a)
		} else {
			addrs = append(addrs, "")
		}
	}
	done <- addrs

}

func invitePeer(addr string, perm string, done chan int) {
	if addr == self {
		done <- 0
	}
	fmt.Println("inviting ", addr)
	check, err := strconv.Atoi(perm)
	if err != nil || check > helper.VIEWER {
		log.Fatal("bad permission in invitees.txt")
	}
	dialed := make(chan int)
	// welcomed := make(chan int)
	go dial(addr, dialed)
	if <-dialed == 0 {
		newID := nodeidreg.getNewID()
		newIDstr := strconv.Itoa(newID)
		//TODO send the new node its permission too
		welcome := Message{ID: helper.RandomID(), Sender: self, Subject: "welcome", Body: addr + "," + newIDstr + "," + perm}
		broadcast(welcome)
		done <- 0
	}
	done <- 1
}

/* send messages to gui */
func sendToGui(msg string) {
	fmt.Println("sending ", msg, "to gui")
	req := "http://localhost:4000/input?msg=" + msg
	_, err := http.Get(req)
	if err != nil {
		log.Println(err)
	}
}

/* listen for messages from gui */
func readInput() {
	http.HandleFunc("/jsclient", func(w http.ResponseWriter, r *http.Request) {
		msg := r.URL.Query()["msg"][0]
		log.Printf("got jsclient message: %s\n", msg)
		m := Message{ID: helper.RandomID(), Sender: self, Subject: "msg", Body: msg}
		if *permission < helper.VIEWER {
			broadcast(m)
		}
	})
	err2 := http.ListenAndServe("localhost:3001", nil)
	if err2 != nil {
		panic("browserListener(): error")
	}
}

/* "interactive mode" */
func readInputStdin() {
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
				fmt.Printf("currently %d peers connected:\n", len(hub.List()))
				hub.PrintAll()
			} else if parts[0] == "dirs" {
				printDirectors()
				// fmt.Printf("currently %d directors:\n", len(directorAddrs))
				// for _, a := range directorAddrs {
				// 	fmt.Printf("\t\t%s\n", a)
				// }
			} else {
				log.Println("readInput: invalid input: input must be of form [subject]#[body]")
			}
			continue
		}
		m := Message{
			ID:      helper.RandomID(),
			Sender:  self,
			Subject: parts[0],
			Body:    parts[1],
		}
		//only directors/editors can send new instructions
		if *permission < helper.VIEWER {
			broadcast(m)
			seen(m.ID)
		} else {
			log.Println("readInput: you dont have permission.")
		}
	}
}
