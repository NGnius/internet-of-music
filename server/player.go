// Created by NGnius 2020-01-05

package main

import (
	"bytes"
	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/flac"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/vorbis"
	"github.com/faiface/beep/wav"
	"io"
	"io/ioutil"
	"os"
	"time"
)

type Player struct {
	streamer        beep.Streamer
	format          beep.Format
	queue           []ReadSeekerCloser
	queueIndex      int
	queueIsComplete bool
	Config          PlayerConfig
	songDone        chan bool
	isSpeakerLocked bool
	isSpeakerInited bool
}

func NewPlayer() (p *Player) {
	p = &Player{
		Config: PlayerConfig{
			BufferedTime: time.Second/100,
		},
	}
	return
}

func (p *Player) Init() {
	p.songDone = make(chan bool)
	p.queueIndex = -1
	p.queue = []ReadSeekerCloser{}
	p.queueIsComplete = true
	// testing
	f, _ := os.Open("/home/ngnius/Music/MusicMP3/5 Seconds Of Summer/Ghostbusters/Girls_Talk_Boys.mp3")
	p.Enqueue(f)
}

func (p *Player) Enqueue(audioFile ReadSeekerCloser) {
	p.queue = append(p.queue, audioFile)
}

func (p *Player) Play() {
	if p.isSpeakerLocked {
		p.isSpeakerLocked = false
		speaker.Unlock()
	}
	if p.queueIsComplete {
		go p.handleSongEnd()
		p.songDone <- true
	}
}

func (p *Player) Pause() {
	if !p.isSpeakerLocked {
		p.isSpeakerLocked = true
		speaker.Lock()
	}
}

func (p *Player) handleSongEnd() {
	fmt.Println("Starting queue handler")
	handlerLoop:
	for {
		next := <-p.songDone
		if len(p.queue) > (p.queueIndex + 1) {
			p.queueIsComplete = false
			if next {
				p.queueIndex++
			}
			fmt.Printf("Now playing index %d\n", p.queueIndex)
			var decodeErr error
			p.streamer, p.format, decodeErr = decodeAudioFile(p.queue[p.queueIndex])
			if decodeErr != nil {
				fmt.Println(decodeErr)
			}
			if (!p.isSpeakerInited) {
				fmt.Println("Initialising speaker")
				p.isSpeakerInited = true
				speaker.Init(p.format.SampleRate, p.format.SampleRate.N(p.Config.BufferedTime))
			}
			fmt.Println("Playing audio")
			speaker.Play(beep.Seq(p.streamer, beep.Callback(func() { p.songDone <- true })))
			fmt.Println("Song end handling done")
		} else {
			fmt.Println("Queue finished, shutting down queue handler")
			p.queueIsComplete = true
			break handlerLoop
		}
	}
}

func decodeAudioFile(f ReadSeekerCloser) (streamer beep.Streamer, format beep.Format, decodeErr error) {
	var data []byte
	data, decodeErr = ioutil.ReadAll(f)
	if decodeErr != nil {
		return
	}
	mime := detectAudioType(data)
	fmt.Println("File decoded as " + mime)
	f.Seek(0, 0)
	switch mime {
	case "audio/mp3":
		streamer, format, decodeErr = mp3.Decode(f)
	case "audio/wav":
		streamer, format, decodeErr = wav.Decode(f)
	case "audio/vorbis":
		streamer, format, decodeErr = vorbis.Decode(f)
	case "audio/flac":
		streamer, format, decodeErr = flac.Decode(f)
	}
	return
}

func detectAudioType(data []byte) string {
	if bytes.Compare(data[0:4], []byte{102, 76, 97, 67}) == 0 { // spells fLaC
		//fmt.Println("flac")
		return "audio/flac"
	}
	if bytes.Compare(data[0:2], []byte{73, 68}) == 0 {
		//fmt.Println("mp3")
		return "audio/mp3"
	}
	if bytes.Compare(data[0:4], []byte{82, 73, 70, 70}) == 0 && bytes.Compare(data[8:12], []byte{87, 65, 86, 69}) == 0 { // spells RIFF and WAVE, respectively
		//fmt.Println("wav")
		return "audio/wav"
	}
	if bytes.Compare(data[29:35], []byte{118, 111, 114, 98, 105, 115}) == 0 { // spells vorbis
		//fmt.Println("vorbis")
		return "audio/vorbis"
	}
	//fmt.Println("No format detected")
	return "?"
}

type PlayerConfig struct {
	BufferedTime time.Duration
}

type ReadSeekerCloser interface {
	io.Reader
	io.Seeker
	io.Closer
}
