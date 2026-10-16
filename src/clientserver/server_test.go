package clientserver

import (
	"io"
	"net"
	"testing"
	"time"

	"RoadmapKV/src/storage"
)

// startTestConnection runs the same connection handler used by the TCP server
// on an ephemeral loopback port.
func startTestConnection(t *testing.T) net.Conn {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("starting test listener: %v", err)
	}

	clientConn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		_ = listener.Close()
		t.Fatalf("connecting to test listener: %v", err)
	}

	serverConn, err := listener.Accept()
	if err != nil {
		_ = clientConn.Close()
		_ = listener.Close()
		t.Fatalf("accepting test connection: %v", err)
	}

	_ = listener.Close()
	t.Cleanup(func() {
		_ = clientConn.Close()
	})

	go handleConnection(serverConn, storage.NewStore())
	return clientConn
}

// readResponse reads exactly the bytes expected from one or more commands.
// Responses are newline-delimited.
func readResponse(t *testing.T, conn net.Conn, want string) string {
	t.Helper()

	response := make([]byte, len(want))
	if _, err := io.ReadFull(conn, response); err != nil {
		t.Fatalf("reading response: %v", err)
	}

	return string(response)
}

func TestHandleConnectionSetGetDelete(t *testing.T) {
	conn := startTestConnection(t)

	// A client writes one complete command, then waits for its response before
	// sending the next command, which models a simple request/response client.
	for _, test := range []struct {
		command  string
		response string
	}{
		{command: "SET name alice\n", response: "OK\n"},
		{command: "GET name\n", response: "alice\n"},
		{command: "DEL name\n", response: "OK\n"},
		{command: "GET name\n", response: "not valid\n"},
	} {
		if _, err := conn.Write([]byte(test.command)); err != nil {
			t.Fatalf("writing %q: %v", test.command, err)
		}

		if got := readResponse(t, conn, test.response); got != test.response {
			t.Errorf("response for %q = %q, want %q", test.command, got, test.response)
		}
	}
}

func TestHandleConnectionPipelinedCommands(t *testing.T) {
	conn := startTestConnection(t)

	// TCP is a byte stream, so several commands can arrive in one write. The
	// server must consume each complete command and preserve their order.
	if _, err := conn.Write([]byte("SET color blue\nGET color\n")); err != nil {
		t.Fatalf("writing pipelined commands: %v", err)
	}

	if got := readResponse(t, conn, "OK\nblue"); got != "OK\nblue" {
		t.Errorf("pipelined response = %q, want %q", got, "OK\nblue")
	}
}

func TestHandleConnectionCommandSplitAcrossWrites(t *testing.T) {
	conn := startTestConnection(t)

	// A network read is not guaranteed to return a whole command. The handler
	// must retain an incomplete command until the remaining bytes arrive.
	if _, err := conn.Write([]byte("SET greeting #hello")); err != nil {
		t.Fatalf("writing first command fragment: %v", err)
	}

	// No response should be available until the newline completes the command.
	_ = conn.SetReadDeadline(time.Now().Add(20 * time.Millisecond))
	if _, err := conn.Read(make([]byte, 2)); err == nil {
		t.Fatal("received a response before command was complete")
	}
	_ = conn.SetReadDeadline(time.Time{})

	if _, err := conn.Write([]byte(" world#\n")); err != nil {
		t.Fatalf("writing final command fragment: %v", err)
	}

	if got := readResponse(t, conn, "OK\n"); got != "OK\n" {
		t.Errorf("split-command response = %q, want %q", got, "OK\n")
	}
}

func TestHandleConnectionQuitClosesConnection(t *testing.T) {
	conn := startTestConnection(t)

	if _, err := conn.Write([]byte("QUIT\n")); err != nil {
		t.Fatalf("writing QUIT: %v", err)
	}

	if got := readResponse(t, conn, "Quitting...\n"); got != "Quitting...\n" {
		t.Errorf("QUIT response = %q, want %q", got, "Quitting...\n")
	}

	// QUIT should end the handler after its response, so the peer observes EOF.
	_ = conn.SetReadDeadline(time.Now().Add(time.Second))
	_, err := conn.Read(make([]byte, 1))
	if err != io.EOF {
		t.Fatalf("reading after QUIT = %v, want EOF", err)
	}
}
