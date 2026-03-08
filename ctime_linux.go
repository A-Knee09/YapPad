//go:build linux

package main

import (
	"syscall"
	"time"
)

func getCreationTime(stat *syscall.Stat_t) time.Time {
	return time.Unix(stat.Ctim.Sec, stat.Ctim.Nsec)
}
