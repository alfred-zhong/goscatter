package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
)

const usage = `usage: _listen port`

func main() {
	if len(os.Args) < 2 {
		fmt.Println(usage)
		os.Exit(1)
	}

	port, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "port: %s is invalid", os.Args[1])
		os.Exit(2)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(3)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			break
		}

		go func() {
			fmt.Printf("connect from %v\n", conn.RemoteAddr())
			io.Copy(os.Stdout, conn)

			fmt.Printf("\ndisconnect from %v\n", conn.RemoteAddr())
			conn.Close()
		}()
	}
}
