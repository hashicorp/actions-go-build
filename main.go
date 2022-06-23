package main

import (
	"log"

	"github.com/hashicorp/actions-go-build/internal/build"
)

func main() {
	b := build.New()
	if err := b.Run(); err != nil {
		log.Fatal(err)
	}
}
