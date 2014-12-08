package main

import (
	// "bufio"
	"crypto/rand"
	"crypto/tls"
	// "flag"
	// "fmt"
	// "helper"
	"log"
	// "net/http"
	// "os"
	// "strconv"
	// "strings"
	// "sync"
)

func main() {
	cert, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	if err != nil {
		log.Fatal(err)
	}
	config := tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
	config.Rand = rand.Reader
	addr := "174.62.219.8:10443"
	conn, err := tls.Dial("tcp", addr, &config)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	conn.Write([]byte("hi"))
	// client := tls.Client(conn, &config)

	// req := "https://174.62.219.8:10443/"
	// res, err := http.Get(req)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// log.Println(res)
}
