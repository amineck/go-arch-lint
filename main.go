package main

import (
	"os"

	"github.com/amineck/go-arch-lint/internal/app"
)

func main() {
	os.Exit(run())
}

func run() int {
	return app.Execute()
}
