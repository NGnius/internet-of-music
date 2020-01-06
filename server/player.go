// Created by NGnius 2020-01-05

package iomserv

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
	//"os"
	"time"
)

type Player struct {
	streamer        beep.Streamer
	format          beep.Format
	control         *beep.Ctrl
	queue           []ReadSeekerCloser
	queueIndex      int
	queueIsComplete bool
	Config          PlayerConfig
	songDone        chan bool
	isPaused        bool
	isSpeakerInited bool
}

func NewPlayer() (p *Player) {
	p = &Player {
		Config: PlayerConfig{
			BufferedTime: Buffer,
		},
	}
	return
}

func (p *Player) Init() {
	p.songDone = make(chan bool)
	p.queueIndex = -1
	p.queue = []ReadSeekerCloser{}
	p.queueIsComplete = true
}

func (p *Player) Enqueue(audioFile ReadSeekerCloser) {
	p.queue = append(p.queue, audioFile)
}

func (p *Player) EnqueueMany(audioFiles ...ReadSeekerCloser) {
	p.queue = append(p.queue, audioFiles...)
}

func (p *Player) Play() {
	if p.isPaused {
		p.isPaused = false
		p.control.Paused = false
	}
	if p.queueIsComplete {
		go p.handleSongEnd()
		p.songDone <- true
	}
}

func (p *Player) Pause() {
	if !p.isPaused {
		p.isPaused = true
		p.control.Paused = true
	}
}

func (p *Player) Next() {
	if (!p.queueIsComplete) {
		p.songDone <- true
	} else {
		p.queueIndex++
	}
}

func (p *Player) Previous() {
	p.queueIndex--
	if (!p.queueIsComplete) {
		p.songDone <- false
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
			var decodeErr error
			p.streamer, p.format, decodeErr = decodeAudioFile(p.queue[p.queueIndex])
			if decodeErr != nil {
				fmt.Println(decodeErr)
			} else {
				if (!p.isSpeakerInited) {
					p.isSpeakerInited = true
					speaker.Init(p.format.SampleRate, p.format.SampleRate.N(p.Config.BufferedTime))
				}
				p.control = &beep.Ctrl {
					Streamer: p.streamer,
					Paused: p.isPaused,
				}
				speaker.Clear()
				speaker.Play(beep.Seq(p.control, beep.Callback(func() { p.songDone <- true })))
			}
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
	//fmt.Println("File decoded as " + mime)
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
	return ""
}

type PlayerConfig struct {
	BufferedTime time.Duration
}

type ReadSeekerCloser interface {
	io.Reader
	io.Seeker
	io.Closer
}
