// Created by NGnius 2019-12-31

package main

import (
	"fmt"
	"net/http"
	"runtime"
	"time"
)

const (
	DefaultPort = "8080"
)

var (
	Port          = DefaultPort
	Requests  int = 0
	StartTime     = time.Now()
	PlayerInst    = NewPlayer()
)

func init() {
	// TODO: init server
	PlayerInst.Init()
	fmt.Println("Server initialising")
	http.HandleFunc("/debug", debugHandler)
	http.HandleFunc("/", htmlHandler)
	http.HandleFunc("/music", musicHandler)
	http.HandleFunc("/play", playHandler)
	http.HandleFunc("/pause", pauseHandler)
	fmt.Println("Server initialised in " + time.Since(StartTime).String())
}

func main() {
	// TODO: run server
	fmt.Println("Server starting")
	fmt.Println(http.ListenAndServe(":"+Port, nil))
	fmt.Println("Server stopped after " + time.Since(StartTime).String())
}

func handleChores(w http.ResponseWriter, r *http.Request) {
	Requests++
}

func debugHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Debug Handler called")
	handleChores(w, r)
	fmt.Fprintf(w, "Go version: %s\nRequests: %d\nUptime: %s", runtime.Version(), Requests, time.Since(StartTime).String())
}

func htmlHandler(w http.ResponseWriter, r *http.Request) {
	// TODO
	fmt.Println("HTML Handler called")
	handleChores(w, r)
	//fmt.Fprintf(w, "<html>\n<header>\n<title>WIP</title>\n</header>\n<body><h1>Nothing to see here</h1></body>\n</html>")
	debugHandler(w, r)
}

func musicHandler(w http.ResponseWriter, r *http.Request) {
	// TODO
	fmt.Println("Music Handler called")
	handleChores(w, r)
}

func playHandler(w http.ResponseWriter, r *http.Request) {
	// TODO
	fmt.Println("Play Handler called")
	handleChores(w, r)
	PlayerInst.Play()
}

func pauseHandler(w http.ResponseWriter, r *http.Request) {
	// TODO
	fmt.Println("Pause Handler called")
	handleChores(w, r)
	PlayerInst.Pause()
}
