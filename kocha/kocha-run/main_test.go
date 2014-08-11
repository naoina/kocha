package main

import "testing"

func TestRun(t *testing.T) {
	// The below tests do not end because run() have an infinite loop.
	// Any ideas?

	// func() {
	// tempDir, err := ioutil.TempDir("", "Test_runCommand_Run")
	// if err != nil {
	// t.Fatal(err)
	// }
	// defer os.RemoveAll(tempDir)
	// if err := os.Chdir(tempDir); err != nil {
	// t.Fatal(err)
	// }
	// if err := ioutil.WriteFile(filepath.Join(tempDir, "dev.go"), []byte(`
	// package main
	// func main() { panic("expected panic") }
	// `), 0644); err != nil {
	// t.Fatal(err)
	// }
	// cmd := &runCommand{}
	// flags := flag.NewFlagSet("testflags", flag.ExitOnError)
	// cmd.DefineFlags(flags)
	// flags.Parse([]string{})
	// defer func() {
	// if err := recover(); err == nil {
	// t.Error("Expect panic, but not occurred")
	// }
	// }()
	// cmd.Run()
	// }()

	// func() {
	// tempDir, err := ioutil.TempDir("", "Test_runCommand_Run")
	// if err != nil {
	// t.Fatal(err)
	// }
	// defer os.RemoveAll(tempDir)
	// if err := os.Chdir(tempDir); err != nil {
	// t.Fatal(err)
	// }
	// if err := ioutil.WriteFile(filepath.Join(tempDir, "dev.go"), []byte(`
	// package main
	// func main() {}
	// `), 0644); err != nil {
	// t.Fatal(err)
	// }
	// cmd := &runCommand{}
	// flags := flag.NewFlagSet("testflags", flag.ExitOnError)
	// cmd.DefineFlags(flags)
	// flags.Parse([]string{})
	// cmd.Run()
	// binName := filepath.Base(tempDir)
	// if _, err := os.Stat(filepath.Join(tempDir, binName)); err != nil {
	// t.Error("Expect %v is exists, but not", binName)
	// }
	// }()
}
