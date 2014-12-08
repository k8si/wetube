package main

import (
	// "crypto/rand"
	// "bufio"
	"crypto/tls"
	// "crypto/x509"
	// "crypto/x509/pkix"
	"encoding/json"
	"fmt"
	"helper"
	"io/ioutil"
	"log"
	// "math/big"
	"net"
	// "os"
	"strings"
	// "time"
)

func broadcast(msg Message) {
	log.SetPrefix("broadcast: ")
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

	// // tcpAddrstr := addr + ":" + helper.TCP_PORT
	// // tcpAddr, err := net.ResolveTCPAddr("tcp", tcpAddrstr)
	// // if err != nil {
	// // 	log.Fatal(err)
	// // }
	// raddr := &net.TCPAddr{IP: net.ParseIP(addr), Port: 3000, Zone: ""}
	// laddr := &net.TCPAddr{IP: net.ParseIP(self), Port: 3000, Zone: ""}
	// c, err := net.DialTCP("tcp", laddr, raddr)
	// if err != nil {
	// 	fmt.Println("dial error")
	// 	log.Fatal(err)
	// }
	// config := tls.Config{ServerName: self}
	// config.Rand = rand.Reader
	// conn := tls.Client(c, &config)
	hostname, err := net.LookupAddr(addr)
	if err != nil {
		log.Println("err looking up hostname for ", addr)
		log.Fatal(err)
	}
	log.Println("got hostname: ", hostname)

	// //configure tls
	config := tls.Config{ServerName: hostname[0], InsecureSkipVerify: true} // I could not get this work

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
		err := enc.Encode(m)
		if err != nil {
			fmt.Printf("(> %s) dial: error: %s\n", conn.RemoteAddr(), err)
			return
		}
	}
}

// func genCert(): x509.Certificate {
// 	template := x509.Certificate{}
// 	lim := new(big.Int).Lsh(big.NewInt(1), 128)
// 	serialno, err := rand.Int(rand.Reader, lim)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	template.SerialNumber = serialno
// 	template.Subject = pkix.Name{Organization: []string{"whatevz"}}
// 	c, err := x509.CreateCertificate(rand.Reader, &template, &template, )

// }

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
