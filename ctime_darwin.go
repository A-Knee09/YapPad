//go:build darwin

package main

import (
	"syscall"
	"time"
)

func getCreationTime(stat *syscall.Stat_t) time.Time {
	return time.Unix(stat.Ctimespec.Sec, stat.Ctimespec.Nsec)
}
