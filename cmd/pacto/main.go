package main

import (
	"os"

	"pacto/internal/app"
)

func main() {
	os.Exit(run())
}

func run() int {
	return app.Execute()
}
