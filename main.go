package main

import (
	"log"
	"os"

	"flint-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
