// Created by NGnius 2020-01-05

package iomserv

import (
  "flag"
  "fmt"
  "runtime"
  "time"
)

const (
	DefaultPort = "8080"
  DefaultBuffer = time.Second/10
  DefaultMaxMemory int64 = 1024*1024*256 // 256 Mb
  DefaultRootPath = "."
)

var (
  Port string
  Buffer time.Duration
  MaxMemory int64
  RootPath string
  Version bool
)

func initCommandLineArgs() {
  flag.StringVar(&Port, "port", DefaultPort, "Port to listen on")
  flag.DurationVar(&Buffer, "buffer", DefaultBuffer, "Audio buffer length")
  flag.Int64Var(&MaxMemory, "memory", DefaultMaxMemory, "Maximum memory, per request")
  flag.StringVar(&RootPath, "root", DefaultRootPath, "Root working directory")
  flag.BoolVar(&Version, "version", false, "Print version information and exit")
}

func processCommandLineArgs() {
  flag.Parse()
  fmt.Println("Buffer set to "+Buffer.String())
}

func printVersionInfo() {
  fmt.Printf("Internet of Music v%s\nGo v%s\nCreated & maintained by NGnius\n", CurrentVersion, runtime.Version()[2:])
}
