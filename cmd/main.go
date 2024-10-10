package main

import (
	"log"
	"os"
)

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		log.Fatalln("don't know what to do")
	}
	if args[0] == "test" {
		testApp(args[1:])
	} else if args[0] == "main" {
		log.Fatalln("not implemented")
	} else {
		log.Fatalln("invalid app")
	}
}
