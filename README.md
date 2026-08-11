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

| Opção | Descrição |
| --- | --- |
| `-workers N` | Quantidade de workers. Padrão: `1`. |
| `-maxsize N` | Tamanho máximo do batch em MiB. Padrão: `32`. |
| `-v` | Exibe progresso e métricas de processamento em `stderr`. |

```bash
go run ./cmd/batch -workers 8 -maxsize 16 sample_clubes.jsonl
```

Os arquivos `clubs.csv` e `players.csv` são gravados no diretório atual. A escrita de cada CSV é atômica: em caso de erro durante o processamento, o arquivo final anterior é preservado.

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

Execute toda a suíte com saída detalhada:

```bash
go test -v ./...
```

A suíte cobre diretamente as principais regras do enunciado:

| Regra                         | Cobertura                                                                                                | Testes                                                                                            |
| ----------------------------- | -------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------- |
| **Filtro por campeonato**     | Apenas clubes da `SERIE A` e `SERIE B` são escritos nos CSVs.                                            | `TestProcessWritesPlayersOnlyForEligibleClubs`                                                    |
| **Ligação clube → jogadores** | Mantém clubes sem jogadores, associa jogadores ao clube correto e ignora jogadores sem `club_id` válido. | `TestProcessClubsWithoutPlayersFixture`<br>`TestProcessSkipsPlayersWithoutClubID/club_ID_ausente` |
| **Cores**                     | Concatena múltiplas cores com `\|` e mantém o campo vazio quando não informado.                          | `TestClubRow`<br>`TestProcessClubsWithoutPlayersFixture`                                          |
| **Datas**                     | Preserva datas válidas e transforma datas inválidas ou ausentes em campo vazio.                          | `TestNormalizeDate/valida`<br>`TestNormalizeDate/dia_invalido`                                    |
| **Campos ausentes**           | Campos nulos ou inexistentes são escritos como valores vazios no CSV.                                    | `TestProcessMalformedAndMissingFieldsFixture`                                                     |
| **Formato CSV**               | Valida cabeçalhos, UTF-8 e escaping correto de vírgulas, aspas e quebras de linha.                       | `TestProcessWritesHeadersForEmptyInput`<br>`TestProcessCSVEscapingFixture`                        |
| **JSON malformado**           | Linhas inválidas são ignoradas sem interromper o processamento das próximas.                             | `TestProcessMalformedAndMissingFieldsFixture`                                                     |
| **Concorrência e ordem**      | A saída permanece determinística com um ou vários workers, inclusive entre diferentes batches.           | `TestProcessWithOneAndManyWorkersMatchAcrossBatches`                                              |

Os testes usam conjuntos pequenos de **dados de teste** para validar o conteúdo final dos CSVs, enquanto os testes unitários cobrem as transformações isoladamente.

## Gerador para Testes

O utilitário abaixo cria uma entrada JSONL determinística para testes locais. Ele alterna entre `SERIE A`, `SERIE B` e `SERIE C`.

```bash
go run ./cmd/generate -records 1000000 -players 2 -output .local/1m.jsonl
```

## Benchmark

O script `scripts/benchmark-workers.sh` gera uma única entrada de 10 milhões de clubes, com 2 jogadores por clube, e executa o batch com `1`, `2`, `4`, `8` e `16` workers. Cada saída é comparada byte a byte com a execução de 1 worker.

```bash
scripts/benchmark-workers.sh
```

O benchmark requer o GNU `time` com suporte à opção `-v`; o executável `batch` não depende dessa ferramenta. O relatório local é escrito em `.local/benchmark-workers/report.md` e contém apenas os dados relevantes:

| Coluna | Leitura |
| --- | --- |
| Workers | Quantidade de workers usada na execução. |
| Tempo decorrido | Wall time medido pelo GNU `time`. |
| Memória máxima (RSS) | Pico de memória residente da execução. |
| Ganho em relação ao anterior | Redução real de tempo em comparação com a linha anterior; `sem ganho` indica o primeiro ponto em que aumentar workers não ajudou. |

Os resultados abaixo foram gerados por `scripts/benchmark-workers.sh`, com 10 milhões de clubes e 2 jogadores por clube, em um AMD Ryzen 7 5825U with Radeon Graphics, com 8 núcleos e 16 threads:

| Workers | Tempo decorrido | Memória máxima (RSS) | Ganho em relação ao anterior |
| ---: | ---: | ---: | ---: |
| 1 | 2:15.33 | 9.5 MiB | base |
| 2 | 1:21.74 | 10.1 MiB | 39.6% mais rápido |
| 4 | 0:58.96 | 12.2 MiB | 27.9% mais rápido |
| 8 | 0:53.08 | 19.9 MiB | 10.0% mais rápido |
| 16 | 0:49.44 | 37.3 MiB | 6.9% mais rápido |

Nesta execução, todos os aumentos de workers reduziram o tempo. Cada resultado foi validado byte a byte contra a saída com 1 worker.


## Uso de IA

O diretório `AI/` contém o registro da sessão de IA utilizada durante o desenvolvimento do desafio, disponibilizado em dois formatos:

* `AI/sessao_json_raw.json`: transcrição bruta da sessão em JSON.
* `AI/sessao_readable_para_humanos.md`: versão da mesma sessão formatada para leitura humana.

Esses arquivos foram incluídos para manter transparência sobre o uso de ferramentas de IA durante a implementação.
