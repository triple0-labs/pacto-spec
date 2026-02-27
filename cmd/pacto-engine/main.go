package main

import (
	"os"

	"pacto/internal/app"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	return app.Run(args)
}
