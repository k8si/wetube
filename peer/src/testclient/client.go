package main

import (
	// "bufio"
	// "crypto/rand"
	// "crypto/tls"
	// "flag"
	// "fmt"
	// "helper"
	"log"
	"net/http"
	// "os"
	// "strconv"
	// "strings"
	// "sync"
)

func main() {
	req := "https://174.62.219.8:10443/"
	res, err := http.Get(req)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(res)
}
