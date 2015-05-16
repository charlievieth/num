package main

import (
	"bufio"
	"fmt"
	"io"
	"num"
	"os"
)

func main() {
	n := num.New()
	r := bufio.NewReader(os.Stdin)
	b := make([]byte, 4096)
	var (
		err error
		i   int
	)
	for err == nil {
		i, err = r.Read(b)
		if err != nil && err != io.EOF {
			break
		}
		if _, err = n.Write(b[:i]); err != nil {
			break
		}
		if _, err = n.WriteTo(os.Stdout); err != nil {
			break
		}
	}
	if err != nil && err != io.EOF {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}
