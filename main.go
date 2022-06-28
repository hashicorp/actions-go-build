package main

import (
	"log"
	"os"

	"github.com/hashicorp/actions-go-build/internal/commands"
)

func main() {
	if err := commands.Main().Execute(os.Args); err != nil {
		log.Fatal(err)
	}
}
