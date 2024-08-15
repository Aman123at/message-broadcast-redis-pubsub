package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

var redisClient *redis.Client

func PublishMsg(msg string) {
	// defer redisClient.Close()
	err := redisClient.Publish(context.Background(), "user_msgs", msg).Err()
	if err != nil {
		log.Println("Error publishing message:", err)
	} else {
		log.Println("Message published to user_msgs channel")
	}

}

type Message struct {
	Msg string
}

func handleMessageSend(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var reqmsg Message
	decodeerr := json.NewDecoder(r.Body).Decode(&reqmsg)
	if decodeerr != nil {
		http.Error(w, "Error decoding message", http.StatusBadRequest)
	}
	// publish message to redis
	PublishMsg(reqmsg.Msg)
	json.NewEncoder(w).Encode(map[string]string{
		"msg": "Message sent successfully",
	})
}

func main() {
	log.Println("Welcome to message service")
	redisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	router := mux.NewRouter()
	router.HandleFunc("/send-msg", handleMessageSend).Methods("POST")
	log.Fatal(http.ListenAndServe(":8001", router))
}
