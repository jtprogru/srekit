//go:build !unix

package cmd

import "io/fs"

// dirWritable has no portable non-unix implementation that avoids writing,
// so it declines to answer and the caller reports the check as ok. srekit
// ships linux/darwin/freebsd builds; this file exists so the package still
// compiles for anything else.
func dirWritable(_ fs.FileInfo) (bool, bool) {
	return false, false
}
