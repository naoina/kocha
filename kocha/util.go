package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
)

type fileStatus uint8

const (
	fileStatusConflict fileStatus = iota
	fileStatusNoConflict
	fileStatusIdentical
)

func panicOnError(usager usager, format string, a ...interface{}) {
	panic(Error{usager, fmt.Sprintf(format, a...)})
}

func copyTemplate(u usager, srcPath, dstPath string, data map[string]interface{}) {
	tmpl, err := template.ParseFiles(srcPath)
	if err != nil {
		panicOnError(u, "abort: failed to parse template: %v", err)
	}
	var bufFrom bytes.Buffer
	if err := tmpl.Execute(&bufFrom, data); err != nil {
		panicOnError(u, "abort: failed to process template: %v", err)
	}
	printFunc := printCreate
	switch detectConflict(u, bufFrom.Bytes(), dstPath) {
	case fileStatusConflict:
		printConflict(dstPath)
		if !confirmOverwrite(dstPath) {
			printSkip(dstPath)
			return
		}
		printFunc = printOverwrite
	case fileStatusIdentical:
		printIdentical(dstPath)
		return
	}
	dstFile, err := os.Create(dstPath)
	if err != nil {
		panicOnError(u, "abort: failed to create file: %v", err)
	}
	defer dstFile.Close()
	if _, err := io.Copy(dstFile, &bufFrom); err != nil {
		panicOnError(u, "abort: failed to output file: %v", err)
	}
	printFunc(dstPath)
}

func detectConflict(u usager, src []byte, dstPath string) fileStatus {
	if _, err := os.Stat(dstPath); os.IsNotExist(err) {
		return fileStatusNoConflict
	}
	dstBuf, err := ioutil.ReadFile(dstPath)
	if err != nil {
		panicOnError(u, "abort: failed to read file: %v", err)
	}
	if bytes.Equal(src, dstBuf) {
		return fileStatusIdentical
	}
	return fileStatusConflict
}

func confirmOverwrite(dstPath string) bool {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("Overwrite %v? [Yn] ", dstPath)
		yesno, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		switch strings.ToUpper(strings.TrimSpace(yesno)) {
		case "", "YES", "Y":
			return true
		case "NO", "N":
			return false
		}
	}
}

func red(s string) string {
	return fmt.Sprintf("\x1b[31;1m%20s\x1b[0m", s)
}

func green(s string) string {
	return fmt.Sprintf("\x1b[32;1m%20s\x1b[0m", s)
}

func yellow(s string) string {
	return fmt.Sprintf("\x1b[33;1m%20s\x1b[0m", s)
}

func blue(s string) string {
	return fmt.Sprintf("\x1b[34;1m%20s\x1b[0m", s)
}

func magenta(s string) string {
	return fmt.Sprintf("\x1b[35;1m%20s\x1b[0m", s)
}

func cyan(s string) string {
	return fmt.Sprintf("\x1b[36;1m%20s\x1b[0m", s)
}

func printIdentical(path string) {
	fmt.Println(blue("identical"), "", path)
}

func printConflict(path string) {
	fmt.Println(red("conflict"), "", path)
}

func printSkip(path string) {
	fmt.Println(cyan("skip"), "", path)
}

func printOverwrite(path string) {
	fmt.Println(cyan("overwrite"), "", path)
}

func printCreate(path string) {
	fmt.Println(green("create"), "", path)
}
