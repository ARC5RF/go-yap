package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"

	"github.com/ARC5RF/go-blame"
	"github.com/ARC5RF/go-yap"
)

func must(fh *os.File, fe error) *os.File {
	if fe != nil {
		panic(blame.O0(fe))
	}
	return fh
}

func where(file string) string {
	_, here, _, _ := runtime.Caller(0)
	if _, s_err := os.Stat(here); s_err != nil {
		wd, _ := os.Getwd()
		here = wd
	}
	return filepath.Join(filepath.Dir(here), file)
}

var logfile = must(os.OpenFile(where("go_yap_demo.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm))
var console = yap.New(logfile, os.Stdout, 128, 100, true, false)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", homePage)
	mux.HandleFunc("/ws", wsEndpoint)

	console.Websocket.On("client.foo", func(b []byte) error {
		console.Log(string(b))
		return nil
	})

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt

	fmt.Println()

	logfile.Close()
}
