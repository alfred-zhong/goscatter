package goscatter

import (
	"errors"
	"fmt"
	"net"
	"os"

	"github.com/fatih/color"
)

// Server is used to manage scatters. When receiving a connection, server
// will create a Scatter to send and receive messages to remote address.
type Server struct {
	port         int
	remoteAddr   string
	scatterAddrs []string

	listener net.Listener
}

// NewServer creates Server using port and remoteAddr.
func NewServer(port int, remoteAddr string) (*Server, error) {
	if port < 1 || port > 65535 {
		return nil, errors.New("port invalid")
	}

	if _, err := net.ResolveTCPAddr("tcp", remoteAddr); err != nil {
		return nil, errors.New("remoteAddr identified as tcp address is invalid")
	}

	return &Server{
		port:         port,
		remoteAddr:   remoteAddr,
		scatterAddrs: make([]string, 0),
	}, nil
}

// AddScatterAddr add a scatter address to the server.
func (s *Server) AddScatterAddr(scatterAddr string) error {
	if _, err := net.ResolveTCPAddr("tcp", scatterAddr); err != nil {
		return errors.New("scatterAddr identified as tcp address is invalid")
	}
	s.scatterAddrs = append(s.scatterAddrs, scatterAddr)
	return nil
}

// Run the server to receive connections.
func (s *Server) Run() error {
	// listen for connection
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("server listen fail: %v", err)
	}
	defer listener.Close()
	s.listener = listener

	color.Green("server started. Listen at %d\n", s.port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			break
		}

		color.Blue("connect from %v\n", conn.RemoteAddr())

		// create scatter and run
		scatter, err := NewScatter(conn, s.remoteAddr, s.scatterAddrs)
		if err != nil {
			conn.Close()
			continue
		}

		go func() {
			if err := scatter.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Scatter run fails: %v\n", err)
				conn.Close()
				color.Blue("disconnect from %v\n", conn.RemoteAddr())
			}
		}()
	}

	color.Green("server stop.\n")
	return nil
}

// Stop the server.
func (s *Server) Stop() {
	if s.listener != nil {
		s.listener.Close()
	}
}
