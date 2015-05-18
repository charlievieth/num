package main

import (
	"fmt"
	"io"
	"num"
	"os"
)

const bufferSize = num.DefaultBufferSize

func main() {
	n := num.New()
	if err := fmtStream(n, os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}

func fmtStream(n *num.Num, r io.Reader, w io.Writer) error {
	var (
		err error
		i   int
	)
	b := make([]byte, bufferSize)
	for err == nil {
		i, err = r.Read(b)
		if err == nil || err == io.EOF {
			if _, err := n.Write(b[:i]); err != nil {
				return err
			}
			if _, err := n.WriteTo(w); err != nil {
				return err
			}
		}
	}
	if err != nil && err != io.EOF {
		return err
	}
	return nil
}
