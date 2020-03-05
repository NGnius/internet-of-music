// Created by NGnius 2020-02-05

package main

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"time"
)

const (
	FilenameStart = "persisted"
	FilenameEnd   = ".file"
)

// QueueConfig configuration information for RollingQueue
type QueueConfig struct {
	PersistToDisk   bool
	FilenameStart   string
	FilenameEnd     string
	MemBufferSize   int
	EnableOvercache bool
	OvercacheSize   int
	LoadTimeout     time.Duration
}

// RollingQueue file queue where items roll off the end
type RollingQueue struct {
	currentIndex    int
	maximumIndex    int                // maximum enqueued item
	minimumIndex    int                // minimum equeued item
	memBuffer       []ReadSeekerCloser // middle item is the current index
	overflowBuffer  []ReadSeekerCloser // overflow cache for upcoming files
	overflowIndexes []int              // overflow cache files' absolute queue index
	config          QueueConfig
	loadSyncChan    chan bool
}

func NewRollingQueue(qc QueueConfig) (rq RollingQueue) {
	rq.loadSyncChan = make(chan bool)
	rq.config = qc
	// config integrity checks
	// rq.config.MemBufferSize must be >= 1
	if rq.config.MemBufferSize < 1 {
		rq.config.MemBufferSize = 1
	}
	// rq.config.FilenameStart cannot be empty
	if rq.config.FilenameStart == "" {
		rq.config.FilenameStart = FilenameStart
	}
	// rq.config.FilenameEnd cannot be empty
	if rq.config.FilenameEnd == "" {
		rq.config.FilenameEnd = FilenameEnd
	}
	// init memBuffer
	rq.memBuffer = []ReadSeekerCloser{}
	for i := 0; i < (rq.config.MemBufferSize*2)+1; i++ {
		rq.memBuffer = append(rq.memBuffer, nil)
	}
	// init overflowBuffer
	rq.overflowBuffer = []ReadSeekerCloser{}
	rq.overflowIndexes = []int{}
	if qc.EnableOvercache {
		for i := 0; i < rq.config.OvercacheSize; i++ {
			rq.overflowBuffer = append(rq.overflowBuffer, nil)
			rq.overflowIndexes = append(rq.overflowIndexes, -1)
		}
	}
	rq.currentIndex = -1
	go rq.loadComplete(true)
	return
}

// utility
// indexInBuffer translate an absolute index to a memory buffer location
func (rq *RollingQueue) indexInBuffer(index int) int {
	// when currentIndex is 42 and MemBufferSize is 3
	// [39 40 41 42 43 44 45]
	// index 39 is actually memBuffer[0]// TODO load from disk
	return index + rq.config.MemBufferSize - rq.currentIndex
}

// existsInBugger determine if the absolute index exists in the memory buffer
func (rq *RollingQueue) existsInBuffer(index int) bool {
	return (index < rq.currentIndex+rq.config.MemBufferSize+1) && (index > rq.currentIndex-rq.config.MemBufferSize-1)
}

func (rq *RollingQueue) highestBufferIndex() int {
	return rq.config.MemBufferSize * 2
}

func (rq *RollingQueue) middleBufferIndex() int {
	return rq.config.MemBufferSize
}

func (rq *RollingQueue) generateFilename(index int) string {
	return rq.config.FilenameStart + strconv.Itoa(index) + rq.config.FilenameEnd
}

// file load sync
func (rq *RollingQueue) loadComplete(success bool) {
	rq.loadSyncChan <- success
}

func (rq *RollingQueue) waitForLoadComplete() bool {
	return <-rq.loadSyncChan
}

// persistence
func (rq *RollingQueue) persistIfConfig(index int, file ReadSeekerCloser) error {
	if rq.config.PersistToDisk {
		return rq.persist(index, rq.memBuffer[0])
	}
	return nil
}

func (rq *RollingQueue) persist(index int, file ReadSeekerCloser) (err error) {
	//fmt.Printf("Persisting %d\n", index)
	var data []byte
	var diskFile *os.File
	file.Seek(0, 0)
	data, err = ioutil.ReadAll(file)
	if err != nil {
		return
	}
	file.Close()
	diskFile, err = os.Create(rq.generateFilename(index))
	if err != nil {
		return
	}
	defer diskFile.Close()
	_, err = diskFile.Write(data)
	return
}

func (rq *RollingQueue) loadIndexInto(index, bufferIndex int) {
	if rq.config.EnableOvercache {
		overflowIndex := rq.overflowIndexOfIndex(index)
		if overflowIndex != -1 {
			rq.memBuffer[bufferIndex] = rq.overflowBuffer[overflowIndex]
			rq.overflowBuffer[overflowIndex] = nil
			rq.overflowIndexes[overflowIndex] = -1
			go rq.loadComplete(true)
			return
		}
	}
	filename := rq.generateFilename(index)
	diskFile, openErr := os.Open(filename)
	if openErr != nil {
		go rq.loadComplete(false)
		return
	}
	newRSC, copyErr := copyReader(diskFile)
	diskFile.Close()
	if copyErr != nil {
		go rq.loadComplete(false)
		return
	}
	rq.memBuffer[bufferIndex] = newRSC
	go rq.loadComplete(true)
	os.Remove(filename) // remove file once it's successfully loaded
}

// caching
func (rq *RollingQueue) shiftLeft() {
	// [ 0 1 2 3 4 ] -> [ 1 2 3 4 5 ]
	for i := 1; i <= rq.highestBufferIndex(); i++ {
		rq.memBuffer[i-1] = rq.memBuffer[i]
	}
	rq.memBuffer[rq.highestBufferIndex()] = nil
}

func (rq *RollingQueue) shiftRight() {
	for i := rq.highestBufferIndex(); i > 0; i-- {
		rq.memBuffer[i] = rq.memBuffer[i-1]
	}
	rq.memBuffer[0] = nil
}

// overflow caching
func (rq *RollingQueue) emptyOverflowIndex() int {
	for i := 0; i < rq.config.OvercacheSize; i++ {
		if rq.overflowIndexes[i] == -1 {
			return i
		}
	}
	return -1
}

func (rq *RollingQueue) existsSpaceInOverflow() bool {
	return rq.emptyOverflowIndex() != -1
}

func (rq *RollingQueue) overflowIndexOfIndex(index int) int {
	for i := 0; i < rq.config.OvercacheSize; i++ {
		if rq.overflowIndexes[i] == index {
			return i
		}
	}
	return -1
}

func (rq *RollingQueue) existsIndexInOverflow(index int) bool {
	return rq.overflowIndexOfIndex(index) != -1
}

func (rq *RollingQueue) cacheInOverflow(file ReadSeekerCloser, index, overflowIndex int) {
	rq.overflowBuffer[overflowIndex] = file
	rq.overflowIndexes[overflowIndex] = index
}

// TryCacheInOverflow try to add file to overflow. Returns true on success, false otherwise (ie when buffer is full)
func (rq *RollingQueue) TryCacheInOverflow(index int, file ReadSeekerCloser) bool {
	overflowIndex := rq.emptyOverflowIndex()
	if overflowIndex != -1 {
		rq.cacheInOverflow(file, index, overflowIndex)
		return true
	}
	return false
}

// Next move the queue to the next item and returns that item
// returns nil if the next item does not exist
func (rq *RollingQueue) Next() (ReadSeekerCloser, error) {
	if !rq.HasNext() {
		return nil, errors.New("NoNextItem")
	}
	if !rq.waitForLoadComplete() {
		go rq.loadComplete(false)
		return nil, errors.New("LoadFailure")
	}
	// persist previous item that will roll off the memBuffer
	if (rq.currentIndex - rq.config.MemBufferSize) >= 0 {
		err := rq.persistIfConfig(rq.currentIndex-rq.config.MemBufferSize, rq.memBuffer[0])
		if err != nil {
			return nil, err
		}
		if !rq.config.PersistToDisk {
			rq.minimumIndex++
		}
		//fmt.Printf("Minimum index is now %d\n", rq.minimumIndex)
	}
	rq.shiftLeft()
	rq.currentIndex++
	start := time.Now()
	if (rq.currentIndex + rq.config.MemBufferSize) < rq.maximumIndex {
		go rq.loadIndexInto(rq.currentIndex+rq.config.MemBufferSize, rq.highestBufferIndex())
	} else {
		go rq.loadComplete(true)
	}
	//loadWaitLoop:
	for {
		if rq.memBuffer[rq.middleBufferIndex()] != nil {
			return rq.memBuffer[rq.middleBufferIndex()], nil
		} else if time.Since(start) >= rq.config.LoadTimeout {
			return nil, errors.New("LoadTimeout")
		}
	}
}

func (rq *RollingQueue) HasNext() bool {
	return rq.maximumIndex > rq.currentIndex+1
}

func (rq *RollingQueue) Previous() (ReadSeekerCloser, error) {
	if !rq.HasPrevious() {
		return nil, errors.New("NoPreviousItem")
	}
	if !rq.waitForLoadComplete() {
		go rq.loadComplete(false)
		return nil, errors.New("LoadFailure")
	}
	var err error
	if rq.currentIndex+rq.config.MemBufferSize < rq.maximumIndex {
		// persist upcoming item that will roll off the memBuffer
		overflowSuccess := rq.TryCacheInOverflow(rq.config.MemBufferSize+rq.currentIndex, rq.memBuffer[rq.highestBufferIndex()])
		if !overflowSuccess {
			err = rq.persist(rq.config.MemBufferSize+rq.currentIndex, rq.memBuffer[rq.highestBufferIndex()])
		}
		if err != nil {
			return nil, err
		}
	}
	rq.shiftRight()
	rq.currentIndex--
	start := time.Now()
	if (rq.currentIndex-rq.config.MemBufferSize) >= 0 && rq.config.PersistToDisk {
		go rq.loadIndexInto(rq.currentIndex-rq.config.MemBufferSize, 0)
	} else {
		go rq.loadComplete(true)
	}
	//loadWaitLoop:
	for {
		if rq.memBuffer[rq.middleBufferIndex()] != nil {
			return rq.memBuffer[rq.middleBufferIndex()], nil
		} else if time.Since(start) >= rq.config.LoadTimeout {
			return nil, errors.New("LoadTimeout")
		}
	}
}

func (rq *RollingQueue) HasPrevious() bool {
	return (rq.existsInBuffer(rq.currentIndex-1) || rq.config.PersistToDisk) && rq.currentIndex > rq.minimumIndex
}

// Now get the current queue item
func (rq *RollingQueue) Now() (ReadSeekerCloser, error) {
	if !rq.HasNow() {
		return nil, errors.New("NoCurrentItem")
	}
	return rq.memBuffer[rq.middleBufferIndex()], nil
}

func (rq *RollingQueue) HasNow() bool {
	return rq.currentIndex != -1
}

// Index get the absolute index of the current item
func (rq *RollingQueue) Index() int {
	return rq.currentIndex
}

func (rq *RollingQueue) Append(file ReadSeekerCloser) (err error) {
	// A file may be stored (by priority):
	// - in the memBuffer cache
	// -	 in the overflow cache
	// - on disk
	if rq.existsInBuffer(rq.maximumIndex) {
		rq.memBuffer[rq.indexInBuffer(rq.maximumIndex)] = file
		rq.maximumIndex++
	} else {
		// file must go in overflow cache or is persisted
		overflowSuccess := rq.TryCacheInOverflow(rq.maximumIndex, file)
		if overflowSuccess {
			rq.maximumIndex++
			return
		}
		// try to persist
		err = rq.persist(rq.maximumIndex, file)
		if err != nil {
			return
		}
		rq.maximumIndex++
	}
	return
}

func (rq *RollingQueue) AppendCopy(file ReadSeekerCloser) (err error) {
	var data []byte
	data, err = ioutil.ReadAll(file)
	if err != nil {
		return
	}
	newFilelike := NewWrapCloser(bytes.NewReader(data)) // manually implemented Close()
	err = rq.Append(newFilelike)
	return
}

func (rq *RollingQueue) AppendFile(filename string) (err error) {
	var diskFile *os.File
	diskFile, err = os.Open(filename)
	if err != nil {
		return
	}
	err = rq.Append(diskFile)
	return
}

// Close call Close on all containing files and cleanup persisted files
func (rq *RollingQueue) Close() (err error) {
	if !rq.waitForLoadComplete() {
		go rq.loadComplete(false)
		return errors.New("LoadFailure")
	}
	// memBuffer
	minBufferIndex := 0
	if rq.existsInBuffer(rq.minimumIndex) {
		minBufferIndex = rq.indexInBuffer(rq.minimumIndex)
	}
	maxBufferIndex := rq.highestBufferIndex()
	if rq.existsInBuffer(rq.maximumIndex - 1) {
		maxBufferIndex = rq.indexInBuffer(rq.maximumIndex - 1)
	}
	for i := minBufferIndex; i <= maxBufferIndex; i++ {
		err = rq.memBuffer[i].Close()
		if err != nil {
			return
		}
	}
	// overflow cache
	if rq.config.EnableOvercache {
		for i := 0; i < rq.config.OvercacheSize; i++ {
			if rq.overflowIndexes[i] != -1 {
				//fmt.Println(rq.overflowIndexes[i])
				err = rq.overflowBuffer[i].Close()
				rq.overflowBuffer[i] = nil
				rq.overflowIndexes[i] = -1
				if err != nil {
					return
				}
			}
		}
	}
	// persisted previous items
	if rq.config.PersistToDisk {
		for i := 0; i < rq.currentIndex-rq.config.MemBufferSize; i++ {
			os.Remove(rq.generateFilename(i))
		}
	}
	// persisted next items
	for i := rq.currentIndex + rq.config.MemBufferSize + 1; i < rq.maximumIndex; i++ {
		os.Remove(rq.generateFilename(i))
	}
	return
}

// WrapCloser fake Closer wrapper for ReadSeekers
type WrapCloser struct {
	file io.ReadSeeker
}

func NewWrapCloser(file io.ReadSeeker) *WrapCloser {
	return &WrapCloser{file: file}
}

func (b *WrapCloser) Close() error {
	return nil
}

func (b *WrapCloser) Seek(offset int64, whence int) (int64, error) {
	return b.file.Seek(offset, whence)
}

func (b *WrapCloser) Read(p []byte) (n int, err error) {
	n, err = b.file.Read(p)
	return
}

func copyReader(reader io.Reader) (ReadSeekerCloser, error) {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	res := NewWrapCloser(bytes.NewReader(data))
	return res, nil
}
