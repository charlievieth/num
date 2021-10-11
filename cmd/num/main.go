package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charlievieth/num"
	"github.com/spf13/pflag"
)

var (
	InputFile  string
	OutputFile string
)

func init() {
	pflag.Usage = Usage
	pflag.StringVarP(&InputFile, "input", "i", "",
		"read input from FILE instead of standard input")
	pflag.StringVarP(&OutputFile, "output", "o", "",
		"write result to FILE instead of standard output")
}

func Usage() {
	const message = "Usage: %s [OPTION]... [TEXT]...\n" +
		"Add thousands separators to TEXT and write the result to standard output.\n" +
		"\n" +
		"With no TEXT or FILE, or when FILE is -, read standard input.\n\n"
	fmt.Fprintf(os.Stdout, message, filepath.Base(os.Args[0]))
	pflag.PrintDefaults()

	const example = "\nEXAMPLES:\n" +
		"\n" +
		"  $ echo '123456' | %[1]s\n" +
		"\n" +
		"  Will print '123,456' on standard output.\n" +
		"\n" +
		"  $ %[1]s '123456'\n" +
		"\n" +
		"  Will print '123,456' on standard output.\n" +
		"\n" +
		"  $ %[1]s -i FILE\n" +
		"\n" +
		"  Will read from FILE, add thousands separators\n" +
		"  and print the result on standard output.\n"
	fmt.Fprintf(os.Stdout, example, filepath.Base(os.Args[0]))
}

func formatText(out *os.File, args []string) error {
	var buf bytes.Buffer
	for _, s := range args {
		buf.Reset()
		r := strings.NewReader(s)
		if err := num.NewEncoder(&buf).Encode(r); err != nil {
			return err
		}
		buf.WriteByte('\n')
		if _, err := buf.WriteTo(out); err != nil {
			return err
		}
	}
	return nil
}

func realMain() error {
	out := os.Stdout
	if OutputFile != "" && OutputFile != "-" {
		f, err := os.OpenFile(OutputFile, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer f.Close()
		out = f
	}

	in := os.Stdin
	if InputFile != "" && InputFile != "-" {
		f, err := os.Open(InputFile)
		if err != nil {
			return err
		}
		defer f.Close()
		in = f
	}

	if pflag.NArg() != 0 {
		return formatText(out, pflag.Args())
	}

	// stream
	return num.NewEncoder(out).Encode(in)
}

func main() {
	pflag.Parse()

	if pflag.NArg() != 0 && InputFile != "" {
		fmt.Fprintln(os.Stderr, "error: TEXT and '--input' cannot both be specified")
		pflag.Usage()
		os.Exit(1)
	}

	if err := realMain(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
