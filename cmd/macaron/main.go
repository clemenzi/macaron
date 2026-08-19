package main

import (
	"os"

	"github.com/clemenzi/macaron/internal/macaron"
)

func main() {
	os.Exit(macaron.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
