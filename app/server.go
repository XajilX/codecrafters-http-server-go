package main

import (
	"fmt"
	"net"
	"os"
	"slices"
	"strings"
)

func Resp_text_plain(prot, s string) []byte {
	return []byte(fmt.Sprintf(
		"%s 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s",
		prot,
		len(s),
		s,
	))
}

func Resp_file(prot, path string) []byte {
	dir := os.Args[2]
	file, err := os.Open(dir + path)
	if err != nil {
		fmt.Println("Error opening file")
		return []byte("HTTP/1.1 404 Not Found\r\n\r\n")
	}
	defer file.Close()
	data := make([]byte, 1024)
	count, err := file.Read(data)
	if err != nil {
		fmt.Println("Error reading file")
		return []byte("HTTP/1.1 404 Not Found\r\n\r\n")
	}
	resp_head := []byte(fmt.Sprintf(
		"%s 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n",
		prot,
		count,
	))
	return slices.Concat(resp_head, data[:count])
}

func Handler(conn net.Conn) {
	req := make([]byte, 1024)
	conn.Read(req)
	reqs := strings.Split(string(req), "\r\n")
	reql := strings.Split(reqs[0], " ")
	if len(reql) != 3 {
		fmt.Println("Error parsing request")
		os.Exit(1)
	}
	meth, path, prot := reql[0], reql[1], reql[2]
	if meth != "GET" {
		conn.Write([]byte(prot + " 404 Not Found\r\n\r\n"))
		os.Exit(0)
	}
	ua := ""
	for _, s := range reqs {
		if strings.HasPrefix(s, "User-Agent: ") {
			ua, _ = strings.CutPrefix(s, "User-Agent: ")
		}
	}
	path_seg := strings.Split(path, "/")
	len_seg := len(path_seg)
	if len_seg == 2 && path_seg[1] == "" {
		conn.Write([]byte(prot + " 200 OK\r\n\r\n"))
	} else if len_seg == 3 && path_seg[1] == "echo" {
		str := path_seg[2]
		conn.Write(Resp_text_plain(prot, str))
	} else if len_seg == 2 && path_seg[1] == "user-agent" {
		conn.Write(Resp_text_plain(prot, ua))
	} else if len_seg == 3 && path_seg[1] == "files" {
		conn.Write(Resp_file(prot, path_seg[2]))
	} else {
		conn.Write([]byte(prot + " 404 Not Found\r\n\r\n"))
	}
}

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}
		go Handler(conn)
	}
}
