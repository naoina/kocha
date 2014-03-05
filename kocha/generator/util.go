package generator

import (
	"path/filepath"
	"runtime"
	"time"
)

var Now = time.Now // for test.

// SkeletonDir returns the directory of skeletons.
func SkeletonDir(name string) string {
	_, filename, _, _ := runtime.Caller(0)
	baseDir := filepath.Dir(filename)
	return filepath.Join(baseDir, "skeleton", name)
}
