package main

import (
	"encoding/json"
	"fmt"
	"helper"
	"log"
	"net"
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

/*
	newpeer := Message{ID: helper.RandomID(), Sender: self, Subject: "invite", Body: addr}
	broadcast(newpeer)
*/

func serve(c net.Conn) {
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

		fmt.Printf("(< %v) serve: RCVD: %v\n", c.RemoteAddr(), m)

		go dial(m.Sender, nil)

		/* ROUTES */
		switch m.Subject {

		// //try to connect to another peer
		// case "invite":
		// 	fmt.Printf("(< %s) serve: inviting peer @ %s\n", c.RemoteAddr(), m.Body)
		// 	done := make(chan int)
		// 	go dial(m.Body, done)
		// 	if <-done == 0 {
		// 		hi := Message{Sender: self, ID: helper.RandomID(), Subject: "msg", Body: "hi"}
		// 		fmt.Printf("(< %s) serve: sending hi to %s\n", c.RemoteAddr(), m.Body)
		// 		broadcast(hi)
		// 	}
		case "ping":
			fmt.Printf("*** got ping for %s ***\n", m.Body)
			//if we try to dial an empty address we end up dialing 127.0.0.1
			if len(m.Body) > 0 {
				go dial(m.Body, nil)
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
			// sendToGui(m.Body)

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
		}
	}
	addDirector(m.Sender)
	broadcast(m)
}

var ping = Message{ID: helper.RandomID(), Sender: self, Subject: "ping"}

func sendPing() {
	if hub.Size() > 0 {
		for _, a := range hub.ListAddrs() {
			log.Println("ping: address ", a)
			ping.Body = a
			broadcast(ping)
		}
	} else {
		fmt.Println("ping: no peers connected.")
	}
}
