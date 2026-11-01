package clientserver

import (
	"bufio"
	"errors"
	"io"
	"net"
	"os"
	"strings"
)

type Client struct {
	conn   net.Conn
	reader *bufio.Reader
}

func ConnectServer() (*Client, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	conn, err := net.Dial("tcp", "localhost:"+port)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}, nil
}

func (c *Client) send(command string) error {
	if c == nil || c.conn == nil {
		return errors.New("client is not connected")
	}

	if !strings.HasSuffix(command, "\n") {
		command += "\n"
	}

	data := []byte(command)

	for len(data) > 0 {
		n, err := c.conn.Write(data)
		if err != nil {
			return err
		}
		if n == 0 {
			return io.ErrShortWrite
		}
		data = data[n:]
	}

	return nil
}

func (c *Client) receive() (string, error) {
	if c == nil || c.reader == nil {
		return "", errors.New("client is not connected")
	}

	response, err := c.reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return response, nil
}

func (c *Client) Execute(command string) (string, error) {
	if err := c.send(command); err != nil {
		return "", err
	}

	return c.receive()
}

func (c *Client) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}

	return c.conn.Close()
}
