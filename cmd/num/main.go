package main

import (
	"fmt"
	"os"

	"git.vieth.io/num"
)

func main() {
	if err := num.NewEncoder(os.Stdout).Encode(os.Stdin); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}
