package clientserver

import (
	"io"
	"net"
	"testing"
	"time"

	"RoadmapKV/src/storage"
)

// startTestConnection runs the same connection handler used by the TCP server,
// but replaces a real socket with an in-memory connection pair. This keeps the
// test focused on the protocol while avoiding a fixed port and a forever-running
// listener.
func startTestConnection(t *testing.T) net.Conn {
	t.Helper()

	serverConn, clientConn := net.Pipe()
	t.Cleanup(func() {
		_ = clientConn.Close()
	})

	go handleConnection(serverConn, storage.NewStore())
	return clientConn
}

// readResponse reads exactly the bytes expected from one or more commands.
// Responses currently have no delimiter, so the test must know the expected
// response length instead of waiting for a newline.
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
		{command: "SET name alice\n", response: "OK"},
		{command: "GET name\n", response: "alice"},
		{command: "DEL name\n", response: "OK"},
		{command: "GET name\n", response: "not valid"},
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

	if got := readResponse(t, conn, "OKblue"); got != "OKblue" {
		t.Errorf("pipelined response = %q, want %q", got, "OKblue")
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

	if got := readResponse(t, conn, "OK"); got != "OK" {
		t.Errorf("split-command response = %q, want %q", got, "OK")
	}
}

func TestHandleConnectionQuitClosesConnection(t *testing.T) {
	conn := startTestConnection(t)

	if _, err := conn.Write([]byte("QUIT\n")); err != nil {
		t.Fatalf("writing QUIT: %v", err)
	}

	if got := readResponse(t, conn, "Quitting..."); got != "Quitting..." {
		t.Errorf("QUIT response = %q, want %q", got, "Quitting...")
	}

	// QUIT should end the handler after its response, so the peer observes EOF.
	_ = conn.SetReadDeadline(time.Now().Add(time.Second))
	_, err := conn.Read(make([]byte, 1))
	if err != io.EOF {
		t.Fatalf("reading after QUIT = %v, want EOF", err)
	}
}
