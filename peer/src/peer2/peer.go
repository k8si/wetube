package main

import (
	// "bufio"
	// "crypto/rand"
	// "crypto/tls"
	"flag"
	// "fmt"
	// "helper"
	"log"
	"net/http"
	// "os"
	// "strconv"
	// "strings"
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

func handler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("example server.\n"))
}

func main() {
	http.HandleFunc("/", handler)
	log.Printf("about to listen on 10443")
	err := http.ListenAndServeTLS(":10443", "cert.pem", "key.pem", nil)
	if err != nil {
		log.Fatal(err)
	}
}
