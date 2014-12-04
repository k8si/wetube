package main

import (
	// "client"
	"fmt"
	// "gui"
	"newmarch"
	"os"
)

func main() {
	usage := "./app [gen-keys]"
	fmt.Println(usage)
	args := os.Args
	if len(args) < 2 {
		os.Exit(1)
	}
	opt := args[1]
	if opt == "gen-keys" {
		newmarch.GenRSAKeys()
		newmarch.GenX509Cert()
	} else {
		panic("bad cmdline args: " + opt)
	}
	// if opt == "client" {
	// 	client.RunClient()
	// } else if opt == "gui" {
	// 	gui.RunGUI()
	// } else {
	// 	fmt.Println(usage)
	// 	fmt.Println(opt, " not yet implemented")
	// }

}
