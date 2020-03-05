// Created by NGnius 2020-01-05

package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/faiface/beep"
	"github.com/faiface/beep/flac"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/vorbis"
	"github.com/faiface/beep/wav"

	//"os"
	"time"
)

type Player struct {
	streamer        beep.Streamer
	format          beep.Format
	control         *beep.Ctrl
	queue           *RollingQueue
	Config          PlayerConfig
	songDone        chan bool
	isPaused        bool
	isSpeakerInited bool
	isHandling      bool
}

func NewPlayer() (p *Player) {
	qc := QueueConfig{
		PersistToDisk:   false,
		MemBufferSize:   2,
		EnableOvercache: true,
		OvercacheSize:   2,
	}
	rq := NewRollingQueue(qc)
	p = &Player{
		queue: &rq,
		Config: PlayerConfig{
			BufferedTime: Buffer,
			SampleRate:   SampleRate,
			Quality:      Quality,
		},
	}
	return
}

func (p *Player) Init() {
	p.songDone = make(chan bool)
	rq := NewRollingQueue(QueueConfig{
		PersistToDisk:   false,
		MemBufferSize:   2,
		EnableOvercache: true,
		OvercacheSize:   2,
	})
	p.queue = &rq
}

func (p *Player) Enqueue(audioFile ReadSeekerCloser) {
	p.queue.Append(audioFile)
}

func (p *Player) EnqueueMany(audioFiles ...ReadSeekerCloser) {
	for _, f := range audioFiles {
		p.queue.Append(f)
	}
}

func (p *Player) Play() {
	if p.isPaused {
		p.isPaused = false
		if p.control != nil {
			p.control.Paused = false
		}
	}
	if !p.isHandling && p.queue.HasNext() {
		go p.handleSongEnd()
		p.songDone <- true
	}
}

func (p *Player) Pause() {
	if !p.isPaused {
		p.isPaused = true
		if p.control != nil {
			p.control.Paused = true
		}
	}
}

func (p *Player) Next() {
	p.songDone <- true
}

func (p *Player) Previous() {
	if p.queue.HasPrevious() {
		p.queue.Previous()
		p.songDone <- false
	}
}

func (p *Player) handleSongEnd() {
	p.isHandling = true
	fmt.Println("Starting queue handler")
handlerLoop:
	for {
		advance := <-p.songDone
		if (advance && p.queue.HasNext()) || (!advance) {
			if advance {
				p.queue.Next()
			}
			var decodeErr error
			nowF, nowErr := p.queue.Now()
			if nowErr != nil {
				fmt.Println(nowErr)
			}
			p.streamer, p.format, decodeErr = decodeAudioFile(nowF)
			if decodeErr != nil {
				fmt.Println(decodeErr)
			} else {
				targetSR := beep.SampleRate(p.Config.SampleRate)
				p.streamer = beep.Resample(p.Config.Quality, p.format.SampleRate, targetSR, p.streamer)
				if !p.isSpeakerInited {
					p.isSpeakerInited = true
					speaker.Init(targetSR, targetSR.N(p.Config.BufferedTime))
				}
				p.control = &beep.Ctrl{
					Streamer: p.streamer,
					Paused:   p.isPaused,
				}
				speaker.Clear()
				speaker.Play(beep.Seq(p.control, beep.Callback(func() { p.songDone <- true })))
			}
		} else {
			speaker.Clear()
			p.isHandling = false
			fmt.Println("Queue finished, shutting down queue handler")
			break handlerLoop
		}
	}
}

func decodeAudioFile(f ReadSeekerCloser) (streamer beep.Streamer, format beep.Format, decodeErr error) {
	f.Seek(0, 0)
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
	SampleRate   int64
	Quality      int
}

type ReadSeekerCloser interface {
	io.Reader
	io.Seeker
	io.Closer
}
