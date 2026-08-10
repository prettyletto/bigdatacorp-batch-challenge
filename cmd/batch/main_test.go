package main

import (
	"bytes"
	"encoding/csv"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var errOutputUnavailable = errors.New("output unavailable")

type errorWriter struct{}

func (errorWriter) Write([]byte) (int, error) {
	return 0, errOutputUnavailable
}

func TestRunWithoutInput(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run(nil, &stdout, &stderr)

	if code == 0 {
		t.Fatal("esperava código de saída diferente de zero.")
	}

	if stdout.Len() != 0 {
		t.Fatalf("stdout deveria estar vazio, recebeu %q", stdout.String())
	}

	if !strings.Contains(stderr.String(), "uso:") {
		t.Fatalf(
			"esperava mensagem de uso em stderr, recebeu %q",
			stderr.String(),
		)
	}
}

func TestRunRejectNonJSONLInput(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run(
		[]string{"clubes.json"},
		&stdout,
		&stderr,
	)

	if code == 0 {
		t.Fatal("esperava código de saída diferente de zero.")
	}

	if !strings.Contains(stderr.String(), ".jsonl") {
		t.Fatalf(
			"esperava erro sobre extensão .jsonl, recebeu %q",
			stderr.String(),
		)
	}
}

func TestRunWithMissingFile(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	missingPath := filepath.Join(t.TempDir(), "arquivo-que-nao-existe.jsonl")

	code := run(
		[]string{missingPath},
		&stdout,
		&stderr,
	)

	if code == 0 {
		t.Fatal("esperava código de saída diferente de zero.")
	}

	if !strings.Contains(stderr.String(), "erro ao abrir") {
		t.Fatalf(
			"esperava erro ao abrir arquivo, recebeu %q",
			stderr.String(),
		)
	}
}

func TestRunProcessesValidJSONLFile(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	inputPath, err := filepath.Abs(filepath.Join("..", "..", "sample_clubes.jsonl"))
	if err != nil {
		t.Fatalf("resolvendo caminho da amostra: %v", err)
	}
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatalf("obtendo diretorio de trabalho: %v", err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatalf("mudando diretorio de trabalho: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(workingDirectory) })

	code := run([]string{inputPath}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("codigo de saida = %d, queria 0; stderr = %q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr deveria estar vazio, recebeu %q", stderr.String())
	}
	if !strings.Contains(stdout.String(), "Registros lidos                   6") {
		t.Fatalf("esperava contagem da amostra em stdout, recebeu %q", stdout.String())
	}
	file, err := os.Open("clubs.csv")
	if err != nil {
		t.Fatalf("abrindo clubs.csv gerado: %v", err)
	}
	defer file.Close()
	rows, err := csv.NewReader(file).ReadAll()
	if err != nil {
		t.Fatalf("lendo clubs.csv gerado: %v", err)
	}
	if len(rows) != 6 {
		t.Fatalf("linhas em clubs.csv = %d, queria 6", len(rows))
	}
}

func TestIsJSONL(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{name: "extensao minuscula", path: "clubes.jsonl", want: true},
		{name: "extensao maiuscula", path: "CLUBES.JSONL", want: true},
		{name: "sem extensao", path: "clubes", want: false},
		{name: "extensao diferente", path: "clubes.json", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isJSONL(tt.path); got != tt.want {
				t.Fatalf("isJSONL(%q) = %t, queria %t", tt.path, got, tt.want)
			}
		})
	}
}

func TestFilterChampionship(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
		ok    bool
	}{
		{name: "serie A", value: "SERIE A", want: "SERIE A", ok: true},
		{name: "serie B", value: "SERIE B", want: "SERIE B", ok: true},
		{name: "espacos ao redor", value: "  SERIE A  ", want: "SERIE A", ok: true},
		{name: "campeonato nao elegivel", value: "SERIE C", want: "", ok: false},
		{name: "campeonato ausente", value: "", want: "", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := filterChampionship(tt.value)
			if got != tt.want || ok != tt.ok {
				t.Fatalf("filterChampionship(%q) = (%q, %t), queria (%q, %t)", tt.value, got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestProcessCountsAllRecordsAndSkipsMalformedOnes(t *testing.T) {
	input := strings.NewReader("{\"club_id\":\"SCCP\",\"championship\":\"SERIE A\"}\njson invalido\n{\"club_id\":\"SEP\",\"championship\":\"SERIE B\"}")

	stats, err := process(input, &bytes.Buffer{})

	if err != nil {
		t.Fatalf("process retornou erro: %v", err)
	}
	if stats.RecordsRead != 3 {
		t.Fatalf("Registros lidos = %d, queria 3", stats.RecordsRead)
	}
	if stats.MalformedRecords != 1 {
		t.Fatalf("Registros malformados = %d, queria 1", stats.MalformedRecords)
	}
	if stats.SerieAClubs != 1 || stats.SerieBclubs != 1 || stats.FilteredChampionship != 0 {
		t.Fatalf("metricas de campeonato = %+v, queria 1 clube em cada serie e nenhum filtrado", stats)
	}
}

func TestProcessHandlesEmptyInput(t *testing.T) {
	var output bytes.Buffer
	stats, err := process(strings.NewReader(""), &output)

	if err != nil {
		t.Fatalf("process retornou erro: %v", err)
	}
	if stats != (Stats{}) {
		t.Fatalf("estatisticas = %+v, queria estatisticas vazias", stats)
	}
	if output.String() != "Id do Clube,Nome,Campeonato,Data de Fundação,Cidade,Estado,País,Estádio,Presidente,Apelido,Cores\n" {
		t.Fatalf("saida = %q, queria apenas o cabecalho", output.String())
	}
}

func TestProcessReadsSampleFile(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "sample_clubes.jsonl"))
	if err != nil {
		t.Fatalf("abrindo arquivo de amostra: %v", err)
	}
	defer file.Close()

	var output bytes.Buffer
	stats, err := process(file, &output)

	if err != nil {
		t.Fatalf("process retornou erro: %v", err)
	}
	if stats.RecordsRead != 6 {
		t.Fatalf("Registros lidos = %d, queria 6", stats.RecordsRead)
	}
	if stats.MalformedRecords != 0 {
		t.Fatalf("Registros malformados = %d, queria 0", stats.MalformedRecords)
	}
	if stats.SerieAClubs != 3 {
		t.Fatalf("Registros SERIE A = %d, queria 3", stats.SerieAClubs)
	}
	if stats.SerieBclubs != 2 {
		t.Fatalf("Registros SERIE B = %d, queria 2", stats.SerieBclubs)
	}
	if stats.FilteredChampionship != 1 {
		t.Fatalf("Registros filtrados = %d, queria 1", stats.FilteredChampionship)
	}
	if stats.ClubRowsWritten != 5 {
		t.Fatalf("Linhas de clubes geradas = %d, queria 5", stats.ClubRowsWritten)
	}
	rows, err := csv.NewReader(strings.NewReader(output.String())).ReadAll()
	if err != nil {
		t.Fatalf("lendo CSV gerado: %v", err)
	}
	if len(rows) != 6 {
		t.Fatalf("linhas no CSV = %d, queria 6", len(rows))
	}
}

func TestProcessEscapesCSVFields(t *testing.T) {
	input := strings.NewReader("{\"club_id\":\"SCCP\",\"name\":\"Clube, do \\\"Povo\\\"\",\"championship\":\"SERIE A\",\"president\":\"Ana\\nSilva\"}")
	var output bytes.Buffer

	_, err := process(input, &output)

	if err != nil {
		t.Fatalf("process retornou erro: %v", err)
	}
	rows, err := csv.NewReader(strings.NewReader(output.String())).ReadAll()
	if err != nil {
		t.Fatalf("lendo CSV gerado: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("linhas no CSV = %d, queria 2", len(rows))
	}
	if rows[1][1] != "Clube, do \"Povo\"" {
		t.Fatalf("nome = %q, queria campo com virgula e aspas preservadas", rows[1][1])
	}
	if rows[1][8] != "Ana\nSilva" {
		t.Fatalf("presidente = %q, queria campo com quebra de linha preservada", rows[1][8])
	}
}

func TestProcessReturnsOutputErrors(t *testing.T) {
	stats, err := process(strings.NewReader(""), errorWriter{})

	if !errors.Is(err, errOutputUnavailable) {
		t.Fatalf("erro = %v, queria erro do escritor", err)
	}
	if stats != (Stats{}) {
		t.Fatalf("estatisticas = %+v, queria estatisticas vazias", stats)
	}
}

func TestClubRow(t *testing.T) {
	nickname := "Timão"
	tests := []struct {
		name string
		club Club
		want []string
	}{
		{
			name: "campos transformados",
			club: Club{
				ClubID:       "SCCP",
				Name:         "Sport Club Corinthians Paulista",
				Championship: "SERIE A",
				FoundingDate: "1910-09-01",
				City:         "São Paulo",
				State:        "SP",
				Country:      "Brasil",
				Stadium:      "Neo Química Arena",
				President:    "Augusto Melo",
				Nickname:     &nickname,
				Colors:       []string{"preto", "branco"},
			},
			want: []string{"SCCP", "Sport Club Corinthians Paulista", "SERIE A", "1910-09-01", "São Paulo", "SP", "Brasil", "Neo Química Arena", "Augusto Melo", "Timão", "preto|branco"},
		},
		{
			name: "campos opcionais e data invalida",
			club: Club{
				FoundingDate: "2024-02-30",
			},
			want: []string{"", "", "", "", "", "", "", "", "", "", ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := clubRow(tt.club)
			if strings.Join(got, "\x00") != strings.Join(tt.want, "\x00") {
				t.Fatalf("clubRow() = %#v, queria %#v", got, tt.want)
			}
		})
	}
}

func TestNormalizeDate(t *testing.T) {
	tests := []struct {
		value string
		want  string
	}{
		{value: "1910-09-01", want: "1910-09-01"},
		{value: " 1910-09-01 ", want: "1910-09-01"},
		{value: "1910-09-31", want: ""},
		{value: "01/09/1910", want: ""},
		{value: "", want: ""},
	}

	for _, tt := range tests {
		if got := normalizeDate(tt.value); got != tt.want {
			t.Fatalf("normalizeDate(%q) = %q, queria %q", tt.value, got, tt.want)
		}
	}
}

func TestPrintStats(t *testing.T) {
	var output bytes.Buffer

	printStats(&output, Stats{RecordsRead: 12, MalformedRecords: 3})

	want := "Resultados do Batch\n\nMétrica                           Total\nRegistros lidos                   12\nRegistros Malformados             3\nClubes Filtrados(SEM CAMPEONATO)  0\nClubes Série A                    0\nClubes Série B                    0\nLinhas de Clubes Geradas          0\n"
	if output.String() != want {
		t.Fatalf("saida = %q, queria %q", output.String(), want)
	}
}
