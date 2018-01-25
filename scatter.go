package goscatter

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"sync"
)

type Scatter struct {
	mainAddr     *net.TCPAddr
	scatterAddrs []*net.TCPAddr

	c            net.Conn
	mainConn     net.Conn
	scatterConns []net.Conn
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

	return &Scatter{mainAddr: mAddr, scatterAddrs: sAddrs, c: conn}, nil
}

// Run the scatter and pass data
func (s *Scatter) Run() error {
	mConn, err := net.DialTCP("tcp", nil, s.mainAddr)
	if err != nil {
		return fmt.Errorf("tcp dial to the main address fail: %v", err)
	}
	s.mainConn = mConn

	// copy data between in connection and the main out connection
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		io.Copy(s.c, s.mainConn)
		wg.Done()
	}()
	go func() {
		io.Copy(s.mainConn, s.c)
		wg.Done()
	}()

	// copy data from in connection to scatter connections
	go func() {
		// dial all of scatterAddrs
		s.scatterConns = make([]net.Conn, 0, len(s.scatterAddrs))
		for _, scatterAddr := range s.scatterAddrs {
			if sConn, err := net.DialTCP("tcp", nil, scatterAddr); err == nil {
				s.scatterConns = append(s.scatterConns, sConn)
			}
		}

		// drop all data from scatter connections
		go func() {
			readers := make([]io.Reader, 0, len(s.scatterConns))
			for _, sConn := range s.scatterConns {
				readers = append(readers, sConn)
			}
			smr := io.MultiReader(readers...)
			io.Copy(ioutil.Discard, smr)
		}()

		// copy data from in connection to scatter connections
		writers := make([]io.Writer, 0, len(s.scatterConns))
		for _, sConn := range s.scatterConns {
			writers = append(writers, sConn)
		}
		smw := io.MultiWriter(writers...)
		io.Copy(smw, s.c)
	}()

	// Wait blocks until any of connection is closed
	wg.Wait()

	// close connections
	s.c.Close()
	s.mainConn.Close()
	for _, sConn := range s.scatterConns {
		sConn.Close()
	}

	return nil
}
