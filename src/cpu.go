package main

import (
	"bufio"
	"bytes"
	"errors"
	"io/ioutil"
	"strings"
)

// CPUStats represente les stats CPU
type CPUStats struct {
	User    uint64
	Nice    uint64
	Sys     uint64
	Idle    uint64
	Wait    uint64
	Irq     uint64
	SoftIrq uint64
	Stolen  uint64
}

// Sum adds all "stats" (cpu time)
func (s *CPUStats) Sum() uint64 {
	return s.Idle + s.Irq + s.Nice + s.SoftIrq + s.Stolen + s.Sys + s.User + s.Wait
}

// GetCPUStats retourne les stats CPU
func GetCPUStats() (*map[string]CPUStats, error) {
	procStats, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(bytes.NewReader(procStats))
	cStats := make(map[string]CPUStats)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "cpu") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 8 {
			return nil, errors.New("bad data found in /proc/stat - " + line)
		}
		cStats[parts[0]] = CPUStats{
			User:    parseUint64(parts[1]),
			Nice:    parseUint64(parts[2]),
			Sys:     parseUint64(parts[3]),
			Idle:    parseUint64(parts[4]),
			Wait:    parseUint64(parts[5]),
			Irq:     parseUint64(parts[6]),
			SoftIrq: parseUint64(parts[7]),
			Stolen:  parseUint64(parts[8]),
		}
	}
	return &cStats, nil
}
