package main

import (
	"os"
)

func IsSymlink(path string) (ret bool, err error) {
	fi, err := os.Lstat(path)
	if err != nil {
		return
	}

	ret = fi.Mode()&os.ModeSymlink == os.ModeSymlink
	return
}
