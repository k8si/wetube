package main

import (
	"encoding/json"
	"fmt"
	"helper"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

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

func serve(c net.Conn) {
	log.SetPrefix("serve: ")
	fmt.Printf("(< %s) serve: accepted connection.\n", c.RemoteAddr())

	sendPing()
	d := json.NewDecoder(c)
	for {
		var m Message
		err := d.Decode(&m)

		if err != nil {
			log.Printf("(< %s) serve: error: %s\n", c.RemoteAddr(), err)
			break
		}
		//ignore messages from myself
		if m.Sender == self {
			continue
		}
		//ignore messages already seen
		if seen(m.ID) {
			continue
		}

		fmt.Printf("(< %v) serve: RCVD: id=%s sender=%s subj=%s body=%s\n", c.RemoteAddr(), m.ID, m.Sender, m.Subject, m.Body)

		verified := m.verify()
		if !verified {
			fmt.Println("failed to verify message")
			continue
		}

		go dial(m.Sender, nil)

		/* ROUTES */
		switch m.Subject {
		case "ping":
			addr := m.Body
			if addr == "" || len(addr) == 0 || addr == self {
				log.Printf("bad ping: %s\n", addr)
			} else {
				fmt.Printf("*** got ping for %s ***\n", m.Body)
				log.Printf("good ping: %s. about to dial.", addr)
			}

		//message requesting some info about me
		case "request":
			if m.Body == "id" {
				idstr := strconv.Itoa(nodeID)
				response := Message{Sender: self, ID: helper.RandomID(), Subject: "response", Body: self + "," + idstr}
				broadcast(response)
				fmt.Printf("(< %s) serve: broadcasted response\n", c.RemoteAddr())
			}

		//response to my request for info
		case "response":
			fmt.Printf("(< %s) serve: received response\n", c.RemoteAddr())
			parts := strings.Split(m.Body, ",")
			idstr := parts[1]
			id, err := strconv.Atoi(idstr)
			if err != nil {
				log.Fatal("bad nodeid received")
			}
			nodeIDs.Lock()
			nodeIDs.m[parts[0]] = id
			nodeIDs.Unlock()
			n := len(hub.List())
			if len(nodeIDs.m) == n {
				allReceived <- true
			}

		case "vote":
			fmt.Printf("(< %s) serve: received vote\n", c.RemoteAddr())
			parts := strings.Split(m.Body, ",")
			idstr := parts[1]
			id, err := strconv.Atoi(idstr)
			if err != nil {
				log.Fatal("bad nodeid received")
			}
			votes.Lock()
			votes.m[parts[0]] = id
			votes.Unlock()
			n := len(hub.List())
			if len(votes.m) == n {
				allVotesReceived <- true
			}

		case "msg":
			broadcast(m)
			sendToGui(m.Body)

		case "welcome", "newdirector":
			theresANewDirector(m)

		default:
			log.Fatalf("(< %s) serve: invalid message: %s %s %s", m.Sender, m.Subject, m.Body)
			// go dial(m.Sender, nil)
		}

	}
	caddr := c.RemoteAddr()
	c.Close()
	fmt.Printf("(< %s) serve: connection closed.\n", caddr)
	bareAddr := strings.Split(caddr.String(), ":")[0]
	hub.Remove(bareAddr)
}

func theresANewDirector(m Message) {
	log.SetPrefix("serve: ")
	parts := strings.Split(m.Body, ",")
	//this is the first director and this info is targeted at this node
	if len(parts) == 3 {
		if nDirectors() == 0 && parts[0] == self {
			nid, err := strconv.Atoi(parts[1])
			if err != nil {
				log.Fatalf("bad nodeid err: %s", err)
			}
			nodeID = nid
			perm, err := strconv.Atoi(parts[2])
			if err != nil {
				log.Fatalf("bad permission err: %s", err)
			}
			*permission = perm
			fmt.Printf("*** set permission to %d ***\n", *permission)
			//send permission lvl to gui so user can see it
			sendToGui("perm&" + parts[2])

		}
	}
	addDirector(m.Sender)
	broadcast(m)
}

/* send messages to gui */
func sendToGui(msg string) {
	fmt.Println("sendToGui(): got msg = ", msg)
	p := *permission
	fmt.Printf("have permission = %d\n", p)
	ps := strconv.Itoa(p)
	pm := "&perm=" + ps
	// r := "http://localhost:4000/input?msg=" + pm
	// _, err := http.Get(r)
	// if err != nil {
	// 	log.Println(err)
	// }

	fmt.Println("sending ", msg, "to gui")
	req := "http://localhost:4000/input?msg=" + msg + pm
	_, err := http.Get(req)
	if err != nil {
		log.Println(err)
	}
}

var ping = Message{ID: helper.RandomID(), Subject: "ping"}

func sendPing() {
	log.SetPrefix("sendPing: ")
	if hub.Size() > 0 {
		ping.Sender = self
		for _, a := range hub.ListAddrs() {
			log.Println("ping: address ", a)
			ping.Body = a
			broadcast(ping)
		}
	} else {
		fmt.Println("ping: no peers connected.")
	}
}
