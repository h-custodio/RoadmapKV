package client

import (
	//"fmt"
	"log"
	"net"
)

func ConnectServer() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte("Hello World! Aaaaaand veryyyyyyyy longgggggggggggg wordddddddddddddddddd\n"))
	if err != nil {
		log.Println("Write error:", err)
		return
	}

	log.Println("Message sent!")
}
