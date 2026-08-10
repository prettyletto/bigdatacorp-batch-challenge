package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
)

type Stats struct {
	RecordsRead      int64
	MalformedRecords int64
}

type Club struct {
	ClubID       string   `json:"club_id"`
	Name         string   `json:"name"`
	Championship string   `json:"championship"`
	FoundingDate string   `json:"founding_date"`
	City         string   `json:"city"`
	State        string   `json:"state"`
	Country      string   `json:"country"`
	Stadium      string   `json:"stadium"`
	President    string   `json:"president"`
	Nickname     *string  `json:"nickname"`
	Players      []Player `json:"players"`
	Colors       []string `json:"colors"`
}

type Player struct {
	PlayerID    string `json:"player_id"`
	Name        string `json:"name"`
	Age         *int   `json:"age"`
	Goals       *int   `json:"goals"`
	DebutDate   string `json:"debut_date"`
	Position    string `json:"position"`
	ShirtNumber *int   `json:"shirt_number"`
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "uso: batch <input.jsonl>")
		return 1
	}

	inputPath := os.Args[1]

	if !isJSONL(inputPath) {
		fmt.Fprintf(stderr, "erro: arquivo fora da extensão .jsonl")
		return 1
	}

	file, err := os.Open(inputPath)
	if err != nil {
		fmt.Fprintf(stderr, "erro ao abrir o arquivo de entrada: %v\n", err)
		return 1
	}
	defer file.Close()

	stats, err := process(file)
	if err != nil {
		fmt.Fprintf(stderr, "erro ao processar arquivo: %v\n", err)
		return 1
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

			var club Club
			if err := json.Unmarshal(line, &club); err != nil {
				stats.MalformedRecords++
				continue
			}
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
	fmt.Fprintf(w,"Registros Malformados\t%d\n", stats.MalformedRecords)
}
