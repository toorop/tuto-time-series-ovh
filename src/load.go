package main

import (
	"errors"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
)

// Loadavg représente le load à un instant 'timestamp'
type Loadavg struct {
	timestamp int64
	current   float64
	avg5      float64
	avg15     float64
}

// GetLoadAvg retourne la mesure de la charge
func GetLoadAvg() (*Loadavg, error) {
	loadAvgLine, err := ioutil.ReadFile("/proc/loadavg")
	if err != nil {
		return nil, err
	}
	parts := strings.Fields(string(loadAvgLine))
	if len(parts) < 3 {
		return nil, errors.New("bad format" + string(loadAvgLine))
	}

	load := &Loadavg{
		timestamp: time.Now().Unix(),
	}

	if load.current, err = strconv.ParseFloat(parts[0], 32); err != nil {
		return nil, err
	}

	if load.avg5, err = strconv.ParseFloat(parts[1], 32); err != nil {
		return nil, err
	}

	if load.avg15, err = strconv.ParseFloat(parts[2], 32); err != nil {
		return nil, err
	}

	return load, nil
}
