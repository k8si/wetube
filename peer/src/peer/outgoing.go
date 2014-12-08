package main

import (
	// "crypto/rand"
	// "bufio"
	"crypto/rsa"
	"crypto/tls"
	// "crypto/x509"
	// "crypto/x509/pkix"
	"encoding/json"
	"fmt"
	"helper"
	"io/ioutil"
	"log"
	// "math/big"
	// "net"
	// "os"
	"strings"
	// "time"
)

func broadcast(msg Message) {
	log.SetPrefix("broadcast: ")
	// msg.sign()
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

	// hostname, err := net.LookupAddr(addr)
	// if err != nil {
	// 	log.Println("err looking up hostname for ", addr)
	// 	log.Fatal(err)
	// }
	// log.Println("got hostname: ", hostname)

	// //configure tls
	// config := tls.Config{ServerName: hostname[0], InsecureSkipVerify: true} // I could not get this work without InsecureSkipVerify
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

	cs := conn.ConnectionState()
	ncerts := len(cs.PeerCertificates)
	fmt.Println("num peer certs: ", ncerts)
	if ncerts == 1 {
		pcert := cs.PeerCertificates[0]
		pubkey := pcert.PublicKey.(*rsa.PublicKey)
		addkey(addr, pubkey)
	}

	// tlsconn, ok := conn.(*tls.Conn)
	// if ok {
	// 	cs := tlsconn.ConnectionState()
	// 	fmt.Println(len(cs.PeerCertificates))
	// }
	// // addr := strings.Split(conn.RemoteAddr().String(), ":")[0]
	// // cs := conn.(tls.Conn).ConnectionState
	// tlsconn, ok := conn.(*tls.Conn)
	// if ok {
	// 	cs := tlsconn.ConnectionState()
	// 	log.Println(cs)
	// }

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

func readInCert() []byte {
	b, err := ioutil.ReadFile("cert.pem")
	if err != nil {
		return nil
	}
	return b
	// f, err := os.Open("cert.pem")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer f.Close()
	// scanner := bufio.NewScanner(f)
	// return scanner.Bytes()
}
