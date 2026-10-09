package clientserver

import (
	"fmt"
	"io"
	"log"
	"net"

	"RoadmapKV/src/command"
	"RoadmapKV/src/storage"
)

func CreateServer(s *storage.Store) {
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
		go handleConnection(conn, s)
	}
}

func handleConnection(conn net.Conn, s *storage.Store) {
	// clean up and close the resource when this function returns
	defer conn.Close()

	// buffer for byte stream
	var data []byte

	for {
		streamBuffer := make([]byte, 1024)
		incomingSize, err := conn.Read(streamBuffer)
		if err != nil {
			if err == io.EOF {
				log.Println("Client disconnected")
			} else {
				log.Println("Error reading from connection:", err)
			}
			return
		}

		// convert buffer slice bytes into string
		data = append(data, streamBuffer[:incomingSize]...)

		// originally not within a sub loop
		// consume as many complete commands as are currently buffered within a read cycle
		for {
			res := command.Parse(data)
			if !res.IsComplete {
				break
			}

			command.Execute(s, string(res.Line[0]), res.Line[1:])
			// shifts the data buffer forward
			// the state
			data = data[res.BytesConsumed:]
		}
	}
}
