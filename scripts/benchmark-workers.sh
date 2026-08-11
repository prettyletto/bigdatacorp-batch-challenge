#!/usr/bin/env bash
set -euo pipefail

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
work_dir="$root_dir/.local/benchmark-workers"
batch_bin="$work_dir/batch"
generate_bin="$work_dir/generate"
input="$work_dir/10m-2-players.jsonl"
report="$work_dir/report.md"
summary_rows="$work_dir/summary-rows.md"
records=10000000
players=2
worker_counts=(1 2 4 8 16)

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
  local workers="$1"
  local run_dir="$work_dir/workers-${workers}"

  mkdir -p "$run_dir"

  (
    cd "$run_dir"
    "$time_bin" -v -o time.txt "$batch_bin" -workers "$workers" "$input" > stats.txt
  )
}

mkdir -p "$work_dir"
info "Compilando os binarios locais..."
go build -o "$batch_bin" "$root_dir/cmd/batch"
go build -o "$generate_bin" "$root_dir/cmd/generate"
info "Usando GNU time: $time_bin"

info "Gerando uma entrada com $records clubes e $players jogadores por clube..."
"$generate_bin" -records "$records" -players "$players" -output "$input" > /dev/null

: > "$summary_rows"
baseline_dir=""
previous_seconds=""
plateau_workers=""

for workers in "${worker_counts[@]}"; do
  info "Processando com $workers worker(s)..."
  run_batch "$workers"

  run_dir="$work_dir/workers-${workers}"
  if [[ -z "$baseline_dir" ]]; then
    baseline_dir="$run_dir"
  elif ! cmp -s "$baseline_dir/clubs.csv" "$run_dir/clubs.csv" || ! cmp -s "$baseline_dir/players.csv" "$run_dir/players.csv"; then
    printf 'erro: os CSVs com %s workers diferem da execução com 1 worker.\n' "$workers" >&2
    exit 1
  fi

  elapsed="$(time_value 'Elapsed (wall clock)' "$run_dir/time.txt")"
  elapsed_seconds="$(wall_seconds "$elapsed")"
  rss="$(format_mib "$(time_value 'Maximum resident set size' "$run_dir/time.txt")")"

  if [[ -z "$previous_seconds" ]]; then
    gain="base"
  elif awk -v current="$elapsed_seconds" -v previous="$previous_seconds" 'BEGIN { exit !(current < previous) }'; then
    gain="$(awk -v current="$elapsed_seconds" -v previous="$previous_seconds" 'BEGIN { printf "%.1f%% mais rápido", (1 - current / previous) * 100 }')"
  else
    gain="sem ganho"
    if [[ -z "$plateau_workers" ]]; then
      plateau_workers="$workers"
    fi
  fi

  printf '| %s | %s | %s | %s |\n' "$workers" "$elapsed" "$rss" "$gain" >> "$summary_rows"
  previous_seconds="$elapsed_seconds"
done

{
  printf '# Benchmark de Workers\n\n'
  printf 'Gerado por `scripts/benchmark-workers.sh` com uma única entrada de %s clubes e %s jogadores por clube.\n\n' "$records" "$players"
  printf 'Cada execução é comparada byte a byte com a saída de 1 worker. O GNU `time -v` é necessário apenas para o benchmark; o executável `batch` não depende dele.\n\n'
  printf '| Workers | Tempo decorrido | Memória máxima (RSS) | Ganho em relação ao anterior |\n'
  printf '| ---: | ---: | ---: | ---: |\n'
  awk '1' "$summary_rows"
  printf '\n'
  if [[ -n "$plateau_workers" ]]; then
    printf 'O primeiro aumento sem benefício de tempo foi em `%s` workers.\n' "$plateau_workers"
  else
    printf 'Todos os aumentos de workers reduziram o tempo nesta execução.\n'
  fi
} > "$report"

info "Benchmark concluído. Relatório: $report"
