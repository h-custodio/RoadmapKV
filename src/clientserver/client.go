package clientserver

import (
	"fmt"
	"log"
	"net"
	"os"
)

func ConnectServer() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	conn, err := net.Dial("tcp", "localhost:"+port)
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	message := "Filler"
	_, err = conn.Write([]byte(message))
	if err != nil {
		log.Println("Write error:", err)
		return
	}

	// read the server's response
	buffer := make([]byte, 1024)
	incomingSize, err := conn.Read(buffer)
	if err != nil {
		log.Fatalf("Failed to write: %v", err)
	}

	fmt.Printf("SERVER RESPONSE: %s", string(buffer[:incomingSize]))
}

func handleClientConnection(conn net.Conn) {

}
