package filesystem

import (
	"os"
	"path/filepath"
	"regexp"
)

// FindFilesInDirectory finds files in a directory matching a pattern
func FindFilesInDirectory(dir string, pattern *regexp.Regexp) []string {
	files := []string{}
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if pattern.MatchString(path) {
			files = append(files, path)
		}
		return nil
	})
	return files
}
