package clientserver

import (
	"bufio"
	"errors"
	"io"
	"net"
	"strconv"
	"testing"
	"time"
)

func newTestClient(t *testing.T) (*Client, net.Conn) {
	t.Helper()

	clientConn, serverConn := net.Pipe()
	t.Cleanup(func() {
		_ = clientConn.Close()
		_ = serverConn.Close()
	})

	return &Client{
		conn:   clientConn,
		reader: bufio.NewReader(clientConn),
	}, serverConn
}

func TestConnectServer(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("starting test listener: %v", err)
	}
	defer listener.Close()

	t.Setenv("PORT", strconv.Itoa(listener.Addr().(*net.TCPAddr).Port))

	accepted := make(chan net.Conn, 1)
	go func() {
		conn, err := listener.Accept()
		if err == nil {
			accepted <- conn
		}
	}()

	client, err := ConnectServer()
	if err != nil {
		t.Fatalf("connecting to test server: %v", err)
	}
	defer client.Close()

	select {
	case conn := <-accepted:
		defer conn.Close()
	case <-time.After(time.Second):
		t.Fatal("test server did not accept connection")
	}
}

func TestConnectServerUnavailable(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("starting test listener: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	_ = listener.Close()

	t.Setenv("PORT", strconv.Itoa(port))

	client, err := ConnectServer()
	if err == nil {
		_ = client.Close()
		t.Fatal("ConnectServer succeeded without an available server")
	}
	if client != nil {
		t.Fatal("ConnectServer returned a client with an error")
	}
}

func TestSendAddsSingleNewline(t *testing.T) {
	client, serverConn := newTestClient(t)
	serverReader := bufio.NewReader(serverConn)

	for _, command := range []string{"PING", "PONG\n"} {
		received := make(chan string, 1)
		go func() {
			command, err := serverReader.ReadString('\n')
			if err != nil {
				t.Errorf("reading command: %v", err)
				return
			}
			received <- command
		}()

		if err := client.send(command); err != nil {
			t.Fatalf("sending %q: %v", command, err)
		}

		want := command
		if command == "PING" {
			want += "\n"
		}
		if got := <-received; got != want {
			t.Errorf("received command = %q, want %q", got, want)
		}
	}
}

func TestSendRejectsUninitializedClient(t *testing.T) {
	var client *Client
	if err := client.send("PING"); err == nil {
		t.Fatal("send succeeded for nil client")
	}

	if err := (&Client{}).send("PING"); err == nil {
		t.Fatal("send succeeded for client without a connection")
	}
}

func TestReceiveReadsFragmentedResponse(t *testing.T) {
	client, serverConn := newTestClient(t)

	go func() {
		_, _ = serverConn.Write([]byte("PON"))
		_, _ = serverConn.Write([]byte("G\n"))
	}()

	response, err := client.receive()
	if err != nil {
		t.Fatalf("receiving fragmented response: %v", err)
	}
	if response != "PONG\n" {
		t.Fatalf("response = %q, want %q", response, "PONG\n")
	}
}

func TestReceiveReturnsServerDisconnect(t *testing.T) {
	client, serverConn := newTestClient(t)
	_ = serverConn.Close()

	_, err := client.receive()
	if !errors.Is(err, io.EOF) {
		t.Fatalf("receive error = %v, want EOF", err)
	}
}

func TestReceiveRejectsUninitializedClient(t *testing.T) {
	var client *Client
	if _, err := client.receive(); err == nil {
		t.Fatal("receive succeeded for nil client")
	}

	if _, err := (&Client{}).receive(); err == nil {
		t.Fatal("receive succeeded for client without a reader")
	}
}

func TestExecutePersistentSession(t *testing.T) {
	conn := startTestConnection(t)
	client := &Client{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}
	defer client.Close()

	for _, test := range []struct {
		command  string
		response string
	}{
		{command: "SET name alice", response: "OK\n"},
		{command: "GET name", response: "alice\n"},
		{command: "DEL name", response: "OK\n"},
		{command: "GET name", response: "not valid\n"},
		{command: "QUIT", response: "Quitting...\n"},
	} {
		response, err := client.Execute(test.command)
		if err != nil {
			t.Fatalf("executing %q: %v", test.command, err)
		}
		if response != test.response {
			t.Errorf("response for %q = %q, want %q", test.command, response, test.response)
		}
	}
}

func TestExecutePropagatesSendError(t *testing.T) {
	client, serverConn := newTestClient(t)
	_ = client.Close()
	_ = serverConn.Close()

	if _, err := client.Execute("PING"); err == nil {
		t.Fatal("execute succeeded after client connection was closed")
	}
}

func TestCloseIsSafeForUninitializedClient(t *testing.T) {
	var client *Client
	if err := client.Close(); err != nil {
		t.Fatalf("closing nil client: %v", err)
	}

	if err := (&Client{}).Close(); err != nil {
		t.Fatalf("closing client without a connection: %v", err)
	}
}
