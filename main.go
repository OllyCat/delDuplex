package main

import (
	"fmt"
	"os"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		usage(os.Args[0])
		os.Exit(1)
	}
}

func usage(prog string) {
	fmt.Printf("Using:\n\n%s <dir name> [<dir name> ...]\n", prog)
}
