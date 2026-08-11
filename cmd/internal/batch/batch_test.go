package batch

import (
	"bytes"
	"encoding/csv"
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

var errOutputUnavailable = errors.New("output unavailable")

type errorWriter struct{}

func (errorWriter) Write([]byte) (int, error) {
	return 0, errOutputUnavailable
}

func TestFilterChampionship(t *testing.T) {
	for _, tt := range []struct {
		value string
		want  string
		ok    bool
	}{
		{value: "SERIE A", want: "SERIE A", ok: true},
		{value: "SERIE B", want: "SERIE B", ok: true},
		{value: "  SERIE A  ", want: "SERIE A", ok: true},
		{value: "SERIE C"},
		{value: ""},
	} {
		got, ok := filterChampionship(tt.value)
		if got != tt.want || ok != tt.ok {
			t.Fatalf("filterChampionship(%q) = (%q, %t), want (%q, %t)", tt.value, got, ok, tt.want, tt.ok)
		}
	}
}

func TestProcessCountsMalformedRecordsAndEligibleClubs(t *testing.T) {
	input := strings.NewReader("{\"club_id\":\"SCCP\",\"championship\":\"SERIE A\"}\ninvalid json\n{\"club_id\":\"SEP\",\"championship\":\"SERIE B\"}")

	stats, err := Process(input, &bytes.Buffer{}, &bytes.Buffer{})

	if err != nil {
		t.Fatalf("Process returned error: %v", err)
	}
	want := Stats{RecordsRead: 3, MalformedRecords: 1, SerieAClubs: 1, SerieBClubs: 1, ClubRowsWritten: 2}
	if stats != want {
		t.Fatalf("stats = %+v, want %+v", stats, want)
	}
}

func TestProcessWritesHeadersForEmptyInput(t *testing.T) {
	var clubsOutput bytes.Buffer
	var playersOutput bytes.Buffer

	stats, err := Process(strings.NewReader(""), &clubsOutput, &playersOutput)

	if err != nil {
		t.Fatalf("Process returned error: %v", err)
	}
	if stats != (Stats{}) {
		t.Fatalf("stats = %+v, want empty stats", stats)
	}
	if clubsOutput.String() != "Id do Clube,Nome,Campeonato,Data de Fundação,Cidade,Estado,País,Estádio,Presidente,Apelido,Cores\n" {
		t.Fatalf("clubs output = %q, want only the header", clubsOutput.String())
	}
	if playersOutput.String() != "Id do Clube,Id do Jogador,Nome,Idade,Gols,Data de Estreia,Posição,Número da Camisa\n" {
		t.Fatalf("players output = %q, want only the header", playersOutput.String())
	}
}

func TestProcessReadsSampleFile(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "..", "sample_clubes.jsonl"))
	if err != nil {
		t.Fatalf("opening sample file: %v", err)
	}
	defer file.Close()

	var clubsOutput bytes.Buffer
	var playersOutput bytes.Buffer
	stats, err := Process(file, &clubsOutput, &playersOutput)

	if err != nil {
		t.Fatalf("Process returned error: %v", err)
	}
	wantStats := Stats{RecordsRead: 6, FilteredChampionship: 1, SerieAClubs: 3, SerieBClubs: 2, ClubRowsWritten: 5, PlayerRowsWritten: 8}
	if stats != wantStats {
		t.Fatalf("stats = %+v, want %+v", stats, wantStats)
	}
	if got := readCSV(t, clubsOutput.String()); len(got) != 6 {
		t.Fatalf("clubs rows = %d, want 6", len(got))
	}
	if got := readCSV(t, playersOutput.String()); len(got) != 9 {
		t.Fatalf("players rows = %d, want 9", len(got))
	}
}

func TestProcessMalformedAndMissingFieldsFixture(t *testing.T) {
	stats, clubsRows, playersRows := processFixture(t, "malformed_and_missing_fields.jsonl")

	wantStats := Stats{RecordsRead: 2, MalformedRecords: 1, SerieAClubs: 1, ClubRowsWritten: 1, PlayerRowsWritten: 1}
	if stats != wantStats {
		t.Fatalf("stats = %+v, want %+v", stats, wantStats)
	}
	if got, want := clubsRows[1], []string{"INC", "", "SERIE A", "", "", "", "", "", "", "", ""}; !slices.Equal(got, want) {
		t.Fatalf("club row = %#v, want %#v", got, want)
	}
	if got, want := playersRows[1], []string{"INC", "", "Jogador sem identificador", "", "", "", "", ""}; !slices.Equal(got, want) {
		t.Fatalf("player row = %#v, want %#v", got, want)
	}
}

func TestProcessInvalidDatesFixture(t *testing.T) {
	stats, clubsRows, playersRows := processFixture(t, "invalid_dates.jsonl")

	wantStats := Stats{RecordsRead: 1, SerieAClubs: 1, ClubRowsWritten: 1, PlayerRowsWritten: 1}
	if stats != wantStats {
		t.Fatalf("stats = %+v, want %+v", stats, wantStats)
	}
	if clubsRows[1][3] != "" {
		t.Fatalf("founding date = %q, want empty value for invalid date", clubsRows[1][3])
	}
	if playersRows[1][5] != "" {
		t.Fatalf("debut date = %q, want empty value for invalid date", playersRows[1][5])
	}
}

func TestProcessClubsWithoutPlayersFixture(t *testing.T) {
	stats, clubsRows, playersRows := processFixture(t, "clubs_without_players.jsonl")

	wantStats := Stats{RecordsRead: 2, SerieAClubs: 1, SerieBClubs: 1, ClubRowsWritten: 2}
	if stats != wantStats {
		t.Fatalf("stats = %+v, want %+v", stats, wantStats)
	}
	if len(clubsRows) != 3 {
		t.Fatalf("clubs rows = %d, want 3", len(clubsRows))
	}
	if len(playersRows) != 1 {
		t.Fatalf("players rows = %d, want only the header", len(playersRows))
	}
	if clubsRows[1][10] != "" || clubsRows[2][10] != "" {
		t.Fatalf("colors = %q and %q, want empty values", clubsRows[1][10], clubsRows[2][10])
	}
}

func TestProcessCSVEscapingFixture(t *testing.T) {
	stats, clubsRows, playersRows := processFixture(t, "csv_escaping.jsonl")

	wantStats := Stats{RecordsRead: 1, SerieBClubs: 1, ClubRowsWritten: 1, PlayerRowsWritten: 1}
	if stats != wantStats {
		t.Fatalf("stats = %+v, want %+v", stats, wantStats)
	}
	if got, want := clubsRows[1][1], "Clube, \"Especial\""; got != want {
		t.Fatalf("club name = %q, want %q", got, want)
	}
	if got, want := clubsRows[1][4], "Rio\nClaro"; got != want {
		t.Fatalf("city = %q, want %q", got, want)
	}
	if got, want := clubsRows[1][8], "Ana, \"A\""; got != want {
		t.Fatalf("president = %q, want %q", got, want)
	}
	if got, want := clubsRows[1][10], "azul, claro|branco"; got != want {
		t.Fatalf("colors = %q, want %q", got, want)
	}
	if got, want := playersRows[1][1], "CSV-1"; got != want {
		t.Fatalf("player ID = %q, want %q", got, want)
	}
	if got, want := playersRows[1][2], "Jogador, \"Um\""; got != want {
		t.Fatalf("player name = %q, want %q", got, want)
	}
	if got, want := playersRows[1][6], "Meia\nOfensivo"; got != want {
		t.Fatalf("player position = %q, want %q", got, want)
	}
}

func TestProcessWritesPlayersOnlyForEligibleClubs(t *testing.T) {
	input := strings.NewReader(strings.Join([]string{
		`{"club_id":"SCCP","championship":"SERIE A","players":[{"player_id":"SCCP-10","name":"Rodrigo Garro","age":26,"goals":8,"debut_date":"2024-01-18","position":"Meia","shirt_number":10}]}`,
		`{"club_id":"AVA","championship":"SERIE B","players":[]}`,
		`{"club_id":"NAC","championship":"SEM CAMPEONATO","players":[{"player_id":"NAC-1"}]}`,
	}, "\n"))
	var clubsOutput bytes.Buffer
	var playersOutput bytes.Buffer

	stats, err := Process(input, &clubsOutput, &playersOutput)

	if err != nil {
		t.Fatalf("Process returned error: %v", err)
	}
	wantStats := Stats{RecordsRead: 3, FilteredChampionship: 1, SerieAClubs: 1, SerieBClubs: 1, ClubRowsWritten: 2, PlayerRowsWritten: 1}
	if stats != wantStats {
		t.Fatalf("stats = %+v, want %+v", stats, wantStats)
	}
	wantRows := [][]string{
		{"Id do Clube", "Id do Jogador", "Nome", "Idade", "Gols", "Data de Estreia", "Posição", "Número da Camisa"},
		{"SCCP", "SCCP-10", "Rodrigo Garro", "26", "8", "2024-01-18", "Meia", "10"},
	}
	if got := readCSV(t, playersOutput.String()); !slices.EqualFunc(got, wantRows, slices.Equal[[]string]) {
		t.Fatalf("players rows = %#v, want %#v", got, wantRows)
	}
}

func TestProcessEscapesCSVFields(t *testing.T) {
	input := strings.NewReader("{\"club_id\":\"SCCP\",\"name\":\"Clube, do \\\"Povo\\\"\",\"championship\":\"SERIE A\",\"president\":\"Ana\\nSilva\"}")
	var clubsOutput bytes.Buffer
	var playersOutput bytes.Buffer

	_, err := Process(input, &clubsOutput, &playersOutput)

	if err != nil {
		t.Fatalf("Process returned error: %v", err)
	}
	rows := readCSV(t, clubsOutput.String())
	if rows[1][1] != "Clube, do \"Povo\"" || rows[1][8] != "Ana\nSilva" {
		t.Fatalf("CSV fields = %#v, want escaped values to round-trip", rows[1])
	}
}

func TestProcessReturnsOutputErrors(t *testing.T) {
	for _, tt := range []struct {
		clubsOutput   io.Writer
		playersOutput io.Writer
	}{
		{clubsOutput: errorWriter{}, playersOutput: &bytes.Buffer{}},
		{clubsOutput: &bytes.Buffer{}, playersOutput: errorWriter{}},
	} {
		stats, err := Process(strings.NewReader(""), tt.clubsOutput, tt.playersOutput)
		if !errors.Is(err, errOutputUnavailable) {
			t.Fatalf("error = %v, want output writer error", err)
		}
		if stats != (Stats{}) {
			t.Fatalf("stats = %+v, want empty stats", stats)
		}
	}
}

func TestClubRow(t *testing.T) {
	nickname := "Timão"
	got := clubRow(Club{ClubID: "SCCP", Name: "Corinthians", Championship: "SERIE A", FoundingDate: "1910-09-01", Nickname: &nickname, Colors: []string{"preto", "branco"}})
	want := []string{"SCCP", "Corinthians", "SERIE A", "1910-09-01", "", "", "", "", "", "Timão", "preto|branco"}
	if !slices.Equal(got, want) {
		t.Fatalf("clubRow() = %#v, want %#v", got, want)
	}
}

func TestPlayerRowHandlesEmptyAndZeroValues(t *testing.T) {
	zero := 0
	got := playerRow("SCCP", Player{Age: &zero, Goals: &zero, DebutDate: "2024-02-30", ShirtNumber: &zero})
	want := []string{"SCCP", "", "", "0", "0", "", "", "0"}
	if !slices.Equal(got, want) {
		t.Fatalf("playerRow() = %#v, want %#v", got, want)
	}
}

func TestNormalizeDate(t *testing.T) {
	for _, tt := range []struct{ value, want string }{
		{value: "1910-09-01", want: "1910-09-01"},
		{value: " 1910-09-01 ", want: "1910-09-01"},
		{value: "1910-09-31"},
		{value: "01/09/1910"},
		{value: ""},
	} {
		if got := normalizeDate(tt.value); got != tt.want {
			t.Fatalf("normalizeDate(%q) = %q, want %q", tt.value, got, tt.want)
		}
	}
}

func TestIntValue(t *testing.T) {
	zero := 0
	for _, tt := range []struct {
		value *int
		want  string
	}{
		{value: nil, want: ""},
		{value: &zero, want: "0"},
		{value: intPointer(42), want: "42"},
	} {
		if got := intValue(tt.value); got != tt.want {
			t.Fatalf("intValue(%v) = %q, want %q", tt.value, got, tt.want)
		}
	}
}

func readCSV(t *testing.T, output string) [][]string {
	t.Helper()
	rows, err := csv.NewReader(strings.NewReader(output)).ReadAll()
	if err != nil {
		t.Fatalf("reading CSV: %v", err)
	}
	return rows
}

func processFixture(t *testing.T, filename string) (Stats, [][]string, [][]string) {
	t.Helper()
	file, err := os.Open(filepath.Join("testdata", filename))
	if err != nil {
		t.Fatalf("opening fixture %q: %v", filename, err)
	}
	defer file.Close()

	var clubsOutput bytes.Buffer
	var playersOutput bytes.Buffer
	stats, err := Process(file, &clubsOutput, &playersOutput)
	if err != nil {
		t.Fatalf("processing fixture %q: %v", filename, err)
	}

	return stats, readCSV(t, clubsOutput.String()), readCSV(t, playersOutput.String())
}

func intPointer(value int) *int {
	return &value
}
