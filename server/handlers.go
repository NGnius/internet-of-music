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
    if Debug {
        fmt.Println("HTML Handler called")
    }
	handleChores(w, r)
    urlPath := r.URL.Path
    if urlPath[len(urlPath)-1:] == "/" {
        urlPath = urlPath[:len(urlPath)-1]
    }
    if urlPath == "" || urlPath == "/" {
        urlPath = "index.html"
    }
    if urlPath[len(urlPath)-5:] != ".html" {
        w.WriteHeader(400)
        fmt.Printf("400 error while loading %s\n", urlPath)
        fmt.Fprintf(w, "HTTP 400: Cannot access non-HTML resource %s\n", urlPath)
        return
    }
	file, err := os.Open(filepath.Join(RootPath, "html", urlPath))
	if err != nil {
        w.WriteHeader(404)
        fmt.Printf("404 error while loading %s :: %s\n", filepath.Join(RootPath, "html", urlPath),  err)
		fmt.Fprintf(w, "HTTP 404: Unable to find HTML resource %s\n%s\n", filepath.Join("html", urlPath), err)
		return
	}
	io.Copy(w, file)
}

func musicHandler(w http.ResponseWriter, r *http.Request) {
	if Debug {
        fmt.Println("Music Handler called")
    }
    handleChores(w, r)
	if (r.Method != "POST") {
        w.WriteHeader(405)
        fmt.Fprintf(w, "HTTP 405: Only POST operations are allowed to /music\n")
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
		fmt.Println("(NOT) Handling JSON-encoded files")
        w.WriteHeader(400)
        fmt.Fprintf(w, "HTTP 400: Only form-encoded music is currently supported")
		// TODO: handle json files
        return
	}
    w.WriteHeader(204)
}

func playHandler(w http.ResponseWriter, r *http.Request) {
	handleChores(w, r)
	PlayerInst.Play()
    w.WriteHeader(204)
}

func pauseHandler(w http.ResponseWriter, r *http.Request) {
	handleChores(w, r)
	PlayerInst.Pause()
    w.WriteHeader(204)
}

func nextHandler(w http.ResponseWriter, r *http.Request) {
	handleChores(w, r)
	PlayerInst.Next()
    w.WriteHeader(204)
}

func previousHandler(w http.ResponseWriter, r *http.Request) {
	handleChores(w, r)
	PlayerInst.Previous()
    w.WriteHeader(204)
}

func exitHandler(w http.ResponseWriter, r *http.Request) {
    handleChores(w, r)
    w.WriteHeader(204)
    Server.Close()
}
