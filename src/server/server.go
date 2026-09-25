package server

import (
	"fmt"
	"log"
	"net"
)

func CreateServer() {
	// listen for incoming TCP connections on port 8080
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Server is listening on port 8080...")

	// keep looping to handle further connections
	for {
		// waits for and returns the next connection to the listener
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			// we want to keep handling other connections
			continue
		}

		// handle connection concurrently with goroutine
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	// clean up and close the resource when this function returns
	defer conn.Close()

	// gets bytes coming from client
	// command := parse(conn.Read)

	// buffer for byte stream
	buffer := make([]byte, 1024)

	incomingSize, err := conn.Read(buffer)
	if err != nil {
		log.Println("Error reading from conncection:", err)
		return
	}

	// convert slice bytes into string
	message := string(buffer[:incomingSize])

	// for testing
	log.Println("Received: ", message)

}
