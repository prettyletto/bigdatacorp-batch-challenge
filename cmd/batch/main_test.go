package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/prettyletto/bigdatacorp-batch-challenge/internal/batch"
)

func TestRunWithoutInput(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run(nil, &stdout, &stderr)

	if code == 0 {
		t.Fatal("expected a non-zero exit code")
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout should be empty, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "uso:") {
		t.Fatalf("expected usage message in stderr, got %q", stderr.String())
	}
}

func TestRunRejectsNonJSONLInput(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"clubes.json"}, &stdout, &stderr)

	if code == 0 {
		t.Fatal("expected a non-zero exit code")
	}
	if !strings.Contains(stderr.String(), ".jsonl") {
		t.Fatalf("expected JSONL extension error, got %q", stderr.String())
	}
}

func TestRunRejectsMissingFile(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{filepath.Join(t.TempDir(), "missing.jsonl")}, &stdout, &stderr)

	if code == 0 {
		t.Fatal("expected a non-zero exit code")
	}
	if !strings.Contains(stderr.String(), "erro ao abrir") {
		t.Fatalf("expected input opening error, got %q", stderr.String())
	}
}

func TestRunGeneratesCSVFilesFromSample(t *testing.T) {
	inputPath, err := filepath.Abs(filepath.Join("..", "..", "sample_clubes.jsonl"))
	if err != nil {
		t.Fatalf("resolving sample path: %v", err)
	}
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatalf("getting working directory: %v", err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatalf("changing working directory: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(workingDirectory) })

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{inputPath}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr should be empty, got %q", stderr.String())
	}
	for _, filename := range []string{"clubs.csv", "players.csv"} {
		if _, err := os.Stat(filename); err != nil {
			t.Fatalf("expected %s to be generated: %v", filename, err)
		}
	}
	if !strings.Contains(stdout.String(), "Registros lidos") {
		t.Fatalf("expected statistics in stdout, got %q", stdout.String())
	}
}

func TestIsJSONL(t *testing.T) {
	for _, tt := range []struct {
		path string
		want bool
	}{
		{path: "clubes.jsonl", want: true},
		{path: "CLUBES.JSONL", want: true},
		{path: "clubes", want: false},
		{path: "clubes.json", want: false},
	} {
		if got := isJSONL(tt.path); got != tt.want {
			t.Fatalf("isJSONL(%q) = %t, want %t", tt.path, got, tt.want)
		}
	}
}

func TestPrintStats(t *testing.T) {
	var output bytes.Buffer

	printStats(&output, batch.Stats{RecordsRead: 12, MalformedRecords: 3})

	want := "Resultados do Batch\n\nMétrica                          Total\nRegistros lidos                  12\nRegistros Malformados            3\nClubes ignorados por campeonato  0\nClubes Série A                   0\nClubes Série B                   0\nLinhas de Clubes Geradas         0\nLinhas de Jogadores Geradas      0\n"
	if output.String() != want {
		t.Fatalf("output = %q, want %q", output.String(), want)
	}
}
