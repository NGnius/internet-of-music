// Created by NGnius 2020-01-05

package main

import (
  "path/filepath"
  "fmt"
  "io"
  "net/http"
  "os"
  "runtime"
  "time"
)

var (
    Requests int64 = 0
)

func handleChores(w http.ResponseWriter, r *http.Request) {
	Requests++
}

func debugHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Debug Handler called")
	handleChores(w, r)
	fmt.Fprintf(w, "Go version: %s\nRequests: %d\nUptime: %s", runtime.Version(), Requests, time.Since(StartTime).String())
}

func htmlHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("HTML Handler called")
	handleChores(w, r)
    urlPath := r.URL.Path
    if urlPath == "" || urlPath == "/" {
        urlPath = "index.html"
    }
	file, err := os.Open(filepath.Join(RootPath, "html", urlPath))
	if err != nil {
        w.WriteHeader(404)
        fmt.Printf("404 error while opening %s :: %s\n", filepath.Join(RootPath, "html", urlPath),  err)
		fmt.Fprintf(w, "Unable to open %s\n%s", filepath.Join("html", urlPath), err)
		return
	}
	io.Copy(w, file)
}

func musicHandler(w http.ResponseWriter, r *http.Request) {
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
	handleChores(w, r)
	PlayerInst.Next()
}

func previousHandler(w http.ResponseWriter, r *http.Request) {
	handleChores(w, r)
	PlayerInst.Previous()
}
