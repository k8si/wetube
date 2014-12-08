package main

import (
	// "bufio"
	// "crypto/rand"
	// "crypto/tls"
	"flag"
	"fmt"
	"helper"
	"log"
	"net/http"
	// "os"
	// "strconv"
	// "strings"
	// "sync"
)

var (
	// initialize   = flag.Bool("init", false, "is this the initial node?") //TODO no longer used
	myAddr      = flag.String("ip", "", "your public ip address") //TODO this is just "self"
	interactive = flag.Bool("i", false, "interactive mode")
	permission  = flag.Int("perm", 2, "permission [0=DIR|1=EDIT|2=VIEW")
	self        string
	nodeID      int
)

var hub = &Hub{peers: make(map[string]chan<- Message)}

func main() {
	log.SetPrefix("main: ")
	//specify initialization with cmdline arg for now
	flag.Parse()
	self = *myAddr

	// //configure TLS
	// cert, err := tls.LoadX509KeyPair("cacert.pem", "id_rsa")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// config := tls.Config{Certificates: []tls.Certificate{cert}}
	// config.Rand = rand.Reader
	// http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
	// 	fmt.Println("got test")
	// })

	// err := http.ListenAndServeTLS(":3000", "cacert.pem", "id_rsa", nil)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	log.Println("hi from main")

	go func() {
		http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("got test")
		})
		if err := http.ListenAndServeTLS(":3000", "cacert.pem", "id_rsa", nil); err != nil {
			log.Panic(err)
		}
	}()
	if *permission == 0 {
		done := make(chan int)
		go sendReq(done)
		<-done
		fmt.Println("got response")
	}
	ch := make(chan bool)
	<-ch

	// http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	fmt.Println("got ROOT")
	// })

	// log.Println("setting up server")
	// err := http.ListenAndServeTLS(":3000", "cacert.pem", "id_rsa", nil)
	// if err != nil {
	// 	fmt.Println("there was an error")
	// 	log.Fatal(err)
	// }
	// fmt.Println("listening on :3000")

}

func sendReq(out chan int) {
	fmt.Println("sending req")
	testAddr := helper.EC2
	testReq := "https://" + testAddr + ":3000/test"
	fmt.Println("sending GET for: ", testReq)
	_, err := http.Get(testReq)
	if err != nil {
		fmt.Println("GET: error")
		out <- 1
	}
	out <- 0
}
