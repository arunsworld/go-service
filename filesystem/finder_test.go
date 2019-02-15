package filesystem

import (
	"regexp"
	"testing"
)

func TestFindGoFiles(t *testing.T) {
	pattern := regexp.MustCompile(`.*\.go`)
	files := FindFilesInDirectory("test_filesystem", pattern)
	if len(files) != 3 {
		t.Errorf("Expected 3 files, got: %d", len(files))
	}
}
