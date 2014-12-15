package main

import (
	"bufio"
	"fmt"
	"github.com/go-martini/martini"
	"net"
	"strings"
	"time"
)

func main() {
	m := martini.Classic()
	m.Get("/:host", func(params martini.Params) string {
		nc, err := net.Dial("tcp", "localhost:6669")
		if err != nil {
			return "NS Down"
		}
		nc.Write([]byte("NICK TEST\r\nJOIN #Holla\r\n"))
		reader := bufio.NewReader(nc)
		results := make(chan string)
		go wait_for_hostname(reader, params["host"], results)
		select {
		case res := <-results:
			nc.Close()
			return res
		case <-time.After(time.Second * 11):
			nc.Close()
			return "Down"
		}
		return "wat"
	})
	m.RunOnAddr("0.0.0.0:1236")
}

func wait_for_hostname(reader *bufio.Reader, hostname string, output chan string) {
	timetodie := time.Now().Add(time.Second * 11)
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			output <- "NS ERR"
		}
		if strings.HasPrefix(string(line), fmt.Sprintf(":%s!~", hostname)) {
			output <- "Alive"
		}
		if time.Now().Unix() > timetodie.Unix() {
			output <- "Down"
			return
		}
	}
}
