package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type GithubPush struct {
	Repository RepoInfo `json:"repository"`
}

type RepoInfo struct {
	Git_url string `json:"git_url"`
	Url     string `json:"url"`
	Name    string `json:"name"`
}

var cmd *exec.Cmd = nil

func handleConnection(conn net.Conn) {

	buffer := make([]byte, 16484)

	defer conn.Close()
	var hlen int = -1
	var contentLength int
	read := 0

	for read == 0 || read < hlen+contentLength {

		current, err := conn.Read(buffer[read:])
		read += current
		fmt.Println("read bytes:", read)

		if err != nil {
			fmt.Println("read error:", err.Error())
		}

		if hlen == -1 {
			for i := 3; i < read; i++ {
				if buffer[i-3] == 13 && buffer[i-2] == 10 && buffer[i-1] == 13 && buffer[i] == 10 {
					hlen = i
					break
				}
			}

			header := string(buffer[:hlen-3])
			headers := strings.Split(header, "\r\n")

			for _, h := range headers {

				if strings.Index(h, "Content-Length") >= 0 {
					//fmt.Println(h)
					cl := strings.Split(h, ": ")
					contentLength, err = strconv.Atoi(cl[1])

					if err != nil {
						fmt.Println(err.Error())
					}

				}
			}

			fmt.Println("read", read, "hlen", hlen, "cl", contentLength)

		}

	}

	body := buffer[hlen+1 : hlen+1+contentLength]

	_, err := conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))

	if err != nil {
		fmt.Println("read error:", err.Error())
	}

	bmap := &GithubPush{}

	if err := json.Unmarshal(body, &bmap); err != nil {
		fmt.Println(err.Error())
	}

	var toClone bool = false

	if _, err := os.Stat(bmap.Repository.Name); err != nil {
		if os.IsNotExist(err) {
			toClone = true
		}

	}

	var cout []byte

	if cmd != nil {
		cmd.Process.Kill()
	}

	if toClone {
		cmd = exec.Command("git", "clone", bmap.Repository.Git_url)
	} else {
		cmd = exec.Command("git", "pull")
		cmd.Dir = bmap.Repository.Name
	}

	if cout, err = cmd.CombinedOutput(); err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println(string(cout))

	cmd = exec.Command("go", "build")
	cmd.Dir = bmap.Repository.Name

	if cout, err = cmd.CombinedOutput(); err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println(string(cout))

	cmd = exec.Command("./" + bmap.Repository.Name)
	cmd.Dir = bmap.Repository.Name

	err = cmd.Start()

	if err != nil {
		fmt.Println("error starting process", err.Error())
	}

}

func main() {

	ln, err := net.Listen("tcp", ":6776")
	if err != nil {
		// handle error
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
		}
		go handleConnection(conn)
	}

}
