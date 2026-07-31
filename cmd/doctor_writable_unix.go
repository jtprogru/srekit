//go:build unix

package cmd

import (
	"io/fs"
	"os"
	"syscall"
)

// dirWritable reports whether this process could create an entry in the
// directory info describes, judged from its mode bits and ownership rather
// than by attempting a write — doctor must not touch the filesystem. The
// second return value is false when the answer cannot be determined, in
// which case the caller reports the check as ok rather than guessing.
func dirWritable(info fs.FileInfo) (bool, bool) {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return false, false
	}
	perm := info.Mode().Perm()
	euid := int64(os.Geteuid())
	switch {
	case euid == 0:
		return true, true
	case int64(stat.Uid) == euid:
		return perm&0o200 != 0, true
	case inGroup(int64(stat.Gid)):
		return perm&0o020 != 0, true
	default:
		return perm&0o002 != 0, true
	}
}

func inGroup(gid int64) bool {
	if int64(os.Getegid()) == gid {
		return true
	}
	groups, err := os.Getgroups()
	if err != nil {
		return false
	}
	for _, g := range groups {
		if int64(g) == gid {
			return true
		}
	}
	return false
}
