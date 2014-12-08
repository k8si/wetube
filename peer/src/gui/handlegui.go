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

var browserConn *websocket.Conn

// func RunGUI() {
func main() {
	service := "localhost:" + WEBSOCKET_PORT
	log.Println("gui listening on ", service)
	http.Handle("/jscli", websocket.Handler(estConnection))
	http.HandleFunc("/input", handleInput)
	err := http.ListenAndServe(service, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func handleInput(w http.ResponseWriter, r *http.Request) {
	msg := r.URL.Query()["msg"][0]
	log.Printf("got input message from goclient: %s\n", msg)
	if browserConn == nil {
		log.Fatal("websocket is nil")
	}
	err := websocket.Message.Send(browserConn, []byte(msg))
	if err != nil {
		log.Fatal(err)
	}
}

func estConnection(ws *websocket.Conn) {
	log.Println("establishing connection with browser....")
	browserConn = ws
	var msg string
	err := websocket.Message.Receive(ws, &msg)
	if err != nil {
		_ = websocket.Message.Send(ws, "FAIL:"+err.Error())
		log.Fatal(err)
	}
	log.Println("got message: ", msg)
	websocket.Message.Send(ws, "1")
	//send the message to the client
	sendToClient(msg)
	//listen for new messages from GUI
	listen(ws)
}

func listen(ws *websocket.Conn) {
	//now wait for new messages
	log.Println("listening for new messages from jsclient...")
	for {
		var m string
		err := websocket.Message.Receive(ws, &m)
		if err != nil {
			log.Println("got err!", err.Error())
			break
		}
		log.Println("got new message: ", m)
		res := sendToClient(m)
		//TODO need to relay response back to browser where appropriate
		err = websocket.Message.Send(ws, []byte(res))

	}
	log.Println("done listening.")
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
}
