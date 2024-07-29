package main

import (
	"fmt"
	"net"
	"os"
	"slices"
	"strconv"
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
	if meth == "GET" {
		Handler_get(conn, prot, path, reqs)
	} else if meth == "POST" {

	} else {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
	conn.Close()
}

func Handler_get(conn net.Conn, prot, path string, reqs []string) {
	path_seg := strings.Split(path, "/")
	len_seg := len(path_seg)
	if len_seg == 2 && path_seg[1] == "" {
		conn.Write([]byte(prot + " 200 OK\r\n\r\n"))
	} else if len_seg == 3 && path_seg[1] == "echo" {
		str := path_seg[2]
		conn.Write(Resp_text_plain(prot, str))
	} else if len_seg == 2 && path_seg[1] == "user-agent" {
		ua := ""
		for _, s := range reqs {
			if strings.HasPrefix(s, "User-Agent: ") {
				ua, _ = strings.CutPrefix(s, "User-Agent: ")
			}
		}
		conn.Write(Resp_text_plain(prot, ua))
	} else if len_seg == 3 && path_seg[1] == "files" {
		conn.Write(Resp_file(prot, path_seg[2]))
	} else {
		conn.Write([]byte(prot + " 404 Not Found\r\n\r\n"))
	}
}

func Handler_post(conn net.Conn, prot, path string, reqs []string) {
	path_seg := strings.Split(path, "/")
	len_seg := len(path_seg)
	if len_seg == 3 && path_seg[1] == "files" {
		len_body := 0
		for _, s := range reqs {
			if strings.HasPrefix(s, "User-Agent: ") {
				str_len_body, _ := strings.CutPrefix(s, "User-Agent: ")
				len_body, _ = strconv.Atoi(str_len_body)
			}
		}

		dir := os.Args[2]
		file, err := os.Create(dir + path_seg[2])
		if err != nil {
			fmt.Println("Error creating file")
			conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
			return
		}
		defer file.Close()
		_, err = file.Write([]byte(reqs[len(reqs)-1])[:len_body])
		if err != nil {
			fmt.Println("Error writing file")
			conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
			return
		}
		conn.Write([]byte("HTTP/1.1 201 Created\r\n\r\n"))
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
