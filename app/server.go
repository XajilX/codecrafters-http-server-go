package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"net"
	"os"
	"slices"
	"strconv"
	"strings"
)

func Resp_text_plain(prot, s string, comp_type string) []byte {
	var bs []byte
	is_comp := false
	for _, comp_t := range strings.Split(comp_type, ",") {
		comp_t = strings.Trim(comp_t, " ")
		switch comp_t {
		case "gzip":
			bs = Compress_gzip(s)
			comp_type = comp_t
			is_comp = true
		default:
			bs = []byte(s)
		}
		if is_comp {
			break
		}
	}
	c_enc := ""
	if is_comp {
		c_enc = fmt.Sprintf("Content-Encoding: %s\r\n", comp_type)
	}
	return slices.Concat([]byte(fmt.Sprintf(
		"%s 200 OK\r\n%sContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n",
		prot,
		c_enc,
		len(s),
	)), bs)
}

func Compress_gzip(s string) []byte {
	bs := make([]byte, len(s))
	wr := bytes.NewBuffer(bs)
	gzip_wr := gzip.NewWriter(wr)
	gzip_wr.Write([]byte(s))
	gzip_wr.Close()
	return bs
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
		Handler_post(conn, prot, path, reqs)
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
		comp := ""
		for _, s := range reqs {
			if strings.HasPrefix(s, "Accept-Encoding: ") {
				comp, _ = strings.CutPrefix(s, "Accept-Encoding: ")
			}
		}
		str := path_seg[2]
		conn.Write(Resp_text_plain(prot, str, comp))
	} else if len_seg == 2 && path_seg[1] == "user-agent" {
		ua := ""
		for _, s := range reqs {
			if strings.HasPrefix(s, "User-Agent: ") {
				ua, _ = strings.CutPrefix(s, "User-Agent: ")
			}
		}
		conn.Write(Resp_text_plain(prot, ua, ""))
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
			if strings.HasPrefix(s, "Content-Length: ") {
				str_len_body, _ := strings.CutPrefix(s, "Content-Length: ")
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
