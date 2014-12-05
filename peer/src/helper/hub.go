package helper

import (
	"crypto/tls"
	"fmt"
	"strings"
	// "net/http"
	// "net"
	// "net/http/httputil"
)

type Hub struct {
	Connections map[*Connection]bool
	Broadcast   chan []byte
	Register    chan *Connection
	Unregister  chan *Connection
}

type Connection struct {
	// socket *net.TCPConn
	// socket *httputil.ClientConn
	Socket *tls.Conn
	// socket *net.Conn
	Send chan []byte
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.Register:
			newAddr := c.Socket.RemoteAddr().String()
			fmt.Println("registering new connection to ", newAddr)
			msg := "NEWPEER:" + newAddr
			for conn := range h.Connections {
				//tell the existing peer about the new one
				conn.Send <- []byte(msg)
				thisAddr := conn.Socket.RemoteAddr().String()
				//tell the new peer about the existing one
				msg = "NEWPEER:" + thisAddr
				c.send <- []byte(msg)
			}
			h.Connections[c] = true
		case c := <-h.Unregister:
			fmt.Println("unregistering connection")
			if _, ok := h.Connections[c]; ok {
				c.Socket.Close()
				delete(h.Connections, c)
				close(c.Send)
			}
		case m := <-h.Broadcast:
			// npeers := len(h.connections)
			// fmt.Println("rcvd broadcast.")
			// fmt.Println("sending broadcast: ", m)
			nconn := len(h.Connections)

			senderAddr := strings.Split(string(m), ":")[0]
			fmt.Println("sending broadcast ", m, " rcv'd from ", senderAddr, " to numpeers: ", nconn)
			for c := range h.Connections {
				// targetStr := string(c.socket.RemoteAddr())
				target := strings.Split(c.Socket.RemoteAddr().String(), ":")[0]
				if target != senderAddr {
					fmt.Println("sending to ", c.Socket.RemoteAddr())
					c.Send <- m
				}
				// select {
				// case c.send <- m:
				// default:
				// 	delete(h.connections, c)
				// 	close(c.send)
				// }
			}
		}
	}
}

func (c *Connection) Writer() {
	fmt.Println("starting writer()")
	for message := range c.Send {
		fmt.Println("writing message: ", string(message))
		_, err := c.Socket.Write(message)
		checkError(err)
	}
	c.Socket.Close()
	fmt.Println("closed writer()")
}

func (c *Connection) Reader() {
	fmt.Println("starting reader()")
	for {
		msg := make([]byte, 1024)
		_, err := c.Socket.Read(msg)
		checkError(err)
		fmt.Println("got msg: ", string(msg))
		// h.broadcast <- msg
	}
	c.Socket.Close()
	fmt.Println("closed reader()")
}

func checkError(err error) {
	if err != nil {
		panic(err.Error())
	}
}
