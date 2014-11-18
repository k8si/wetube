package main

import (
	"bufio"
	"code.google.com/p/go.net/websocket"
	"fmt"
	"net"
	"net/http"
	"os"
)

func main() {
	service := "localhost:1337"
	fmt.Println("listening on ", service)
	http.Handle("/websocket/", websocket.Handler(ProcessSocket))
	err := http.ListenAndServe(service, nil)
	checkError(err)
}

func checkError(err error) {
	if err != nil {
		fmt.Println("fatal error: ", err.Error())
		os.Exit(1)
	}
}

func ProcessSocket(ws *websocket.Conn) {
	fmt.Println("in ProcessSocket")
	var msg string
	err := websocket.Message.Receive(ws, &msg)
	if err != nil {
		fmt.Println("ProcessSocket: got error", err)
		_ = websocket.Message.Send(ws, "FAIL:"+err.Error())
		return
	}
	fmt.Println("ProcessSocket: got message", msg, "; doing TCP stuff...")
	service := msg

	tcpAddr, err := net.ResolveTCPAddr("tcp", service)
	if err != nil {
		fmt.Println("error in ResolveTCPAdddr:", err)
		_ = websocket.Message.Send(ws, "FAIL:"+err.Error())
		return
	} else {
		fmt.Println("net.ResolveTCPAddr success")
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Println("error in DialTCP: ", err)
		_ = websocket.Message.Send(ws, "FAIL:"+err.Error())
		return
	} else {
		fmt.Println("net.DialTCP success")
	}
	fmt.Println("done. sending response...")
	//send "1" for "success"
	_ = websocket.Message.Send(ws, "1")
	fmt.Println("done. starting telnet...")
	RunTelnet(ws, conn)
}

func RunTelnet(ws *websocket.Conn, conn *net.TCPConn) {
	fmt.Println("running telnet")
	go ReadSocket(ws, conn)
	//read websocket and write to socket
	crlf := []byte{13, 10}
	var msg string
	for {
		err := websocket.Message.Receive(ws, &msg)
		if err != nil {
			_ = conn.Close()
			break
		}
		_, err = conn.Write([]byte(msg))
		if err != nil {
			break
		}
		fmt.Println("send message to host: ", msg)
		//send slashrslashn as HTTP protocol requires
		_, err = conn.Write(crlf)
		if err != nil {
			break
		}
	}
	fmt.Println("RunTelnet exit")
}

func ReadSocket(ws *websocket.Conn, conn *net.TCPConn) {
	reader := bufio.NewReader(conn)
	var line string = ""
	for {
		if reader == nil {
			break
		}
		buffer, isPrefix, err := reader.ReadLine()
		if err != nil {
			break
		}
		fmt.Println("ReadSocket: got", len(buffer), "bytes")
		line = line + string(buffer)
		if !isPrefix {
			fmt.Println("sending message to websocket: ", line)
			err = websocket.Message.Send(ws, line)
			if err != nil {
				_ = conn.Close()
			}
			line = ""
		}
	}
	fmt.Println("ReadSocket exit")
	ws.Close()
}
