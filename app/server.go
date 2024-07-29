package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	req := make([]byte, 1024)
	conn.Read(req)
	reqs := strings.Split(string(req), "\r\n")
	reql := strings.Split(reqs[0], " ")
	meth, path, prot := reql[0], reql[1], reql[2]
	fmt.Println("Processing...")
	if meth != "GET" || path != "/" {
		conn.Write([]byte(prot + " 404 Not Found\r\n\r\n"))
	} else {
		conn.Write([]byte(prot + " 200 OK\r\n\r\n"))
	}
}
