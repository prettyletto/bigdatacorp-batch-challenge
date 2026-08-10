package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
)

type Stats struct {
	RecordsRead int64
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: batch <input.jsonl>")
		os.Exit(1)
	}

	inputPath := os.Args[1]

	file, err := os.Open(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open input: %v\n", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	var stats Stats

	for {
		line, err := reader.ReadBytes('\n')

		if len(line) > 0 {
			stats.RecordsRead++
		}

		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read input: %v\n", err)
			os.Exit(1)
		}
	}

	printStats(stats)
}

func printStats(stats Stats) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "Batch Results")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Count")
	fmt.Fprintf(w, "Records Read\t%d\n", stats.RecordsRead)
}
