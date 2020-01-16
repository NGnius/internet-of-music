// Created by NGnius 2020-01-05

package main

import (
	"flag"
	"fmt"
	"runtime"
	"time"
)

const (
	DefaultPort             = "8080"
	DefaultBuffer           = time.Second / 10
	DefaultMaxMemory int64  = 1024 * 1024 * 256 // 256 Mb
	DefaultRootPath         = "."
	DefaultSampleRate int64 = 48000
	DefaultQuality int      = 4
)

var (
	Port       string
	Buffer     time.Duration
	MaxMemory  int64
	RootPath   string
	SampleRate int64
	Quality    int
	Version    bool
	Debug      bool
)

func initCommandLineArgs() {
	flag.StringVar(&Port, "port", DefaultPort, "Port to listen on")
	flag.DurationVar(&Buffer, "buffer", DefaultBuffer, "Audio buffer length")
	flag.Int64Var(&MaxMemory, "memory", DefaultMaxMemory, "Maximum memory, per request")
	flag.StringVar(&RootPath, "root", DefaultRootPath, "Root working directory")
	flag.BoolVar(&Version, "version", false, "Print version information and exit")
	flag.Int64Var(&SampleRate, "sample", DefaultSampleRate, "Sample rate to output")
	flag.IntVar(&Quality, "quality", DefaultQuality, "Resampling quality; higher number = higher quality & CPU usage")
    flag.BoolVar(&Debug, "debug", false, "Enable debug endpoints & logging")
}

func processCommandLineArgs() {
	flag.Parse()
}

func VersionString() string {
    return "IoM v"+CurrentVersion
}

func printDebugVersionInfo() {
	fmt.Printf("Internet of Music v%s\nGo v%s\nCreated & maintained by ", CurrentVersion, runtime.Version()[2:])
    fmt.Println(Maintainers)
}
