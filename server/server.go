// Created by NGnius 2019-12-31

package iomserv

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

const (
	CurrentVersion = "0.0.0.2"
)

var (
	StartTime     = time.Now()
	PlayerInst *Player
)

func init() {
	// parse command line arguments
	initCommandLineArgs()
	processCommandLineArgs()
	if Version {
		printVersionInfo()
		os.Exit(0)
	}
	// init server
	PlayerInst = NewPlayer()
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
	// run server
	fmt.Println("Server starting")
	fmt.Println(http.ListenAndServe(":"+Port, nil))
	fmt.Println("Server stopped after " + time.Since(StartTime).String())
}
