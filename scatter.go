package goscatter

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"

	"github.com/fatih/color"
)

type Scatter struct {
	mainAddr     *net.TCPAddr
	scatterAddrs []*net.TCPAddr

	c            net.Conn
	mainConn     net.Conn
	scatterConns []net.Conn

	stopCh chan struct{}
}

// NewScatter creates Scatter used to pass messages between conn and
// connections created by mainAddr and scatterAddrs.
func NewScatter(conn net.Conn, mainAddr string, scatterAddrs []string) (*Scatter, error) {
	if conn == nil {
		return nil, errors.New("conn can't be empty")
	}

	mAddr, err := net.ResolveTCPAddr("tcp", mainAddr)
	if err != nil {
		return nil, errors.New("mainAddr identified as tcp address is invalid")
	}

	sAddrs := make([]*net.TCPAddr, 0, len(scatterAddrs))
	for _, scatterAddr := range scatterAddrs {
		if sAddr, err := net.ResolveTCPAddr("tcp", scatterAddr); err == nil {
			sAddrs = append(sAddrs, sAddr)
		}
	}

	return &Scatter{
		mainAddr:     mAddr,
		scatterAddrs: sAddrs,
		c:            conn,
		stopCh:       make(chan struct{}),
	}, nil
}

// Run the scatter and pass data
func (s *Scatter) Run() error {
	// dial mainAddr
	mConn, err := net.DialTCP("tcp", nil, s.mainAddr)
	if err != nil {
		return fmt.Errorf("tcp dial to the main address fail: %v", err)
	}
	s.mainConn = mConn

	// dial all of scatterAddrs
	s.scatterConns = make([]net.Conn, 0, len(s.scatterAddrs))
	for _, scatterAddr := range s.scatterAddrs {
		if sConn, err := net.DialTCP("tcp", nil, scatterAddr); err == nil {
			s.scatterConns = append(s.scatterConns, sConn)
		}
	}

	// copy data from in connection to main and scatter connections
	go func() {
		bytes := make([]byte, 512)
		for {
			reader := bufio.NewReader(s.c)
			n, err := reader.Read(bytes)
			if err != nil {
				if err == io.EOF {
					break
				}
				continue
			}

			// write bytes to main connection
			s.mainConn.Write(bytes[:n])

			// write bytes to all of scatter connections
			for _, scatterConn := range s.scatterConns {
				scatterConn.Write(bytes[:n])
			}
		}

		s.stopCh <- struct{}{}
	}()

	// copy data from main connection to in connection
	go func() {
		bytes := make([]byte, 512)
		for {
			reader := bufio.NewReader(s.mainConn)
			n, err := reader.Read(bytes)
			if err != nil {
				if err == io.EOF {
					break
				}
				continue
			}

			// write bytes to in connection
			s.c.Write(bytes[:n])
		}

		s.stopCh <- struct{}{}
	}()

	// drop all data from scatter connections
	go func() {
		readers := make([]io.Reader, 0, len(s.scatterConns))
		for _, sConn := range s.scatterConns {
			readers = append(readers, sConn)
		}
		smr := io.MultiReader(readers...)
		io.Copy(ioutil.Discard, smr)
	}()

	// close connections
	<-s.stopCh
	s.c.Close()
	s.mainConn.Close()
	for _, sConn := range s.scatterConns {
		sConn.Close()
	}

	color.Blue("disconnect from %v\n", s.c.RemoteAddr())

	return nil
}
