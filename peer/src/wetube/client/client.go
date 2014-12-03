package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
)

const TCP_PORT = "3000"

func main() {
	service := ":" + TCP_PORT
	fmt.Println("client listening on ", TCP_PORT)
	ln, err := net.Listen("tcp", service)
	if err != nil {
		log.Fatal(err)
	}

}
