package generator

import (
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func TestSkeletonDir(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	baseDir := filepath.Join(filepath.Dir(filename), "skeleton")
	for _, v := range []struct {
		value    string
		expected string
	}{
		{"model", filepath.Join(baseDir, "model")},
		{"controller", filepath.Join(baseDir, "controller")},
		{"foo", filepath.Join(baseDir, "foo")},
		{"bar", filepath.Join(baseDir, "bar")},
	} {
		actual := SkeletonDir(v.value)
		expected := v.expected
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("SkeletonDir(%q) => %q, want %q", v.value, actual, expected)
		}
	}
}
