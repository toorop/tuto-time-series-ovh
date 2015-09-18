package main

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"strings"
	"time"
)

/*
0 - major number
1 - minor mumber
2 - device name
3 - reads completed successfully
4 - reads merged
5 - sectors read
6 - time spent reading (ms)
7 - writes completed
8 - writes merged
9 - sectors written
10 - time spent writing (ms)
11 - I/Os currently in progress
12 - time spent doing I/Os (ms)
13 - weighted time spent doing I/Os (ms)
*/

// Diskstats represents a /proc/diskstats line
type DiskIO struct {
	Reads     uint64
	Writes    uint64
	Timestamp int64
}

// GetDisksIO returns nb reads and nd writes
func GetDisksIO() (stats map[string]DiskIO, err error) {
	stats = make(map[string]DiskIO)
	now := time.Now().Unix()
	ioRaw, err := ioutil.ReadFile("/proc/diskstats")
	if err != nil {
		return
	}
	scanner := bufio.NewScanner(bytes.NewReader(ioRaw))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if strings.HasPrefix(parts[2], "ram") || strings.HasPrefix(parts[2], "loop") || strings.HasPrefix(parts[2], "sr") {
			continue
		}
		stats[parts[2]] = DiskIO{
			Reads:     parseUint64(parts[4]),
			Writes:    parseUint64(parts[8]),
			Timestamp: now,
		}
	}
	return
}
