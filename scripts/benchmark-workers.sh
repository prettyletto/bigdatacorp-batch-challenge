#!/usr/bin/env bash
set -euo pipefail

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
work_dir="$root_dir/.local/benchmark-workers"
batch_bin="$work_dir/batch"
generate_bin="$work_dir/generate"
report="$work_dir/report.md"
summary_rows="$work_dir/summary-rows.md"
details="$work_dir/details.md"
workers=8
record_counts=(100000 1000000 10000000)

info() {
  printf '%s\n' "$*"
}

resolve_gnu_time() {
  local candidate
  local candidates=()

  if [[ -n "${GNU_TIME_BIN:-}" ]]; then
    candidates+=("$GNU_TIME_BIN")
  fi
  candidates+=("/usr/bin/time")
  if command -v gtime > /dev/null 2>&1; then
    candidates+=("$(command -v gtime)")
  fi

  for candidate in "${candidates[@]}"; do
    if [[ -x "$candidate" ]] && "$candidate" -v true > /dev/null 2>&1; then
      printf '%s\n' "$candidate"
      return 0
    fi
  done

  return 1
}

if ! time_bin="$(resolve_gnu_time)"; then
  printf 'erro: o GNU time com suporte a -v e necessario apenas para executar este benchmark.\n' >&2
  printf 'o executavel batch nao depende do GNU time.\n' >&2
  printf 'instale o GNU time ou informe seu caminho com GNU_TIME_BIN=/caminho/para/time.\n' >&2
  exit 1
fi

time_value() {
  local metric="$1"
  local timing_file="$2"
  local value

  value="$(awk -v metric="$metric" 'index($0, metric) { print $NF; exit }' "$timing_file")"
  if [[ -z "$value" ]]; then
    printf 'erro: nao foi possivel ler a metrica "%s" do GNU time em %s.\n' "$metric" "$timing_file" >&2
    return 1
  fi

  printf '%s\n' "$value"
}

wall_seconds() {
  awk -F: '
    NF == 1 { print $1 }
    NF == 2 { print $1 * 60 + $2 }
    NF == 3 { print $1 * 3600 + $2 * 60 + $3 }
  ' <<<"$1"
}

format_mib() {
  awk -v kib="$1" 'BEGIN { printf "%.1f MiB", kib / 1024 }'
}

run_batch() {
  local label="$1"
  local worker_count="$2"
  local input="$3"
  local run_dir="$4"

  mkdir -p "$run_dir"

  (
    cd "$run_dir"
    "$time_bin" -v -o time.txt "$batch_bin" -workers "$worker_count" "$input" > stats.txt
  )
}

mkdir -p "$work_dir"
info "Compilando os binarios locais..."
go build -o "$batch_bin" "$root_dir/cmd/batch"
go build -o "$generate_bin" "$root_dir/cmd/generate"
info "Usando GNU time: $time_bin"

: > "$summary_rows"
: > "$details"

for index in "${!record_counts[@]}"; do
  records="${record_counts[$index]}"
  input="$work_dir/${records}.jsonl"
  sequential_dir="$work_dir/${records}-sequential"
  parallel_dir="$work_dir/${records}-workers-${workers}"

  info "[$((index + 1))/${#record_counts[@]}] Gerando entrada com $records clubes..."
  "$generate_bin" -records "$records" -output "$input" > /dev/null

  info "[$((index + 1))/${#record_counts[@]}] Processando sequencialmente ($records clubes)..."
  run_batch sequential 1 "$input" "$sequential_dir"

  info "[$((index + 1))/${#record_counts[@]}] Processando com $workers workers ($records clubes)..."
  run_batch "workers-${workers}" "$workers" "$input" "$parallel_dir"

  if ! cmp -s "$sequential_dir/clubs.csv" "$parallel_dir/clubs.csv" || ! cmp -s "$sequential_dir/players.csv" "$parallel_dir/players.csv"; then
    printf 'erro: os CSVs sequencial e paralelo diferem para %s clubes.\n' "$records" >&2
    exit 1
  fi

  sequential_wall="$(time_value 'Elapsed (wall clock)' "$sequential_dir/time.txt")"
  parallel_wall="$(time_value 'Elapsed (wall clock)' "$parallel_dir/time.txt")"
  sequential_seconds="$(wall_seconds "$sequential_wall")"
  parallel_seconds="$(wall_seconds "$parallel_wall")"
  reduction="$(awk -v sequential="$sequential_seconds" -v parallel="$parallel_seconds" 'BEGIN { printf "%.1f%%", (1 - parallel / sequential) * 100 }')"
  speedup="$(awk -v sequential="$sequential_seconds" -v parallel="$parallel_seconds" 'BEGIN { printf "%.2fx", sequential / parallel }')"
  sequential_rss="$(format_mib "$(time_value 'Maximum resident set size' "$sequential_dir/time.txt")")"
  parallel_rss="$(format_mib "$(time_value 'Maximum resident set size' "$parallel_dir/time.txt")")"
  sequential_cpu="$(time_value 'Percent of CPU' "$sequential_dir/time.txt")"
  parallel_cpu="$(time_value 'Percent of CPU' "$parallel_dir/time.txt")"

  printf '| %s | %s | %s | %s | %s | %s | %s | %s -> %s |\n' \
    "$records" "$sequential_wall" "$parallel_wall" "$reduction" "$speedup" \
    "$sequential_rss" "$parallel_rss" "$sequential_cpu" "$parallel_cpu" \
    >> "$summary_rows"

  {
    printf '## %s Clubes\n\n' "$records"
    printf 'Os CSVs sequencial e paralelo foram identicos byte a byte.\n\n'
    printf '### Sequencial (`-workers 1`)\n\n```text\n'
    awk '1' "$sequential_dir/time.txt"
    printf '```\n\n### Paralelo (`-workers %d`)\n\n```text\n' "$workers"
    awk '1' "$parallel_dir/time.txt"
    printf '```\n\n'
  } >> "$details"
done

{
  printf '# Benchmark de Workers do Batch\n\n'
  printf 'Compara o processamento sequencial com `%d` workers usando a mesma entrada gerada.\n\n' "$workers"
  printf 'O script interrompe a execucao se qualquer CSV paralelo diferir do sequencial. '
  printf 'O GNU time e necessario apenas para este benchmark; o executavel batch nao depende dele.\n\n'
  printf '| Clubes | Wall sequencial | Wall com %d workers | Reducao de wall time | Speedup | RSS sequencial | RSS com %d workers | CPU sequencial -> %d workers |\n' "$workers" "$workers" "$workers"
  printf '| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |\n'
  awk '1' "$summary_rows"
  printf '\n'
  awk '1' "$details"
} > "$report"

info "Benchmark concluido. Relatorio: $report"
