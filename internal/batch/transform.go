package batch

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

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

var playerHeader = []string{
	"Id do Clube",
	"Id do Jogador",
	"Nome",
	"Idade",
	"Gols",
	"Data de Estreia",
	"Posição",
	"Número da Camisa",
}

type processedRecord struct {
	malformed bool
	filtered  bool

	championship string
	clubRow      []string
	playersRows  [][]string
}

func processLine(line []byte) processedRecord {
	var club Club

	if err := json.Unmarshal(line, &club); err != nil {
		return processedRecord{
			malformed: true,
		}
	}

	championship, ok := filterChampionship(club.Championship)
	if !ok {
		return processedRecord{
			filtered: true,
		}
	}

	playerRows := make([][]string, len(club.Players))

	for i, player := range club.Players {
		playerRows[i] = playerRow(club.ClubID, player)
	}

	return processedRecord{
		championship: championship,
		clubRow:      clubRow(club),
		playersRows:  playerRows,
	}
}

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

func playerRow(clubID string, player Player) []string {
	return []string{
		clubID,
		player.PlayerID,
		player.Name,
		intValue(player.Age),
		intValue(player.Goals),
		normalizeDate(player.DebutDate),
		player.Position,
		intValue(player.ShirtNumber),
	}
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

func intValue(value *int) string {
	if value == nil {
		return ""
	}

	return strconv.Itoa(*value)
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
