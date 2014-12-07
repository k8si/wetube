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

func serve(c net.Conn) {
	fmt.Printf("(< %s) serve: accepted connection.\n", c.RemoteAddr())
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

		/* ROUTES */
		switch m.Subject {

		//try to connect to another peer
		case "invite":
			fmt.Printf("(< %s) serve: inviting peer @ %s\n", c.RemoteAddr(), m.Body)
			done := make(chan int)
			go dial(m.Body, done)
			if <-done == 0 {
				hi := Message{Sender: self, ID: helper.RandomID(), Subject: "msg", Body: "hi"}
				fmt.Printf("(< %s) serve: sending hi to %s\n", c.RemoteAddr(), m.Body)
				broadcast(hi)
			}

		//msg received when director invites; try to connect back to director
		case "welcome":
			parts := strings.Split(m.Body, ",")
			directorAddr = parts[0]
			nodeID, err = strconv.Atoi(parts[1])
			if err != nil {
				log.Fatal(err)
			}
			go dial(directorAddr, nil)

		//ping sent when a peer closes; send "ack" to indicate you're still alive
		case "ping":
			ack := Message{Sender: self, ID: helper.RandomID(), Subject: "ack", Body: "ACK"}
			broadcast(ack)

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
			//addr,id
			parts := strings.Split(m.Body, ",")
			idstr := parts[1]
			id, err := strconv.Atoi(idstr)
			if err != nil {
				log.Fatal("bad nodeid received")
			}
			nodeIDs.Lock()
			nodeIDs.m[parts[0]] = id
			nodeIDs.Unlock()
			// n := len(hub.List())
			allReceived <- len(nodeIDs.m) == len(hub.List())
			// if len(nodeIDs.m) == n {
			// 	allReceived <- true
			// }

		case "vote":
			fmt.Printf("(< %s) serve: received vote\n", c.RemoteAddr())
			//addr,id
			parts := strings.Split(m.Body, ",")
			idstr := parts[1]
			id, err := strconv.Atoi(idstr)
			if err != nil {
				log.Fatal("bad nodeid received")
			}
			votes.Lock()
			votes.m[parts[0]] = id
			votes.Unlock()
			// n := len(hub.List())
			allVotesReceived <- len(votes.m) == len(hub.List())
			// if len(votes.m) == n {
			// 	allVotesReceived <- true
			// } else {
			// 	fmt.Printf("(< %s) serve: still waiting for more votes\n", c.RemoteAddr())
			// }

		case "ack", "msg":
			broadcast(m)
			sendToGui(m.Body)
			//TODO colleck ACKs, remove from hub if none from [addr] ?

		default:

		}
		go dial(m.Sender, nil)
	}
	caddr := c.RemoteAddr()
	// checkAddr := strings.Split(c.RemoteAddr().String(), ":")[0]
	// hub.Remove(checkAddr)
	c.Close()
	fmt.Printf("(< %s) serve: connection closed.\n", caddr)

	// ping := Message{ID: helper.RandomID(), Sender: self, Subject: "ping", Body: "ping"}
	// broadcast(ping)

	// hub.PrintAll()
	// if checkAddr == directorAddr {
	// 	fmt.Printf("(< %s) serve: the director left.\n", checkAddr)
	// 	go electNewDirector()
	// }
}
