package main

import (
	"fmt"
	"golang.org/x/net/websocket"
	"log"
	"net/http"
)

const WEBSOCKET_PORT = "4000"

func main() {
	service := ":" + WEBSOCKET_PORT
	fmt.Println("listening on ", WEBSOCKET_PORT)
	http.Handle("/jscli", websocket.Handler(estConnection))
	err := http.ListenAndServe(service, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func estConnection(ws *websocket.Conn) {
	fmt.Println("doing initBrowser2ClientSocket()....")
	var msg string
	err := websocket.Message.Receive(ws, &msg)
	if err != nil {
		_ = websocket.Message.Send(ws, "FAIL:"+err.Error())
		log.Fatal(err)
	}
	fmt.Println("got message: ", msg)
	websocket.Message.Send(ws, "1")
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
	}
	fmt.Println("done listening.")
}
