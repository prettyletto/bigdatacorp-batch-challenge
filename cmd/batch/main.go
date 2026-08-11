package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/prettyletto/bigdatacorp-batch-challenge/cmd/internal/batch"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "uso: batch <input.jsonl>")
		return 1
	}

	inputPath := args[0]

	if !isJSONL(inputPath) {
		fmt.Fprintln(stderr, "erro: arquivo fora da extensão .jsonl")
		return 1
	}

	inputFile, err := os.Open(inputPath)
	if err != nil {
		fmt.Fprintf(stderr, "erro ao abrir o arquivo de entrada: %v\n", err)
		return 1
	}
	defer inputFile.Close()

	clubsFile, err := os.Create("clubs.csv")
	if err != nil {
		fmt.Fprintf(stderr, "erro ao criar clubs.csv: %v\n", err)
		return 1
	}
	defer clubsFile.Close()

	playersFile, err := os.Create("players.csv")
	if err != nil {
		fmt.Fprintf(stderr, "erro ao criar players.csv: %v\n", err)
		return 1
	}
	defer playersFile.Close()

	stats, err := batch.Process(inputFile, clubsFile, playersFile)
	if err != nil {
		fmt.Fprintf(stderr, "erro ao processar arquivo: %v\n", err)
		return 1
	}

	printStats(stdout, stats)

	return 0
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
