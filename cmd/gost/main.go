package main

import (
	"log"
)

func main() {
	log.Println("GoST — modular ECS terminal emulator starting...")
	if err := StartGame(); err != nil {
		log.Fatal(err)
	}
}

