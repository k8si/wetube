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

var votes = struct {
	m map[string]int
	sync.Mutex
}{m: make(map[string]int)}

func takeOffice() {
	fmt.Println("** taking office ***")
	directorAddr = self
	*permission = 0
	welcome = Message{ID: helper.RandomID(), Sender: self, Subject: "welcome"}
	nodeidreg = NodeIDRegistry{id: 0}
}

func vote(id int, addr string) {
	idstr := strconv.Itoa(id)
	m := Message{Sender: self, ID: helper.RandomID(), Subject: "vote", Body: addr + "," + idstr}
	broadcast(m)
	consensus := false
	chosen := nodeID
	if <-allReceived {
		for {
			//block until all votes received
			if <-allVotesReceived {
				fmt.Println("*** all votes received ***")
				for _, id := range votes.m {
					fmt.Printf(" \t\t vote [ %d ] \n", id)
					if id != chosen {
						consensus = false
					}
				}
				break
			}
			continue
		}
	}
	if consensus {
		//then I am the new director
		log.Println("** I am the new director **")
		takeOffice()
		speech := Message{Sender: self, ID: helper.RandomID(), Subject: "welcome", Body: self + "," + strconv.Itoa(chosen)}
		broadcast(speech)
	}
}

func electNewDirector() {
	log.Println("*** starting election ***")
	conns := hub.List()
	log.Println("connected to ", len(conns), "peers")
	hub.PrintAll()
	if len(conns) == 0 {
		//then I am the new director
		log.Println("** I am the new director **")
		takeOffice()
	} else {
		//request peers' nodeid's
		request := Message{Sender: self, ID: helper.RandomID(), Subject: "request", Body: "id"}
		broadcast(request)
		electID := nodeID
		electAddr := self
		//block untill all node ids received
		for {
			if <-allReceived {
				fmt.Println("** received all nodeids **")
				for addr, id := range nodeIDs.m {
					fmt.Printf("\t %s %d\n", addr, id)
					if id < electID {
						electID = id
						electAddr = addr
					}
				}
				break
			}
			continue
		}
		fmt.Printf("** Im voting for id=%d @ addr=%s **\n", electID, electAddr)
		// allReceived = make(chan bool)
		//send vote
		vote(electID, electAddr)
	}
	//vote for the node with the min(nodeid) -- should be no ties cuz of the way nodeIDs are assigned
	// vote.ID = helper.RandomID()
}
