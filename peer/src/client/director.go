package main

import (
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"net"
	// "net/http/httputil"
	"time"
)

func invitePeers() {
	fmt.Println("in invitePeers()....")
	for _, addr := range knownAddrs {
		if addr == myipaddr {
			continue
		}
		c := make(chan net.Conn)
		go invite(addr, c)
		cf := &tls.Config{Rand: rand.Reader}
		ssl := tls.Client(<-c, cf)
		s := make(chan []byte)
		// thing := httputil.NewClientConn(ssl, nil)
		// conn := &connection{socket: thing, send: s}
		conn := &connection{socket: ssl, send: s}
		conn.socket.Write([]byte("welcome"))
		h.register <- conn
		go conn.writer()
		go conn.reader()
	}
}

func invite(addr string, c chan net.Conn) {
	fmt.Println("inviting peer @ ", addr)
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr+":"+TCP_PORT)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("tcpaddr ", tcpAddr, "resolved.")
	tcpConn := tryInvite(tcpAddr)
	if tcpConn == nil {
		for {
			fmt.Println("retrying...")
			tcpConn := tryInvite(tcpAddr)
			if tcpConn != nil {
				break
			}
			waitFor := 10
			time.Sleep(time.Duration(waitFor) * time.Second)
			continue
		}
	}
	c <- tcpConn
}

func tryInvite(tcpAddr *net.TCPAddr) *net.TCPConn {
	tcpConn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil
	}
	return tcpConn
}
