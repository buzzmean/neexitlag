//go:build linux

package main

import (
	"os"
	"syscall"
)

func rootPrivateCA(path string) bool {
	f, e := os.Lstat(path)
	if e != nil || !f.Mode().IsRegular() || f.Mode().Perm() != 0600 {
		return false
	}
	s, ok := f.Sys().(*syscall.Stat_t)
	return ok && s.Uid == 0 && s.Gid == 0
}
