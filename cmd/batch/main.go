package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"
)

type Stats struct {
	RecordsRead          int64
	MalformedRecords     int64
	FilteredChampionship int64
	SerieAClubs          int64
	SerieBClubs          int64
	ClubRowsWritten      int64
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

var clubHeader = []string{
	"Id do Clube",
	"Nome",
	"Campeonato",
	"Data de Fundação",
	"Cidade",
	"Estado",
	"País",
	"Estádio",
	"Presidente",
	"Apelido",
	"Cores",
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

	outputFile, err := os.Create("clubs.csv")
	if err != nil {
		fmt.Fprintf(stderr, "erro ao criar clubs.csv: %v\n", err)
		return 1
	}

	stats, err := process(inputFile, outputFile)
	if err != nil {
		fmt.Fprintf(stderr, "erro ao processar arquivo: %v\n", err)
		return 1
	}

	outputFile.Close()
	printStats(stdout, stats)

	return 0
}

func process(r io.Reader, output io.Writer) (Stats, error) {
	reader := bufio.NewReader(r)

	clubsWriter := csv.NewWriter(output)

	if err := clubsWriter.Write(clubHeader); err != nil {
		return Stats{}, err
	}

	var stats Stats

	for {
		line, err := reader.ReadBytes('\n')

		if len(line) > 0 {
			stats.RecordsRead++

			var club Club

			if err := json.Unmarshal(line, &club); err != nil {
				stats.MalformedRecords++
				continue
			} else {
				championship, ok := filterChampionship(club.Championship)
				if !ok {
					stats.FilteredChampionship++
					continue
				}

				switch championship {
				case "SERIE A":
					stats.SerieAClubs++
				case "SERIE B":
					stats.SerieBClubs++
				}

				if err := clubsWriter.Write(clubRow(club)); err != nil {
					return stats, err
				}
				stats.ClubRowsWritten++
			}
		}

		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return stats, err
		}
	}

	clubsWriter.Flush()

	if err := clubsWriter.Error(); err != nil {
		return stats, err
	}
	return stats, nil
}

// WRITERS (Later i need to put this into package)
func clubRow(club Club) []string {
	nickname := ""
	if club.Nickname != nil {
		nickname = *club.Nickname
	}

	return []string{
		club.ClubID,
		club.Name,
		club.Championship,
		normalizeDate(club.FoundingDate),
		club.City,
		club.State,
		club.Country,
		club.Stadium,
		club.President,
		nickname,
		strings.Join(club.Colors, "|"),
	}
}

// HELPERS
func isJSONL(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ".jsonl")
}

func filterChampionship(value string) (string, bool) {
	championship := strings.TrimSpace(value)

	switch championship {
	case "SERIE A", "SERIE B":
		return championship, true
	default:
		return "", false
	}
}

func normalizeDate(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		return ""
	}

	return t.Format("2006-01-02")
}

func printStats(stdout io.Writer, stats Stats) {
	w := tabwriter.NewWriter(stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "Resultados do Batch")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Métrica\tTotal")
	fmt.Fprintf(w, "Registros lidos\t%d\n", stats.RecordsRead)
	fmt.Fprintf(w, "Registros Malformados\t%d\n", stats.MalformedRecords)
	fmt.Fprintf(w, "Clubes Filtrados(SEM CAMPEONATO)\t%d\n", stats.FilteredChampionship)
	fmt.Fprintf(w, "Clubes Série A\t%d\n", stats.SerieAClubs)
	fmt.Fprintf(w, "Clubes Série B\t%d\n", stats.SerieBClubs)
	fmt.Fprintf(w, "Linhas de Clubes Geradas \t%d\n", stats.ClubRowsWritten)
}
