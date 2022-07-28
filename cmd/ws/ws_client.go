package main

import (
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

var done chan interface{}
var interrupt chan os.Signal

func receiveHandler(connection *websocket.Conn) {
	defer close(done)

	for {
		_, msg, err := connection.ReadMessage()
		if err != nil {
			log.Fatal("Error in receive: ", err)
			return
		}

		log.Printf("Received: %s\n", msg)
	}
}

func main() {
	done = make(chan interface{})
	interrupt = make(chan os.Signal)

	signal.Notify(interrupt, os.Interrupt) // Notify the interrupt channel for SIGINT

	// socketUrl := "ws://localhost:8080" + "/socket" //* URL for websocket connection

	channelId := "30315311-56ba-484d-b500-c1e08305511f"
	// channelID := "85968b60-4cce-457a-a05c-21119bf9ad20"

	// thingKey := "120ce059-0a8b-4fe7-8db9-85bdd1d3aece"
	thingKey := "c02ff576-ccd5-40f6-ba5f-c85377aad529"

	socketUrl := "ws://localhost:8190/channels/" + channelId + "/messages/?authorization=" + thingKey
	// socketUrl := "ws://localhost:5937/metrics/"
	// fmt.Println(socketUrl)

	// socketUrl := "ws://localhost:5937"

	conn, _, err := websocket.DefaultDialer.Dial(socketUrl, nil)
	if err != nil {
		log.Fatal("Error connecting to Websocket Server: ", err)
	} else {
		log.Println("Connected to the ws adapter")
	}

	defer conn.Close() //todo: Close the connection before exiting main goroutine

	go receiveHandler(conn)

	for {
		select {

		case <-interrupt:
			log.Println("Interrupt occured, closing the connection...")
			conn.Close()
			err := conn.WriteMessage(websocket.TextMessage, []byte("closed this ws client just now"))
			if err != nil {
				log.Println("Error during closing websocket: ", err)
				return
			}

			select {
			case <-done: // Close channel will give 0
				log.Println("Receiver Channel Closed! Exiting...")

			case <-time.After(time.Duration(1) * time.Second): // Did not receive anything from done channel
				log.Println("Timeout in closing receiving channel. Exiting...")
			}
			return
		}
	}
}
