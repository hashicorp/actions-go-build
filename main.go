package main

import (
	"log"
	"os"

	"github.com/hashicorp/actions-go-build/pkg/commands2"
)

func main() {
	status, err := commands2.MakeCLI(os.Args[1:]).Run()
	if err != nil {
		log.Println(err)
	}
	os.Exit(status)
}
