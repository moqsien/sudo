package utils

import (
	"os"

	"github.com/gogf/gf/os/gfile"
)

func CheckSocketFile(sockPath string) {
	_, err := os.Stat(sockPath)
	if !os.IsNotExist(err) {
		_ = gfile.Remove(sockPath)
	}
}

func FormatSocketPath(sockFileName string) string {
	return gfile.TempDir(sockFileName)
}
