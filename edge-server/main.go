package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all connections by default
	},
}
var wsClients []*websocket.Conn

// Handle WebSocket connections
func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	log.Printf("New Socket CONN >> %v", ws.RemoteAddr())
	wsClients = append(wsClients, ws)
	log.Println(wsClients)
	if err != nil {
		log.Fatal(err)
	}
}

func initSubscriber() {
	rClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	pubsub := rClient.Subscribe(context.Background(), "user_msgs")
	// defer pubsub.Close()
	_, err := pubsub.Receive(context.Background())
	if err != nil {
		fmt.Println("Error subscribing to channel:", err)
		return
	}

	// Go channel to receive messages
	ch := pubsub.Channel()

	// Process messages
	for msg := range ch {
		for _, client := range wsClients {
			// Send message to all connected clients
			mynewmsg := fmt.Sprintf("\nREPLY FROM EDGE SERVER >> %s\n", msg.Payload)
			err = client.WriteMessage(websocket.TextMessage, []byte(mynewmsg))
			if err != nil {
				log.Println("Write Error:", err)

			}
		}
	}
}

func main() {
	log.Println("EDGE server for Message Broadcast")

	http.HandleFunc("/", handleConnections)
	// subscribe for messages from redis pubsub
	go initSubscriber()
	log.Println("Server started on :8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe Error:", err)
	}
}
