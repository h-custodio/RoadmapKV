package main

import (
	"RoadmapKV/src/clientserver"
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	client, err := clientserver.ConnectServer()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to connect to server:", err)
		return
	}
	defer client.Close()

	// creates a buffered reader
	// since NewScanner strips newline chars
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Enter a command")

	for {
		fmt.Println("> ")

		// read input up to and including the delimiter character ('\n')
		input, err := reader.ReadString('\n')
		if err != nil {
			// handle EOF (Ctrl+D / Ctrl+Z) gracefully
			if err == io.EOF {
				fmt.Println("\nExiting due to EOF.")
				break
			}
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			break
		}

		response, err := client.Execute(input)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to execute command:", err)
			continue
		}

		fmt.Print(response)

		if strings.EqualFold(strings.TrimSpace(input), "QUIT") {
			break
		}
	}
}
