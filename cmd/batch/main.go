package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
)

type Stats struct {
	RecordsRead int64
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "uso: batch <input.jsonl>")
		os.Exit(1)
	}

	inputPath := os.Args[1]

	file, err := os.Open(inputPath)
	if err != nil {
		fmt.Fprintf(stderr, "erro ao abrir o arquivo de entrada: %v\n", err)
	}
	defer file.Close()

	stats, err := process(file)
	if err != nil {
		fmt.Fprintf(stderr, "erro ao processar arquivo: %v\n", err)
	}

	printStats(stdout, stats)

	return 0
}

func process(r io.Reader) (Stats, error) {
	reader := bufio.NewReader(r)

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
			return stats, err
		}
	}
	return stats, nil
}

func isJSONL(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ".jsonl")
}

func printStats(stdout io.Writer, stats Stats) {
	w := tabwriter.NewWriter(stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "Resultados do Batch")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Métrica\tTotal")
	fmt.Fprintf(w, "Registros lidos\t%d\n", stats.RecordsRead)
}
