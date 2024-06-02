package main

import (
	"fmt"
	"os"
)

func main() {
	os.Exit(0)
}

func usage() {
	fmt.Fprintln(os.Stderr, "Usage: linter testdata")
}
