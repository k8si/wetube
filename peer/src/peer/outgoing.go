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

	defer func() {
		checkAddr := strings.Split(conn.RemoteAddr().String(), ":")[0]
		fmt.Printf("(> %s) dial: connection closed.\n", conn.RemoteAddr())
		conn.Close()
		hub.PrintAll()
		newdirs := make([]string, 0)
		for _, a := range directorAddrs {
			if checkAddr != a {
				newdirs = append(newdirs, a)
			}
		}
		if len(newdirs) == 0 {
			fmt.Printf("(> %s) dial: all directors have left.\n", checkAddr)
			go electNewDirector()
		} else {
			directorAddrs = newdirs
			fmt.Printf("(> %s) dial: %d directors left:\n", checkAddr, len(newdirs))
			for _, a := range directorAddrs {
				fmt.Printf("\t\t%s\n", a)
			}
		}

		// if checkAddr == directorAddr {
		// fmt.Printf("(> %s) dial: director has left.\n", checkAddr)
		// go electNewDirector()
		// }
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

// if err != nil {
// 	ch := make(chan *tls.Conn)
// 	go func(a string, c tls.Config, done chan *tls.Conn) {
// 		for numtries := 0; numtries < 2; numtries += 1 {
// 			fmt.Printf("(> %s) dial: dial error: %s. waiting...\n", a, err)
// 			time.Sleep(time.Duration(10) * time.Second)
// 			fmt.Printf("(> %s) dial: dial error: %s. retrying...\n", a, err)
// 			conn, err = tls.Dial("tcp", a, &c)
// 			if err != nil {
// 				numtries += 1
// 				continue
// 			}
// 			done <- conn
// 		}
// 		done <- nil
// 	}(tcpAddr, config, ch)
// 	conn = <-ch
// 	if conn == nil {
// 		log.Fatal("couldnt get conn")
// 	}
// }
