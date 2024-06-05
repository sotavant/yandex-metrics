package main

import (
	"fmt"
	"os"
)

func main() {
	os.Exit(0) // want "bad expression for use"
	os.Exit(1) // want "bad expression for use"
}

func usage() {
	fmt.Fprintln(os.Stderr, "Usage: linter testdata")
}
