# Benchmark de Workers

Este relatório foi gerado por `scripts/benchmark-workers.sh`. O script cria a mesma entrada para os dois modos, compara `clubs.csv` e `players.csv` byte a byte e interrompe a execução se houver diferença.

Os números dependem da máquina, do sistema de arquivos e da carga do sistema. O GNU `time -v` é necessário apenas para executar o benchmark; o programa `batch` não depende dele.

| Clubes | Wall sequencial | Wall com 8 workers | Redução de wall time | Speedup | RSS sequencial | RSS com 8 workers | CPU sequencial -> 8 workers |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 100000 | 0:01.93 | 0:00.58 | 69.9% | 3.33x | 10.9 MiB | 22.9 MiB | 102% -> 453% |
| 1000000 | 0:10.02 | 0:03.23 | 67.8% | 3.10x | 10.8 MiB | 21.7 MiB | 102% -> 493% |
| 10000000 | 1:49.43 | 0:39.75 | 63.7% | 2.75x | 10.5 MiB | 21.1 MiB | 103% -> 443% |

O relatório completo, incluindo a saída bruta do GNU `time -v`, é produzido em `.local/benchmark-workers/report.md` a cada execução do script.
