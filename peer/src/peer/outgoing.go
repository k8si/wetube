package main

import (
	"crypto/rsa"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"helper"
	"log"
	"strings"
)

func broadcast(msg Message) {
	// log.Printf("broadcasting: subject=%s, body=%s", msg.Subject, msg.Body)
	for _, ch := range hub.List() {
		select {
		case ch <- msg:
		default:
			//okay to drop messages sometimes?
		}
	}
}

func dial(addr string, done chan int) {
	log.SetPrefix("dial: ")
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
	// hostname, err := net.LookupAddr(addr)
	// if err != nil {
	// 	log.Println("err looking up hostname for ", addr)
	// 	log.Fatal(err)
	// }
	// log.Println("got hostname: ", hostname)
	// config := tls.Config{ServerName: hostname[0], InsecureSkipVerify: true} // I could *NOT* get this work without InsecureSkipVerify (I think because of CA issues?)
	config := tls.Config{InsecureSkipVerify: true}
	//try to connect
	fmt.Printf("(> %s) dial: dialing port 3000...\n", addr)
	tcpAddr := addr + ":" + helper.TCP_PORT
	conn, err := tls.Dial("tcp", tcpAddr, &config)
	if err != nil {
		fmt.Printf("(> %s) dial: dial error: %s\n", addr, err)
		//TODO add retries somehow
		if done != nil {
			done <- 1
		}
		return
	}

	fmt.Printf("(> %s) dial: connected.\n", conn.RemoteAddr())
	if done != nil {
		done <- 0
	}

	//grab the peer's public key from the TLS connection state
	cs := conn.ConnectionState()
	ncerts := len(cs.PeerCertificates)
	if ncerts == 1 {
		pcert := cs.PeerCertificates[0]
		pubkey := pcert.PublicKey.(*rsa.PublicKey)
		addkey(addr, pubkey)
	}

	defer func() {
		checkAddr := strings.Split(conn.RemoteAddr().String(), ":")[0]
		fmt.Printf("(> %s) dial: connection closed.\n", conn.RemoteAddr())
		conn.Close()
		hub.PrintAll()
		//check if all the directors have left
		removeDirector(checkAddr) //if checkAddr isn't a director, doesnt do anything
		printDirectors()
		if len(directorMap.connected) == 0 {
			go electNewDirector()
		}
	}()

	sendPing()

	enc := json.NewEncoder(conn)
	for m := range ch {
		m.sign()
		err := enc.Encode(m)
		if err != nil {
			fmt.Printf("(> %s) dial: error: %s\n", conn.RemoteAddr(), err)
			return
		}
	}
}
