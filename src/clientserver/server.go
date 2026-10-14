package clientserver

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"RoadmapKV/src/command"
	"RoadmapKV/src/storage"
)

func CreateServer(s *storage.Store) {
	// allows server port to be dynamic
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// listen for incoming TCP connections
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Server is listening...")

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

	// CONSIDER: using bufio for handing TCP stream instead
	// this version might struggle at scale because of frequent allocatioms
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

		// check if buffer gets too big before parsing
		// arbitrary max data buffer size is 64000
		if len(data) > 64000 {
			log.Println("Client buffer exceeded maximum allowed size. Closing connection.")
			conn.Write([]byte("ERR line too long\n"))
			return
		}

		// consume as many complete commands within the current buffer
		for {
			parseRes := command.Parse(data)
			if !parseRes.IsComplete {
				break
			}

			cmdRes := command.Execute(s, parseRes.Line)

			// shifts the data buffer forward
			// the state
			data = data[parseRes.BytesConsumed:]

			if cmdRes.Response != nil {
				conn.Write(cmdRes.Response)
			}
			if cmdRes.CloseConn {
				// triggers defer conn.Close() and exits the goroutine safely
				return
			}
		}
	}
}
