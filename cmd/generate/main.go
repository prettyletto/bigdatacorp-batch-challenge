package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

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
	Nickname     string   `json:"nickname"`
	Colors       []string `json:"colors"`
	Titles       int      `json:"titles"`
	Players      []Player `json:"players"`
}

type Player struct {
	PlayerID    string `json:"player_id"`
	Name        string `json:"name"`
	Age         int    `json:"age"`
	Goals       int    `json:"goals"`
	DebutDate   string `json:"debut_date"`
	Position    string `json:"position"`
	ShirtNumber int    `json:"shirt_number"`
	Nationality string `json:"nationality"`
	MarketValue int    `json:"market_value"`
}

func main() {
	var (
		records int
		output  string
		players int
	)

	flag.IntVar(
		&records,
		"records",
		100_000,
		"quantidade de clubes a gerar",
	)

	flag.IntVar(
		&players,
		"players",
		2,
		"quantidade de jogadores por clube",
	)

	flag.StringVar(
		&output,
		"output",
		".local/generated.jsonl",
		"arquivo JSONL de saída",
	)

	flag.Parse()

	if records <= 0 {
		fmt.Fprintln(os.Stderr, "erro: records deve ser maior que zero")
		os.Exit(1)
	}

	if players < 0 {
		fmt.Fprintln(os.Stderr, "erro: players não pode ser negativo")
		os.Exit(1)
	}

	if err := generate(output, records, players); err != nil {
		fmt.Fprintf(os.Stderr, "erro ao gerar arquivo: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("clubes gerados:%d com %d jogadores por clube em %s\n", records, players, output)
}

func generate(path string, records, playersPerClub int) error {
	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("criando diretório de saída: %w", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriterSize(file, 1024*1024)
	encoder := json.NewEncoder(writer)

	for i := 0; i < records; i++ {
		if err := encoder.Encode(makeClub(i, playersPerClub)); err != nil {
			return fmt.Errorf("registro %d: %w", i, err)
		}
	}

	if err := writer.Flush(); err != nil {
		return err
	}

	return file.Sync()
}

func makeClub(index, playersPerClub int) Club {
	clubID := fmt.Sprintf("CLUB-%08d", index)

	championship := "SERIE A"

	switch index % 3 {
	case 1:
		championship = "SERIE B"
	case 2:
		championship = "SERIE C"
	}

	players := make([]Player, playersPerClub)

	for i := 0; i < playersPerClub; i++ {
		players[i] = Player{
			PlayerID: fmt.Sprintf(
				"%s-%03d",
				clubID,
				i+1,
			),
			Name:        fmt.Sprintf("Jogador %d", i+1),
			Age:         18 + ((index + i) % 20),
			Goals:       (index + i) % 25,
			DebutDate:   "2024-01-18",
			Position:    playerPosition(i),
			ShirtNumber: i + 1,
			Nationality: "Brasil",
			MarketValue: 1_000_000 + ((index + i) % 10_000_000),
		}
	}
	return Club{
		ClubID:       clubID,
		Name:         fmt.Sprintf("Clube %d", index),
		Championship: championship,
		FoundingDate: "2000-01-01",
		City:         "Fortaleza",
		State:        "CE",
		Country:      "Brasil",
		Stadium:      fmt.Sprintf("Estádio %d", index),
		President:    fmt.Sprintf("Presidente %d", index),
		Nickname:     fmt.Sprintf("Clube %d", index),
		Colors:       []string{"azul", "branco"},
		Titles:       index % 50,
		Players:      players,
	}
}

func playerPosition(index int) string {
	switch index % 4 {
	case 0:
		return "Goleiro"
	case 1:
		return "Defensor"
	case 2:
		return "Meia"
	default:
		return "Atacante"
	}
}
