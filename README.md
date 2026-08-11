# Desafio Batch BigDataCorp

Lê um arquivo JSONL de clubes de futebol e gera dois CSVs em UTF-8:

- `clubs.csv`: um registro por clube elegível.
- `players.csv`: um registro por jogador dos clubes elegíveis.

São incluídos apenas clubes de `SERIE A` e `SERIE B`. Campos ausentes ou nulos ficam vazios, datas inválidas ficam vazias e registros JSON malformados são ignorados sem interromper o processamento.

## Como executar

É necessário ter o Go `1.26.5` ou compatível instalado.

```bash
go run ./cmd/batch sample_clubes.jsonl
```

Para processar em paralelo, informe a quantidade de workers. O padrão é `1`. A opção `-maxsize` controla o tamanho máximo de cada batch em memória, em MiB; o padrão é `32`.

```bash
go run ./cmd/batch -workers 8 -maxsize 16 sample_clubes.jsonl
```

Um registro individual maior que o limite continua sendo processado normalmente. O limite controla o acúmulo de registros por batch.

Use `-v` para acompanhar o processamento em `stderr`. Em arquivos com 100 kB ou mais, o progresso é exibido a cada 100 mil registros. Em arquivos menores, ele é exibido a cada 10% dos bytes processados. Cada linha reúne o total processado, o total de JSONs malformados e o total de clubes ignorados.

```bash
go run ./cmd/batch -v -workers 8 -maxsize 16 sample_clubes.jsonl 2> processamento.log
```

Os arquivos `clubs.csv` e `players.csv` são gravados no diretório atual. A escrita de cada CSV é atômica: em caso de erro durante o processamento, o arquivo final anterior é preservado.

Exemplo com binário compilado:

```bash
go build -o batch ./cmd/batch
./batch -workers 8 -maxsize 32 caminho/para/clubes.jsonl
```

## Saída

```text
clubs.csv
Id do Clube,Nome,Campeonato,...
SCCP,Sport Club Corinthians Paulista,SERIE A,...

players.csv
Id do Clube,Id do Jogador,Nome,...
SCCP,SCCP-10,Rodrigo Garro,...
```

Os valores CSV são escapados pelo padrão RFC 4180 quando contêm vírgulas, aspas ou quebras de linha.

## Testes

```bash
go test ./...
go test -race ./...
```

## Gerador de carga

O utilitário abaixo cria uma entrada JSONL determinística para testes locais. Ele alterna entre `SERIE A`, `SERIE B` e `SERIE C`.

```bash
go run ./cmd/generate -records 1000000 -players 2 -output .local/1m.jsonl
```

## Benchmark

O script `scripts/benchmark-workers.sh` compara o processamento sequencial com `8` workers para 100 mil, 1 milhão e 10 milhões de clubes. Ele interrompe a execução se os CSVs paralelo e sequencial forem diferentes.

```bash
scripts/benchmark-workers.sh
```

O benchmark requer o GNU `time` com suporte à opção `-v`; o executável `batch` não depende dessa ferramenta. O relatório gerado localmente fica em `.local/benchmark-workers/report.md`.

Resultados obtidos nesta máquina:

| Clubes | Wall sequencial | Wall com 8 workers | Redução de wall time | Speedup | RSS sequencial | RSS com 8 workers | CPU sequencial -> 8 workers |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 100k | 0:01.93 | 0:00.58 | 69.9% | 3.33x | 10.9 MiB | 22.9 MiB | 102% -> 453% |
| 1M | 0:10.02 | 0:03.23 | 67.8% | 3.10x | 10.8 MiB | 21.7 MiB | 102% -> 493% |
| 10M | 1:49.43 | 0:39.75 | 63.7% | 2.75x | 10.5 MiB | 21.1 MiB | 103% -> 443% |

O [relatório completo](assets/benchmark-workers.md) preserva o contexto e indica como reproduzir a medição.


## Uso de IA

O diretório `AI/` contém o registro da sessão de IA utilizada durante o desenvolvimento do desafio, disponibilizado em dois formatos:

* `AI/sessao_json_raw.json`: transcrição bruta da sessão em JSON.
* `AI/sessao_readable_para_humanos.md`: versão da mesma sessão formatada para leitura humana.

Esses arquivos foram incluídos para manter transparência sobre o uso de ferramentas de IA durante a implementação.
