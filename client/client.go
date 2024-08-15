package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"github.com/gorilla/websocket"
)

func SendMsg(msg string) {
	jsonData, _ := json.Marshal(map[string]string{
		"msg": msg,
	})
	req, err := http.NewRequest("POST", "http://localhost:8001/send-msg", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error creating request:", err)
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	_, reqerr := client.Do(req)
	if reqerr != nil {
		fmt.Println("Error sending request:", reqerr)
	}

}
func main() {
	c, _, err := websocket.DefaultDialer.Dial("ws://localhost:8000", nil)
	if err != nil {
		log.Fatal("Dial Error:", err)
	}
	defer c.Close()

	// Channel to capture interrupt signals for graceful shutdown
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Channel to handle messages from the server
	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("Read Error:", err)
				return
			}
			log.Printf("\n%s\n", message)
		}
	}()

	fmt.Print("Type your message ... \n")

	go func() {
		for {
			mReader := bufio.NewReader(os.Stdin)
			str, readerr := mReader.ReadString('\n')
			userMsg := strings.TrimSpace(str)
			if readerr != nil {
				log.Printf("Unable to read message : %v", readerr.Error())
			} else {
				// make http request to message service
				SendMsg(userMsg)
			}
		}
	}()

	for {
		select {
		case <-done:
			return
		case <-interrupt:
			log.Println("Interrupt received, closing connection...")
			// Cleanly close the WebSocket connection
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("Close Error:", err)
				return
			}
			return
		}
	}

}
