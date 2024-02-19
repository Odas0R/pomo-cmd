package main

import (
	"log"
	"os"

	pomo "github.com/odas0r/pomo-cmd"
)

func main() {
	if err := pomo.App.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
