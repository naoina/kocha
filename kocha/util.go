package main

import (
	"fmt"
	"strconv"

	"github.com/naoina/kocha"
)

func printSettingEnv() {
	env, err := kocha.FindSettingEnv()
	if err != nil {
		panic(err)
	}
	fmt.Println("NOTE: You can setting your app by using following environment variables when launching an app:\n")
	for key, value := range env {
		fmt.Printf("%4s%v=%v\n", "", key, strconv.Quote(value))
	}
	fmt.Println()
}
