package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", ":8080", "websocket address")
var upgrader = websocket.Upgrader{}
var clients = make(map[*Client]bool)

type Client struct {
	conn *websocket.Conn
	name string
	queue chan []byte
	kill chan bool
}

func (c *Client) receive() {
	defer func() {
		fmt.Println(c.name + " left.")
		c.conn.Close()
		close(c.queue)
		c.kill <- true
	}()
	for {
		_, rawMessage, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		message := string(rawMessage)
		message = strings.Trim(message, " \n\t")
		message = c.name + ": " + message
		fmt.Println(message)
		c.queue <- []byte(message)
	}
}

func (c *Client) send() {
	for {
		select {
		case end := <-c.kill:
			if end {
				close(c.kill)
				clients[c] = false
			}
			return
		case message := <-c.queue:
			for oc, _ := range clients {
				oc.conn.WriteMessage(websocket.TextMessage, message)
			}
		}
	}
}

func main() {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println("Failed to serve connection")
			fmt.Println(err)
		}
		username := string(strings.Trim(r.Header.Get("x-username"), " \t\n"))
		fmt.Println("Welcome, " + username + "!")

		kill := make(chan bool)
		queue := make(chan []byte)
		c := Client{conn: conn, name: username, queue: queue, kill: kill}
		clients[&c] = true

		go c.receive()
		go c.send()
	})
	fmt.Println("Starting server")
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		fmt.Println("Failed to serve websocket connections")
		os.Exit(1)
	}
}
