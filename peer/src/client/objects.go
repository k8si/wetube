package main

import (
	"crypto/tls"
	"fmt"
	// "net/http"
	// "net"
	// "net/http/httputil"
)

type hub struct {
	connections map[*connection]bool
	broadcast   chan []byte
	register    chan *connection
	unregister  chan *connection
}

type connection struct {
	// socket *net.TCPConn
	// socket *httputil.ClientConn
	socket *tls.Conn
	// socket *net.Conn
	send chan []byte
}

func (h *hub) run() {
	for {
		select {
		case c := <-h.register:
			fmt.Println("registering new connection.")
			h.connections[c] = true
		case c := <-h.unregister:
			fmt.Println("unregistering connection")
			if _, ok := h.connections[c]; ok {
				delete(h.connections, c)
				close(c.send)
			}
		case m := <-h.broadcast:
			// npeers := len(h.connections)
			fmt.Println("rcvd broadcast.")
			for c := range h.connections {
				select {
				case c.send <- m:
				default:
					delete(h.connections, c)
					close(c.send)
				}
			}
		}
	}
}

func (c *connection) writer() {
	fmt.Println("starting writer() for ", c.socket.RemoteAddr())
	for message := range c.send {
		fmt.Println("writing message: ", string(message))
		_, err := c.socket.Write(message)
		if err != nil {
			break
		}
	}
	c.socket.Close()
	fmt.Println("closed writer() for ", c.socket.RemoteAddr())
}

func (c *connection) reader() {
	fmt.Println("starting reader() for ", c.socket.RemoteAddr())
	for {
		msg := make([]byte, 1024)
		_, err := c.socket.Read(msg)
		if err != nil {
			panic(err.Error())
		}
		fmt.Println("got msg: ", string(msg))
		h.broadcast <- msg
	}
	c.socket.Close()
	fmt.Println("closed reader() for ", c.socket.RemoteAddr())
}
