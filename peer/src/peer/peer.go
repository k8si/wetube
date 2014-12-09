package main

/*
references:
- https://code.google.com/p/whispering-gophers/source/browse/master/main.go
- https://gist.github.com/spikebike/2232102
*/

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"helper"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var (
	myAddr      = flag.String("ip", "", "your public ip address") //TODO this is just "self"
	interactive = flag.Bool("interactive", false, "interactive mode")
	permission  = flag.Int("permission", 2, "permission [0=DIR|1=EDIT|2=VIEW]")
	self        string
	nodeID      int
	privkey     *rsa.PrivateKey
	pubkey      *rsa.PublicKey
)

var hub = &Hub{peers: make(map[string]chan<- Message)}

func main() {
	//specify initialization with cmdline arg for now
	flag.Parse()
	self = *myAddr

	//configure TLS
	cert, err := tls.LoadX509KeyPair("server_cert.pem", "server_key.pem")
	if err != nil {
		log.Fatal(err)
	}
	config := tls.Config{Certificates: []tls.Certificate{cert}}
	config.Rand = rand.Reader

	//listen on port 3000 for incoming connections
	listener, err := tls.Listen("tcp", ":3000", &config)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("listening on self=%s\n", self)

	//load and set public, private keys
	raw, err := ioutil.ReadFile("keyout.der") //DER encoded version of server_key.pem because that's what worked....
	if err != nil {
		log.Fatal(err)
	}
	pk, err := x509.ParsePKCS1PrivateKey(raw)
	if err != nil {
		log.Fatal("error parsing private key: ", err)
	}
	privkey = pk

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Fatal(err)
			}
			go serve(conn)
		}
	}()

	if *permission == helper.DIRECTOR {
		takeOffice()
	}

	if *interactive {
		readInputStdin()
	} else {
		readInput()
	}
}

func invitePeer(addr string, perm string, done chan int) {
	log.SetPrefix("invite: ")
	if addr == self {
		done <- 0
		return
	}
	fmt.Println("inviting ", addr)
	check, err := strconv.Atoi(perm)
	if err != nil || check > helper.VIEWER {
		log.Fatal("bad permission in invitees.txt")
	}
	dialed := make(chan int)
	go dial(addr, dialed)
	if <-dialed == 0 {
		newID := nodeidreg.getNewID()
		newIDstr := strconv.Itoa(newID)
		welcome := Message{ID: helper.RandomID(), Sender: self, Subject: "welcome", Body: addr + "," + newIDstr + "," + perm}
		broadcast(welcome)
		done <- 0
		return
	}
	done <- 1
}

/* listen for messages from gui */
func readInput() {
	http.HandleFunc("/jsclient", func(w http.ResponseWriter, r *http.Request) {
		msg := r.URL.Query()["msg"][0]
		// log.Printf("got jsclient message: %s\n", msg)
		m := Message{ID: helper.RandomID(), Sender: self, Subject: "msg", Body: msg}
		if *permission < helper.VIEWER {
			broadcast(m)
		}
	})
	http.HandleFunc("/inv", func(w http.ResponseWriter, r *http.Request) {
		msg := r.URL.Query()
		// log.Printf("got jsclient message: %s\n", msg)
		if *permission == 0 {
			addr := msg["invite"][0]
			perm := msg["perm"][0]
			log.Println("inviting peer @ ", addr, "with permission", perm)
			done := make(chan int)
			go invitePeer(addr, perm, done)
			<-done
		}
	})
	err2 := http.ListenAndServe("localhost:3001", nil)
	if err2 != nil {
		panic("browserListener(): error")
	}
}

/* "interactive mode" -- listen for messages from stdin (for debugging/fun)
the legal messages are:
	msg [msg] -- sends a message to the connected peers (see incoming.go)
	list -- lists all connected peers
	dirs -- lists all connected directors
*/
func readInputStdin() {
	fmt.Println("\n\t\t\t *** INTERACTIVE MODE *** \n")
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

/*
func readInvitees(done chan []string) {
	log.SetPrefix("invite: ")
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
*/
