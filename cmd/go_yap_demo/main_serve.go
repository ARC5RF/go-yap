package main

import (
	"log"
	"net/http"
	"path/filepath"
	"runtime"

	"github.com/ARC5RF/go-yap"
	"github.com/gorilla/websocket"
)

// We'll need to define an Upgrader
// this will require a Read and Write buffer size
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func index() string {
	_, file_name, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file_name), "index.html")
}

func homePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, index())
}

type example_for_websocket struct{}

func (example *example_for_websocket) String() string { return "" }

func (example *example_for_websocket) TokenizeForWebsocket() ([]any, error) {
	return []any{yap.Red("for websocket")}, nil
}

func (example *example_for_websocket) TokenizeForTerminal() ([]any, error) {
	return []any{"for terminal"}, nil
}

func (example *example_for_websocket) TokenizeForFile() ([]any, error) {
	return []any{"for file"}, nil
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	// upgrade this connection to a WebSocket connection
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	// log.Println("Client Connected")

	console.Log(&example_for_websocket{}, yap.NL)
	console.Websocket.Add(ws, "yap.log")
	defer console.Websocket.Rem(ws)

	// listen indefinitely for new messages coming
	// through on our WebSocket connection
	// reader(ws)
	log.Println("Client Disconnected")
}
