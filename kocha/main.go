package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"text/template"
)

type Error struct {
	usager  usager
	message string
}

func (e Error) Error() string {
	return e.message
}

type usager interface {
	Usage() string
}

type command interface {
	Name() string
	Alias() string
	Short() string
	Usage() string
	DefineFlags(*flag.FlagSet)
	Run()
}

var commands = []command{
	&newCommand{},
	&generateCommand{},
}

func usage() {
	var buf bytes.Buffer
	template.Must(template.New("usage").Parse(
		`usage: %s command [arguments]

Commands:
{{range .}}
    {{.Name|printf "%-12s"}} {{.Short}}{{if .Alias}} (alias: "{{.Alias}}"){{end}}{{end}}

`)).Execute(&buf, commands)
	fmt.Fprintf(os.Stderr, buf.String(), os.Args[0])
	os.Exit(1)
}

func main() {
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() == 0 {
		usage()
	}
	progName := os.Args[0]
	defer func() {
		if err := recover(); err != nil {
			if err, ok := err.(Error); ok {
				fmt.Fprintln(os.Stderr, err.message)
				fmt.Fprintf(os.Stderr, "usage: %s %s\n", progName, err.usager.Usage())
				os.Exit(1)
			}
			panic(err)
		}
	}()
	cmdName := flag.Arg(0)
	for _, cmd := range commands {
		switch cmdName {
		case cmd.Name(), cmd.Alias():
			flagSet := flag.NewFlagSet(cmd.Name(), flag.ExitOnError)
			flagSet.Usage = func() {
				fmt.Fprintf(os.Stderr, "usage: %s %s\n", progName, cmd.Usage())
			}
			cmd.DefineFlags(flagSet)
			flagSet.Parse(flag.Args()[1:])
			cmd.Run()
			return
		}
	}
	fmt.Fprintf(os.Stderr, "unknown command: %v\n", cmdName)
	usage()
}
