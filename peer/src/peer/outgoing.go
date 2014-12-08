package main

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"helper"
	"log"
	"strings"
	// "time"
)

func broadcast(msg Message) {
	log.Printf("broadcasting: subject=%s, body=%s", msg.Subject, msg.Body)
	for _, ch := range hub.List() {
		select {
		case ch <- msg:
		default:
			//okay to drop messages sometimes?
		}
	}
}

// func sendCloseSignal(addr string) {
// hub.Remove(addr)
// }

func dial(addr string, done chan int) {
	if addr == self {
		return //dont dial self
	}
	ch := hub.Add(addr)
	if ch == nil {
		if done != nil {
			done <- 1
		}
		return //peer already connected
	}

	defer hub.Remove(addr)

	//configure tls
	cert, err := tls.LoadX509KeyPair("cacert.pem", "id_rsa")
	if err != nil {
		log.Fatal(err)
	}
	config := tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
	config.Rand = rand.Reader

	//try to connect
	fmt.Printf("(> %s) dial: dialing port 3000...\n", addr)
	tcpAddr := addr + ":" + helper.TCP_PORT
	conn, err := tls.Dial("tcp", tcpAddr, &config)

	if err != nil {
		fmt.Printf("(> %s) dial: dial error: %s\n", addr, err)
		//TODO add retries
		if done != nil {
			done <- 1
		}
		return
	}

	fmt.Printf("(> %s) dial: connected.\n", conn.RemoteAddr())
	if done != nil {
		done <- 0
	}

	sendPing(inping)

	defer func() {
		checkAddr := strings.Split(conn.RemoteAddr().String(), ":")[0]
		fmt.Printf("(> %s) dial: connection closed.\n", conn.RemoteAddr())
		conn.Close()
		hub.PrintAll()
		removeDirector(checkAddr) //if checkAddr isn't a director, doesnt do anything
		printDirectors()
		if len(directorMap.connected) == 0 {
			go electNewDirector()
		}
	}()

	enc := json.NewEncoder(conn)
	for m := range ch {
		err := enc.Encode(m)
		if err != nil {
			fmt.Printf("(> %s) dial: error: %s\n", conn.RemoteAddr(), err)
			return
		}
	}
}
