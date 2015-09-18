package main

import "strconv"

// parse uint64 from string
// return 0 on error
func parseUint64(in string) uint64 {
	out, _ := strconv.ParseUint(in, 10, 64)
	return out
}
