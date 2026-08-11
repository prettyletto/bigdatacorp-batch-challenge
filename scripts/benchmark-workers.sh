#!/usr/bin/env bash
set -euo pipefail

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
work_dir="$root_dir/.local/benchmark-workers"
batch_bin="$work_dir/batch"
generate_bin="$work_dir/generate"
report="$work_dir/report.md"
workers=8
record_counts=(100000 1000000 10000000)

mkdir -p "$work_dir"

go build -o "$batch_bin" "$root_dir/cmd/batch"
go build -o "$generate_bin" "$root_dir/cmd/generate"

time_value() {
  local pattern="$1"
  local timing_file="$2"

  awk -v pattern="$pattern" '$0 ~ pattern { print $NF; exit }' "$timing_file"
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
  local records="$1"
  local label="$2"
  local worker_count="$3"
  local input="$4"
  local run_dir="$work_dir/${records}-${label}"

  mkdir -p "$run_dir"

  (
    cd "$run_dir"
    /usr/bin/time -v -o time.txt "$batch_bin" -workers "$worker_count" "$input" > stats.txt
  )
}

{
  printf '# Batch Worker Benchmark\n\n'
  printf 'Runs sequential processing against `%d` workers using the same generated input. '
  printf 'The script stops if either generated CSV differs byte-for-byte.\n\n'
  printf '| Records | Sequential wall | %d-worker wall | Wall-time reduction | Speedup | Sequential RSS | %d-worker RSS | CPU sequential -> %d workers |\n' "$workers" "$workers" "$workers"
  printf '| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |\n'

  for records in "${record_counts[@]}"; do
    input="$work_dir/${records}.jsonl"
    "$generate_bin" -records "$records" -output "$input" > /dev/null

    run_batch "$records" sequential 1 "$input"
    run_batch "$records" "workers-${workers}" "$workers" "$input"

    sequential_dir="$work_dir/${records}-sequential"
    parallel_dir="$work_dir/${records}-workers-${workers}"

    if ! cmp -s "$sequential_dir/clubs.csv" "$parallel_dir/clubs.csv" || ! cmp -s "$sequential_dir/players.csv" "$parallel_dir/players.csv"; then
      printf 'CSV output mismatch for %s records\n' "$records" >&2
      exit 1
    fi

    sequential_wall="$(time_value 'Elapsed \(wall clock\)' "$sequential_dir/time.txt")"
    parallel_wall="$(time_value 'Elapsed \(wall clock\)' "$parallel_dir/time.txt")"
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
      "$sequential_rss" "$parallel_rss" "$sequential_cpu" "$parallel_cpu"

    printf '\n## %s Records\n\n' "$records"
    printf 'CSV outputs matched byte-for-byte.\n\n'
    printf '### Sequential (`-workers 1`)\n\n```text\n'
    awk '1' "$sequential_dir/time.txt"
    printf '```\n\n### Parallel (`-workers %d`)\n\n```text\n' "$workers"
    awk '1' "$parallel_dir/time.txt"
    printf '```\n\n'
  done
} > "$report"

printf 'Benchmark report written to %s\n' "$report"
