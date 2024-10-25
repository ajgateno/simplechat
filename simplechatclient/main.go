package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/url"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", ":8080", "websocket connection address")

func main() {
	flag.Parse()
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Enter a username: ")
	scanner.Scan()
	username := strings.Trim(scanner.Text(), " \n\t")

	headers := make(http.Header)
	headers.Add("x-username", username)
	c, _, err := websocket.DefaultDialer.Dial(u.String(), headers)
	if err != nil {
		fmt.Println("Failed to create websocket connection")
		os.Exit(1)
	}
	defer c.Close()

	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				return
			}
			fmt.Println(string(message))
		}
	}()

	for {
		scanner.Scan()
		message := strings.Trim(scanner.Text(), " \n\t")
		c.WriteMessage(websocket.TextMessage, []byte(message))
	}
}
