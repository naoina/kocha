package generator

import (
	"path/filepath"
	"runtime"
)

// SkeletonDir returns the directory of skeletons.
func SkeletonDir(name string) string {
	_, filename, _, _ := runtime.Caller(0)
	baseDir := filepath.Dir(filename)
	return filepath.Join(baseDir, "skeleton", name)
}
