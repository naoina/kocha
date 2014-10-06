package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func copyAll(srcPath, destPath string) error {
	return filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		dest := filepath.Join(destPath, strings.TrimPrefix(path, srcPath))
		if info.IsDir() {
			err := os.MkdirAll(filepath.Join(dest), 0755)
			return err
		}
		src, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		return ioutil.WriteFile(dest, src, 0644)
	})
}
