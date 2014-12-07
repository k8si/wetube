package main

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"helper"
	"log"
	"strings"
)

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

	defer func() {
		checkAddr := strings.Split(conn.RemoteAddr().String(), ":")[0]
		fmt.Printf("(> %s) dial: connection closed.\n", conn.RemoteAddr())
		conn.Close()
		// ping := Message{ID: helper.RandomID(), Sender: self, Subject: "ping", Body: "ping"}
		// broadcast(ping)
		hub.PrintAll()
		if checkAddr == directorAddr {
			fmt.Printf("(> %s) dial: director has left.\n", checkAddr)
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
