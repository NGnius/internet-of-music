// Created by NGnius 2019-12-31

package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
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
	MaxMemory int64 = 1024*1024*512 // 512 MB
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
	http.HandleFunc("/next", nextHandler)
	http.HandleFunc("/previous", previousHandler)
	fmt.Println("Server initialised in " + time.Since(StartTime).String())
}

func main() {
	// TODO: run server
	fmt.Println("Server starting")
	fmt.Println(http.ListenAndServe("127.0.0.1:"+Port, nil))
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
	file, err := os.Open("html/index.html")
	if err != nil {
		r.Response.StatusCode = 500
		return
	}
	io.Copy(w, file)
}

func musicHandler(w http.ResponseWriter, r *http.Request) {
	// TODO
	fmt.Println("Music Handler called")
	if (r.Method != "POST") {
		fmt.Println("Non-POST request ignored")
		return
	}
	isForm :=  r.ParseMultipartForm(MaxMemory) == nil
	if (isForm) {
		fmt.Println("Handling form-encoded files")
		// handle form files
		for key := range r.MultipartForm.File {
			file, _, err := r.FormFile(key)
			if err == nil {
				fmt.Println("Queuing new file")
				PlayerInst.Enqueue(file)
			} else {
				fmt.Println("Key "+key+" is not a file")
			}
		}
	} else {
		fmt.Println("Handling JSON-encoded files")
		// TODO: handle json files
	}
	handleChores(w, r)
}

func playHandler(w http.ResponseWriter, r *http.Request) {
	handleChores(w, r)
	PlayerInst.Play()
}

func pauseHandler(w http.ResponseWriter, r *http.Request) {
	handleChores(w, r)
	PlayerInst.Pause()
}

func nextHandler(w http.ResponseWriter, r *http.Request) {
	// TODO
	handleChores(w, r)
	PlayerInst.Next()
}

func previousHandler(w http.ResponseWriter, r *http.Request) {
	// TODO
	handleChores(w, r)
	PlayerInst.Previous()
}
