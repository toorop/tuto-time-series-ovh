package main

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"strings"
	"time"
)

// NetIO stocke les stats relative aux IO réseau
type NetIO struct {
	in        uint64
	out       uint64
	timestamp int64
}

// GetNetIO retourne les stats réseau
func GetNetIO() (*NetIO, error) {
	io := &NetIO{
		timestamp: time.Now().Unix(),
	}
	var parts []string
	ioRaw, err := ioutil.ReadFile("/proc/net/netstat")
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(bytes.NewReader(ioRaw))
	IPExtSeen := false
	for scanner.Scan() {
		line := scanner.Text()
		parts = strings.Fields(line)
		if parts[0] != "IpExt:" {
			continue
		}
		if !IPExtSeen {
			IPExtSeen = true
			continue
		}
		io.in = parseUint64(parts[7])
		io.out = parseUint64(parts[8])
		break
	}
	return io, nil
}
