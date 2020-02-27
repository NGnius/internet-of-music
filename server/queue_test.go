// Created by NGnius 2020-02-05

package main

import (
	"io/ioutil"
	"os"
	"strconv"
	"testing"
	"time"
)

const (
	dummyFileExt   = ".txt"
	dummyFileStart = "deleteMe"
)

var (
	test_qc = QueueConfig{
		PersistToDisk:   true,
		MemBufferSize:   2, // buffer length of 5 (2 previous, current, 2 next)
		EnableOvercache: true,
		OvercacheSize:   2, // overflow cache length of 10
		LoadTimeout:     1 * time.Second,
	}
)

func TestFullQueue(t *testing.T) {
	q := NewRollingQueue(test_qc)
	defer cleanupDummyFiles()
	defer q.Close()
	filenames := generateDummyFiles(10)
	for _, filename := range filenames {
		q.AppendFile(filename)
	}
	t.Log("--- q.Next() calls")
	count := 0
	for q.HasNext() {
		f, err := q.Next()
		if err != nil {
			t.Fatalf("q.Next() raised error %s (count = %d)", err, count)
		} else {
			data, readErr := ioutil.ReadAll(f)
			if readErr != nil {
				t.Errorf("ioutil.ReadAll(q.Next()'s file) raised error %s (count = %d)", err, count)
			}
			t.Logf("File %d contents: %s", count, string(data))
		}
		count++
	}
	if count != len(filenames) {
		t.Fatalf("Expected next count of %d, got %d", len(filenames), count)
	}
	t.Log("--- q.Now() call")
	f, err := q.Now()
	if err != nil {
		t.Fatalf("q.Now() raised error %s (count = %d)", err, count)
	} else {
		f.Seek(0, 0)
		data, readErr := ioutil.ReadAll(f)
		if readErr != nil {
			t.Errorf("ioutil.ReadAll(q.Next()'s file) raised error %s (count = %d)", err, count)
		}
		t.Logf("File now contents: %s", string(data))
	}
	t.Log("--- q.Previous() calls")
	count = 0
	for q.HasPrevious() {
		f, err := q.Previous()
		if err != nil {
			t.Fatalf("q.Previous() raised error %s (count = %d)", err, count)
		} else {
			f.Seek(0, 0)
			data, readErr := ioutil.ReadAll(f)
			if readErr != nil {
				t.Errorf("ioutil.ReadAll(q.Previous()'s file) raised error %s (count = %d)", err, count)
			}
			t.Logf("File %d contents: %s", count, string(data))
		}
		count++
	}
	if count != len(filenames)-1 {
		t.Fatalf("Expected previous count of %d, got %d", len(filenames)-1, count)
	}
}

func generateDummyFiles(count int) []string {
	res := []string{}
	for i := 0; i < count; i++ {
		filename := dummyFileStart + strconv.Itoa(i) + dummyFileExt
		f, err := os.Create(filename)
		if err == nil {
			defer f.Close()
			f.WriteString(filename)
			f.Seek(0, 0)
		}
		res = append(res, filename)
	}
	return res
}

func cleanupDummyFiles() {
	for i := 0; ; i++ {
		filename := dummyFileStart + strconv.Itoa(i) + dummyFileExt
		err := os.Remove(filename)
		if err != nil {
			return
		}
	}
}
