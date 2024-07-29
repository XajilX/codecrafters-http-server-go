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
	if len(reql) != 3 {
		fmt.Println("Error parsing request: ", err.Error())
		os.Exit(1)
	}
	meth, path, prot := reql[0], reql[1], reql[2]
	if meth != "GET" {
		conn.Write([]byte(prot + " 404 Not Found\r\n\r\n"))
		os.Exit(0)
	}
	path_seg := strings.Split(path, "/")
	if len(path_seg) == 1 {
		conn.Write([]byte(prot + " 200 OK\r\n\r\n"))
	} else if len(path_seg) == 3 && path_seg[1] == "echo" {
		str := path_seg[2]
		conn.Write([]byte(prot + " 200 OK\r\n"))
		conn.Write([]byte("Content-Type: text/plain\r\n"))
		conn.Write([]byte(fmt.Sprintf("Content-Length: %d\r\n\r\n", len(path_seg[2]))))
		conn.Write([]byte(str))
	} else {
		conn.Write([]byte(prot + " 404 Not Found\r\n\r\n"))
	}
}
