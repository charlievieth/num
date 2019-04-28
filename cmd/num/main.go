package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charlievieth/num"
)

var (
	InputFile  string
	OutputFile string
)

func init() {
	flag.Usage = Usage

	flag.StringVar(&InputFile, "i", "", "read input from FILE instead of standard input")
	flag.StringVar(&InputFile, "-input", "", "read input from FILE instead of standard input")

	flag.StringVar(&OutputFile, "o", "", "write result to FILE instead of standard output")
	flag.StringVar(&OutputFile, "-output", "", "write result to FILE instead of standard output")
}

func Usage() {
	const message = "Usage: %s [OPTION]... [TEXT]...\n" +
		"Add thousands separators to TEXT and write the result to standard output.\n" +
		"\n" +
		"With no TEXT or FILE, or when FILE is -, read standard input.\n\n"
	fmt.Fprintf(flag.CommandLine.Output(), message, filepath.Base(os.Args[0]))
	flag.PrintDefaults()
	const example = "\nEXAMPLES:\n" +
		"\n" +
		"\techo '123456' | %[1]s\n" +
		"\n" +
		"\tWill print '123,456' on standard output.\n" +
		"\n" +
		"\t%[1]s '123456'\n" +
		"\n" +
		"\tWill print '123,456' on standard output.\n" +
		"\n" +
		"\t%[1]s -i FILE\n" +
		"\n" +
		"\tWill read from FILE, add thousands separators\n" +
		"\tand print the result on standard output.\n"
	fmt.Fprintf(flag.CommandLine.Output(), example, filepath.Base(os.Args[0]))
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

	if flag.NArg() != 0 {
		return formatText(out, flag.Args())
	}

	// stream
	return num.NewEncoder(out).Encode(in)
}

func main() {
	flag.Parse()

	if flag.NArg() != 0 && InputFile != "" {
		fmt.Fprintln(flag.CommandLine.Output(), "Error: TEXT and '--input' cannot both be specified")
		flag.Usage()
		os.Exit(1)
	}

	if err := realMain(); err != nil {
		fmt.Fprintln(flag.CommandLine.Output(), "Error:", err)
		os.Exit(1)
	}
}
