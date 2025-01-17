package utils

import (
	"path/filepath"
	"strings"
)

func Slashify(p string) string {
	prefix := filepath.VolumeName(p)
	slashied := filepath.ToSlash(p)
	return strings.TrimPrefix(slashied, prefix)
}
