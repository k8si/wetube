package main

import (
	// "crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	// "io"
	"log"
	// "net"
	// "net/http/httputil"
	"helper"
	"time"
)

const (
	ME    = "174.62.219.8"
	ELNUX = "128.119.243.175"
	EC2   = "54.149.118.210"
)

var knownAddrs = []string{ME, ELNUX, EC2}

func invitePeers() {
	for _, addr := range knownAddrs {
		if addr == myipaddr {
			continue
		}
		cert, err := tls.LoadX509KeyPair("certs/client.pem", "certs/client.key")
		if err != nil {
			log.Fatalf("director: loadkeys: %s", err)
		}
		config := tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
		go invite(addr, config)
		// c := make(chan net.Conn)
		// cert, err := tls.LoadX509KeyPair("jan.newmarch.name.pem", "private.pem")
		// checkError(err)
		// cf := &tls.Config{Rand: rand.Reader, Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true} //ServerName: "jan.newmarch.name"}
		// go invite(addr, c, cf)
		// go invite(addr, cf)
		// ssl := tls.Client(<-c, cf)
		// s := make(chan []byte)
		// // thing := httputil.NewClientConn(ssl, nil)
		// // conn := &connection{socket: thing, send: s}
		// conn := &connection{socket: ssl, send: s}
		// conn.socket.Write([]byte("welcome"))
		// h.register <- conn
		// go conn.writer()
		// go conn.reader()
	}
}

func invite(addr string, config tls.Config) {
	log.Printf("director: invite peer at %s", addr)
	peerAddr := addr + ":" + TCP_PORT
	conn := tryInvite(peerAddr, config)
	if conn == nil {
		numRetries := 0
		for {
			log.Printf("dial failed. retrying %s. num retries=%d", peerAddr, numRetries)
			conn = tryInvite(peerAddr, config)
			if conn != nil {
				break
			}
			numRetries += 1
			waitFor := 10
			time.Sleep(time.Duration(waitFor) * time.Second)
			continue
		}
	}
	// defer conn.Close()
	log.Println("director: connected to: ", conn.RemoteAddr())
	state := conn.ConnectionState()
	for _, v := range state.PeerCertificates {
		fmt.Println(x509.MarshalPKIXPublicKey(v.PublicKey))
		fmt.Println(v.Subject)
	}
	log.Println("director: handshake: ", state.HandshakeComplete)
	log.Println("director: mutual: ", state.NegotiatedProtocolIsMutual)
	s := make(chan []byte)
	// thing := httputil.NewClientConn(ssl, nil)
	// conn := &connection{socket: thing, send: s}
	c := &helper.Connection{Socket: conn, Send: s}
	// c.socket.Write([]byte("welcome"))
	h.Register <- c
	go c.Writer()
	go c.Reader()
	// message := "hello\n"
	// c.send <- []byte(message)
	// n, err := io.WriteString(conn, message)
	// if err != nil {
	// 	log.Fatalf("director: write: %s", err)
	// }
	// log.Printf("director: wrote %q (%d bytes)", message, n)
	// reply := make([]byte, 256)
	// n, err = conn.Read(reply)
	// log.Printf("director: read %q (%d bytes)", string(reply[:n]), n)
	// log.Print("director: exiting")
}

func tryInvite(addr string, config tls.Config) *tls.Conn {
	conn, err := tls.Dial("tcp", addr, &config)
	if err != nil {
		return nil
	}
	return conn
}

// func invite(addr string, config *tls.Config) {
// 	fmt.Println("inviting peer @ ", addr)
// 	// tcpAddr, err := net.ResolveTCPAddr("tcp", addr+":"+TCP_PORT)
// 	// if err != nil {
// 	// 	panic(err.Error())
// 	// }
// 	// fmt.Println("tcpaddr ", tcpAddr, "resolved.")
// 	// tcpConn := tryInvite(tcpAddr, config)
// 	targetAddr := addr + ":" + TCP_PORT
// 	tlsConn := tryInvite(targetAddr, config)
// 	if tlsConn == nil {
// 		for {
// 			fmt.Println("retrying...")
// 			tlsConn := tryInvite(targetAddr, config)
// 			if tlsConn != nil {
// 				break
// 			}
// 			waitFor := 10
// 			time.Sleep(time.Duration(waitFor) * time.Second)
// 			continue
// 		}
// 	}
// 	// c <- tcpConn
// 	ssl := tls.Client(tlsConn, config)
// 	s := make(chan []byte)
// 	// thing := httputil.NewClientConn(ssl, nil)
// 	// conn := &connection{socket: thing, send: s}
// 	conn := &connection{socket: ssl, send: s}
// 	// conn.socket.Write([]byte("welcome"))
// 	h.register <- conn
// 	go conn.writer()
// 	go conn.reader()
// }
