package main

import (
	"client"
	"fmt"
	"gui"
	"os"
)

func main() {
	usage := "./app [client|gui]"
	args := os.Args
	if len(args) < 2 {
		os.Exit(1)
	}
	opt := args[1]
	if opt == "client" {
		client.RunClient()
	} else if opt == "gui" {
		gui.RunGUI()
	} else {
		fmt.Println(usage)
		fmt.Println(opt, " not yet implemented")
	}

}
