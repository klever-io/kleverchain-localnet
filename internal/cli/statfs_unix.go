//go:build !windows

package cli

import "syscall"

func freeDiskBytes(path string) (uint64, error) {
	var s syscall.Statfs_t
	if err := syscall.Statfs(path, &s); err != nil {
		return 0, err
	}
	bsize := s.Bsize
	if bsize < 0 {
		bsize = 0
	}
	return s.Bavail * uint64(bsize), nil
}
