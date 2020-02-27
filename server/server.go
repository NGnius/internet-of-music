// Created by NGnius 2019-12-31

package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"
)

const (
	CurrentVersion = "0.0.0.4"
)

var (
	Maintainers = []string{"NGnius"}
	StartTime   = time.Now()
	PlayerInst  *Player
	HandlerMux  *http.ServeMux
	Server      *http.Server
)

func Initialize() {
	// handle interrupt (terminate) signal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		s := <-signalChan
		fmt.Println("Received terminate signal " + s.String())
		Exit()
	}()
	// parse command line arguments
	initCommandLineArgs()
	processCommandLineArgs()
	if Version {
		printDebugVersionInfo()
		os.Exit(0)
	}
	fmt.Printf("Version: %s\n", VersionString())
	// init server
	PlayerInst = NewPlayer()
	PlayerInst.Init()
	fmt.Println("Server initialising")
	HandlerMux = http.NewServeMux()
	HandlerMux.HandleFunc("/", htmlHandler)
	HandlerMux.HandleFunc("/music", musicHandler)
	HandlerMux.HandleFunc("/play", playHandler)
	HandlerMux.HandleFunc("/pause", pauseHandler)
	HandlerMux.HandleFunc("/next", nextHandler)
	HandlerMux.HandleFunc("/previous", previousHandler)
	if Debug {
		HandlerMux.HandleFunc("/exit", exitHandler)
		HandlerMux.HandleFunc("/debug", debugHandler)
	}
	Server = &http.Server{
		Addr:    ":" + Port,
		Handler: HandlerMux,
	}
	fmt.Println("Server initialised in " + time.Since(StartTime).String())
}

func Run() {
	// run server
	fmt.Println("Server starting")
	fmt.Println(Server.ListenAndServe())
	fmt.Println("Server stopped after " + time.Since(StartTime).String())
}

func Exit() {
	Server.Close()
}

func main() {
	Initialize()
	Run()
}

func init() {
}
