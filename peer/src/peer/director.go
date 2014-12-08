package main

import (
	"fmt"
	"helper"
	"sync"
)

var (
	nodeidreg NodeIDRegistry
	welcome   Message
)

var directorMap = struct {
	connected map[string]bool
	mu        sync.Mutex
}{connected: map[string]bool{}}

func takeOffice() {
	fmt.Println("*** taking office ***")
	addDirector(self)
	*permission = 0
	welcome = Message{ID: helper.RandomID(), Sender: self, Subject: "welcome"}
	nodeidreg = NodeIDRegistry{id: 0}
}

func addDirector(addr string) {
	fmt.Printf("*** adding director @ %s. ***\n", addr)
	// other peers could be declaring themselves director at this very moment
	directorMap.mu.Lock()
	ok := directorMap.connected[addr]
	if !ok {
		directorMap.connected[addr] = true
	}
	directorMap.mu.Unlock()
	printDirectors()
}

func removeDirector(checkAddr string) {
	fmt.Printf("*** director @ %s left. ***\n", checkAddr)
	directorMap.mu.Lock()
	if ok := directorMap.connected[checkAddr]; ok {
		delete(directorMap.connected, checkAddr)
	}
	directorMap.mu.Unlock()
	printDirectors()

}

func printDirectors() {
	fmt.Printf("*** %d directors remaining: ***\n", len(directorMap.connected))
	for k, _ := range directorMap.connected {
		fmt.Printf("\t\t%s\n", k)
	}
}

func nDirectors() int {
	return len(directorMap.connected)
}
