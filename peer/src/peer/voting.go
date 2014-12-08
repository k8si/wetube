package main

import (
	"fmt"
	"helper"
	"log"
	"strconv"
	"sync"
)

var (
	allReceived      = make(chan bool)
	allVotesReceived = make(chan bool)
)

var nodeIDs = struct {
	m map[string]int
	sync.Mutex
}{m: make(map[string]int)}

func electNewDirector() {
	log.Println("*** starting election ***")
	n := len(hub.List())
	log.Println("connected to ", n, "peers")
	hub.PrintAll()
	if n == 0 {
		//then I am the new director
		log.Println("** I am the new director **")
		takeOffice()
	} else {
		//request peers' nodeid's
		request := Message{Sender: self, ID: helper.RandomID(), Subject: "request", Body: "id"}
		broadcast(request)
		electID := nodeID
		electAddr := self
		fmt.Printf("*** waiting for %d responses ... ***\n", n)
		<-allReceived
		fmt.Println("*** all nodeIDs received. ***")
		for addr, id := range nodeIDs.m {
			fmt.Printf("\t %s %d\n", addr, id)
			if id < electID {
				electID = id
				electAddr = addr
			}
		}
		//send vote
		vote(electID, electAddr)
	}
}

var votes = struct {
	m map[string]int
	sync.Mutex
}{m: make(map[string]int)}

func vote(id int, addr string) {
	fmt.Printf("** voting for id=%d @ addr=%s **\n", id, addr)
	idstr := strconv.Itoa(id)
	m := Message{Sender: self, ID: helper.RandomID(), Subject: "vote", Body: addr + "," + idstr}
	broadcast(m)
	consensus := true
	chosen := nodeID
	<-allVotesReceived
	fmt.Println("*** all votes received ***")
	votes.m[self] = nodeID
	for a, id := range votes.m {
		fmt.Printf(" \t\t %s vote [ %d ] \n", a, id)
		if id != chosen {
			consensus = false
		}
	}
	if consensus {
		//then I am the new director
		log.Println("** I am the new director **")
		takeOffice()
		speech := Message{Sender: self, ID: helper.RandomID(), Subject: "newdirector", Body: "itsme"}
		broadcast(speech)
	}
}

func takeOffice() {
	fmt.Println("** taking office ***")
	directorAddrs = append(directorAddrs, self)
	*permission = 0
	welcome = Message{ID: helper.RandomID(), Sender: self, Subject: "welcome"}
	nodeidreg = NodeIDRegistry{id: 0}
}
