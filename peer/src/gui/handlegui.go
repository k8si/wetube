package main

import (
	"fmt"
	"golang.org/x/net/websocket"
	"log"
	// "net"
	"net/http"
)

const (
	WEBSOCKET_PORT = "4000"
	TCP_PORT       = "3000"
	ACK            = "1"
)

// func RunGUI() {
func main() {
	service := "localhost:" + WEBSOCKET_PORT
	fmt.Println("gui listening on ", service)
	http.Handle("/jscli", websocket.Handler(estConnection))
	err := http.ListenAndServe(service, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func estConnection(ws *websocket.Conn) {
	fmt.Println("establishing connection with browser....")
	var msg string
	err := websocket.Message.Receive(ws, &msg)
	if err != nil {
		_ = websocket.Message.Send(ws, "FAIL:"+err.Error())
		log.Fatal(err)
	}
	fmt.Println("got message: ", msg)
	websocket.Message.Send(ws, "1")
	//send the message to the client
	sendToClient(msg)
	listen(ws)
}

func listen(ws *websocket.Conn) {
	//now wait for new messages
	fmt.Println("listening for new messages...")
	for {
		var m string
		err := websocket.Message.Receive(ws, &m)
		if err != nil {
			fmt.Println("got err!", err.Error())
			break
		}
		fmt.Println("got new message: ", m)
		res := sendToClient(m)
		//TODO need to relay response back to browser where appropriate
		err = websocket.Message.Send(ws, []byte(res))
	}
	fmt.Println("done listening.")
}

func sendToClient(msg string) string {
	fmt.Println("relaying message: ", msg)
	req := "http://localhost:3001/jsclient?msg=" + msg
	res, err := http.Get(req)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(res)
	if res.StatusCode != 200 {
		panic("bad status code: " + string(res.StatusCode))
	}
	return "1"

	// addr, err := net.ResolveTCPAddr("tcp", "localhost:"+TCP_PORT)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// conn, err := net.DialTCP("tcp", nil, addr)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// conn.Write([]byte(msg))
	// response := make([]byte, 1024)
	// _, err = conn.Read(response)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// m := string(response)
	// fmt.Println("got response: ", m)
	// conn.Close()
	// // if m != ACK {
	// // log.Fatal("error in sendToClient for msg: ", msg)
	// // }
	// fmt.Println("success.\n")
	// return m
}
