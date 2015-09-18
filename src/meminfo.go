package main

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"strings"
)

// GetMemStats parse /proc/meminfo and return stats
func GetMemStats() (map[string]uint64, error) {
	memStatsRaw, err := ioutil.ReadFile("/proc/meminfo")
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(bytes.NewReader(memStatsRaw))
	memStats := make(map[string]uint64)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		memStats[parts[0][:len(parts[0])-1]] = parseUint64(parts[1])
	}
	return memStats, nil
}
