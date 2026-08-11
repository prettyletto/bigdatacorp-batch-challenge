package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/prettyletto/bigdatacorp-batch-challenge/internal/batch"
)

type atomicOutput struct {
	file        *os.File
	destination string
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet(
		"batch",
		flag.ContinueOnError,
	)

	flags.SetOutput(stderr)

	var workers int

	flags.IntVar(
		&workers,
		"workers",
		1,
		"quantidadede workers para processamento",
	)

	flags.Usage = func() {
		fmt.Fprintln(stderr, "uso: batch [-workers N] <input.jsonl>")
	}

	if err := flags.Parse(args); err != nil {
		return 1
	}

	if flags.NArg() != 1 {
		flags.Usage()
		return 1
	}

	if workers < 1 {
		fmt.Fprintln(stderr, "erro: workers deve ser maior que zero")
		return 1
	}

	inputPath := flags.Arg(0)

	if !isJSONL(inputPath) {
		fmt.Fprintf(stderr, "erro: arquivo fora da extensão .jsonl")
		return 1
	}

	inputFile, err := os.Open(inputPath)
	if err != nil {
		fmt.Fprintf(stderr, "erro ao abrir o arquivo de entrada: %v\n", err)
		return 1
	}
	defer inputFile.Close()

	clubsOutput, err := createAtomicOutput("clubs.csv")
	if err != nil {
		fmt.Fprintf(stderr, "erro ao criar clubs.csv: %v\n", err)
		return 1
	}
	defer clubsOutput.Discard()

	playersOutput, err := createAtomicOutput("players.csv")
	if err != nil {
		fmt.Fprintf(stderr, "erro ao criar players.csv: %v\n", err)
		return 1
	}
	defer playersOutput.Discard()

	stats, err := batch.ProcessWithOptions(inputFile, clubsOutput.file, playersOutput.file, batch.Options{Workers: workers})
	if err != nil {
		fmt.Fprintf(stderr, "erro ao processar arquivo: %v\n", err)
		return 1
	}
	if err := clubsOutput.Commit(); err != nil {
		fmt.Fprintf(stderr, "erro ao finalizar clubs.csv: %v\n", err)
		return 1
	}
	if err := playersOutput.Commit(); err != nil {
		fmt.Fprintf(stderr, "erro ao finalizar players.csv: %v\n", err)
		return 1
	}

	printStats(stdout, stats)

	return 0
}

func createAtomicOutput(destination string) (*atomicOutput, error) {
	file, err := os.CreateTemp(
		filepath.Dir(destination),
		"."+filepath.Base(destination)+"-*",
	)
	if err != nil {
		return nil, err
	}

	return &atomicOutput{file: file, destination: destination}, nil
}

func (output *atomicOutput) Commit() error {
	if err := output.file.Close(); err != nil {
		return err
	}

	return os.Rename(output.file.Name(), output.destination)
}

func (output *atomicOutput) Discard() {
	_ = output.file.Close()
	_ = os.Remove(output.file.Name())
}

func isJSONL(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ".jsonl")
}

func printStats(stdout io.Writer, stats batch.Stats) {
	w := tabwriter.NewWriter(stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "Resultados do Batch")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Métrica\tTotal")
	fmt.Fprintf(w, "Registros lidos\t%d\n", stats.RecordsRead)
	fmt.Fprintf(w, "Registros Malformados\t%d\n", stats.MalformedRecords)
	fmt.Fprintf(w, "Clubes ignorados por campeonato\t%d\n", stats.FilteredChampionship)
	fmt.Fprintf(w, "Clubes Série A\t%d\n", stats.SerieAClubs)
	fmt.Fprintf(w, "Clubes Série B\t%d\n", stats.SerieBClubs)
	fmt.Fprintf(w, "Linhas de Clubes Geradas \t%d\n", stats.ClubRowsWritten)
	fmt.Fprintf(w, "Linhas de Jogadores Geradas \t%d\n", stats.PlayerRowsWritten)
}
