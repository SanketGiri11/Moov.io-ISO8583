package main

import (
	"net/http"

	"go.elastic.co/apm/module/apmhttp"

	"main.go/serviceimpl"
)

// type ClientManager struct {
// 	clients    map[*Client]bool
// 	broadcast  chan []byte
// 	register   chan *Client
// 	unregister chan *Client
// }

func main() {

	// flagMode := flag.String("mode", "client", "start in client or server mode")
	// flag.Parse()
	//if strings.ToLower(*flagMode) == "client" {
	//startClientMode()
	//}
	// router.HandleFunc("/posts", getPosts).Methods("GET")
	mux := serviceimpl.NewRouter()
	//http.Handle("/", mux)
	http.ListenAndServe(":8500", apmhttp.Wrap(mux))

}

// func (manager *ClientManager) start() {
// 	for {
// 		select {
// 		case connection := <-manager.register:
// 			manager.clients[connection] = true
// 			fmt.Println("Added new connection!")
// 		case connection := <-manager.unregister:
// 			if _, ok := manager.clients[connection]; ok {
// 				close(connection.data)
// 				delete(manager.clients, connection)
// 				fmt.Println("A connection has terminated!")
// 			}
// 		case message := <-manager.broadcast:
// 			for connection := range manager.clients {
// 				select {
// 				case connection.data <- message:
// 				default:
// 					close(connection.data)
// 					delete(manager.clients, connection)
// 				}
// 			}
// 		}
// 	}
// }

// func (manager *ClientManager) receive(client *Client) {
// 	for {
// 		message := make([]byte, 4096)
// 		length, err := client.socket.Read(message)
// 		if err != nil {
// 			manager.unregister <- client
// 			client.socket.Close()
// 			break
// 		}
// 		if length > 0 {
// 			fmt.Println("RECEIVED: " + string(message))
// 			manager.broadcast <- message
// 		}
// 	}
// }

// func (manager *ClientManager) send(client *Client) {
// 	defer client.socket.Close()
// 	for {
// 		select {
// 		case message, ok := <-client.data:
// 			if !ok {
// 				return
// 			}
// 			client.socket.Write(message)
// 		}
// 	}
// }

// func startServerMode() {
// 	fmt.Println("Starting server...")
// 	listener, error := net.Listen("tcp", ":8500")
// 	if error != nil {
// 		fmt.Println(error)
// 	}
// 	fmt.Println("Connection succesful")
// 	manager := ClientManager{
// 		clients:    make(map[*Client]bool),
// 		broadcast:  make(chan []byte),
// 		register:   make(chan *Client),
// 		unregister: make(chan *Client),
// 	}
// 	go manager.start()
// 	fmt.Println("Connection succesful manager")
// 	for {
// 		connection, _ := listener.Accept()
// 		if error != nil {
// 			fmt.Println(error)
// 		}
// 		client := &Client{socket: connection, data: make(chan []byte)}
// 		manager.register <- client
// 		go manager.receive(client)
// 		go manager.send(client)
// 		fmt.Println("Connection succesful server")
// 	}

// }
