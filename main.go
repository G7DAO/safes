package main

import (
	"fmt"
	"os"
)

var (
	rpcURL     string
	safeAPIURL string
)

func main() {
	command := CreateRootCommand()
	err := command.Execute()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
