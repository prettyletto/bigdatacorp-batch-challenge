// provavelmente cli/engine depois.
package main

import (
	"bytes"
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

	code := run(
		[]string{"arquivo-que-nao-existe.jsonl"},
		&stdout,
		&stderr,
	)

	if code == 0 {
		t.Fatal("esperava código de saída diferente de zero.")
	}

	if !strings.Contains(stderr.String(), ".jsonl") {
		t.Fatalf(
			"esperava erro ao abrir arquivo, recebeu %q",
			stderr.String(),
		)
	}
}

