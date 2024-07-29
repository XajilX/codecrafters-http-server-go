package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func Resp_text_plain(prot, s string) string {
	return fmt.Sprintf(
		"%s 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s",
		prot,
		len(s),
		s,
	)
}

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
		conn.Write([]byte(Resp_text_plain(prot, str)))
	} else if len_seg == 2 && path_seg[1] == "user-agent" {
		conn.Write([]byte(Resp_text_plain(prot, ua)))
	} else {
		conn.Write([]byte(prot + " 404 Not Found\r\n\r\n"))
	}
}
