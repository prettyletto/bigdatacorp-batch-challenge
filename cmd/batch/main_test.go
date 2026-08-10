package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
	originalArgs := os.Args
	os.Args = []string{"batch", missingPath}
	t.Cleanup(func() { os.Args = originalArgs })

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
	inputPath := filepath.Join("..", "..", "sample_clubes.jsonl")
	originalArgs := os.Args
	os.Args = []string{"batch", inputPath}
	t.Cleanup(func() { os.Args = originalArgs })

	code := run([]string{inputPath}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("codigo de saida = %d, queria 0; stderr = %q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr deveria estar vazio, recebeu %q", stderr.String())
	}
	if !strings.Contains(stdout.String(), "Registros lidos                      6") {
		t.Fatalf("esperava contagem da amostra em stdout, recebeu %q", stdout.String())
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

	stats, err := process(input)

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
	stats, err := process(strings.NewReader(""))

	if err != nil {
		t.Fatalf("process retornou erro: %v", err)
	}
	if stats != (Stats{}) {
		t.Fatalf("estatisticas = %+v, queria estatisticas vazias", stats)
	}
}

func TestProcessReadsSampleFile(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "sample_clubes.jsonl"))
	if err != nil {
		t.Fatalf("abrindo arquivo de amostra: %v", err)
	}
	defer file.Close()

	stats, err := process(file)

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
}

func TestPrintStats(t *testing.T) {
	var output bytes.Buffer

	printStats(&output, Stats{RecordsRead: 12, MalformedRecords: 3})

	want := "Resultados do Batch\n\nMétrica                              Total\nRegistros lidos                      12\nRegistros Malformados                3\nRegistros SERIE A                    0\nRegistros SERIE B                    0\nRegistros Filtrados(SEM CAMPEONATO)  0\n"
	if output.String() != want {
		t.Fatalf("saida = %q, queria %q", output.String(), want)
	}
}
