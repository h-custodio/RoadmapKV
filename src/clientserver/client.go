package clientserver

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

	// _, err = conn.Write([]byte("Hello from Client!\n"))
	// if err != nil {
	// 	log.Println("Write error:", err)
	// 	return
	// }

	// basic parser test
	conn.Write([]byte("He"))
	conn.Write([]byte("llo fro"))
	conn.Write([]byte("m Client!\n"))

	log.Println("Message sent!")
}
