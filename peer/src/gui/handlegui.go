package main

import (
	"fmt"
	"golang.org/x/net/websocket"
	"log"
	// "net"
	"flag"
	"net/http"
	"strings"
)

const (
	WEBSOCKET_PORT = "4000"
	TCP_PORT       = "3000"
	ACK            = "1"
)

var browserConn *websocket.Conn
var service = flag.String("service", ":4000", "[IP addr]:[port] to listen on")

func main() {
	flag.Parse()
	log.Println("gui listening on ", *service)
	http.Handle("/jscli", websocket.Handler(estConnection))
	http.HandleFunc("/input", handleInput)
	err := http.ListenAndServe(*service, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func handleInput(w http.ResponseWriter, r *http.Request) {
	msg := r.URL.Query()["msg"][0]
	//TODO make this smoother
	p := r.URL.Query()["perm"][0]
	pmsg := "perm&" + p
	log.Printf("got input message from goclient: %s\n", msg)
	log.Printf("have permission %s\n", p)
	if browserConn == nil {
		log.Fatal("websocket is nil")
	}
	err := websocket.Message.Send(browserConn, []byte(pmsg))
	if err != nil {
		log.Fatalf("err sending message to jsclient: %s\n", err)
	}
	err = websocket.Message.Send(browserConn, []byte(msg))
	if err != nil {
		log.Fatalf("err sending message to jsclient: %s\n", err)
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
			log.Fatal("err reading from jsclient:", err.Error())
		}
		log.Println("got new message: ", m)
		res := sendToClient(m)
		err = websocket.Message.Send(ws, []byte(res))
	}
	log.Println("done listening.")
}

func sendToClient(msg string) string {
	fmt.Println("relaying message: ", msg)
	var req string
	if strings.HasPrefix(msg, "invite") {
		req = "http://localhost:3001/inv?" + msg
	} else {
		req = "http://localhost:3001/jsclient?msg=" + msg
	}
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
