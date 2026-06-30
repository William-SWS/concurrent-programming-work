# AGENTS.md

## Project

Bar dos Filósofos — simulação Go do _Drinking Philosophers Problem_ (Chandy & Misra 1984). 4 soluções concorrentes, cada uma em pacote separado.

## Entrypoint & running

- `cmd/runner/main.go` — CLI com flags `-solucao`, `-grafo`, `-rodadas`
- `make run SOL=arbitro GRAFO=data/caso2_bar_6.txt RODADAS=6`
- `make all` → roda as 4 soluções nos 3 grafos → salva em `results/<caso>/<solucao>.txt`
- `make compare` → lê `results/` e imprime tabela comparativa (requer `make all` primeiro)
- `make race` → `go test -race ./...` (detecta data races + deadlock via timeout)

## Cases (data files)

| Caso | Arquivo | Nós | Rodadas |
|------|---------|-----|---------|
| 1 | `data/caso1_jantar_5.txt` | 5 | 6 |
| 2 | `data/caso2_bar_6.txt` | 6 | 6 |
| 3 | `data/caso3_bar_12.txt` | 12 | 3 |

Todas as soluções usam estes mesmos 3 arquivos; round counts são fixos por caso.

## Implementation status

| Solução | Pacote | Status |
|---------|--------|--------|
| Ordenação de Recursos | `solucoes/ordenacao/` | Implementada |
| Árbitro (Garçom) | `solucoes/arbitro/` | Implementada, mas teste `TestSemDeadlock` tem `t.Skip` **stale** (remover o Skip) |
| Chandy-Misra | `solucoes/chandy_misra/` | Implementada |
| Randomized Backoff | `solucoes/backoff/` | Implementada |

## Architecture

- `core.Solver` interface: `Name() string` + `Run(g *Graph, rounds int) []*Philosopher`
- `core.Graph`: matriz de adjacência `[][]bool`, parsed pelo `LoadGraph()`
- `core.Philosopher`: ID, Degree, Metrics. Chamar `Record(state, duration)` para acumular tempo
- `core.Metrics`: acumula `Tranquilo`, `ComSede`, `Bebendo` + contagem `Drinks`
- `core.Report()` gera saída TSV estável — `scripts/compare_results.py` faz o parsing pelas linhas `# tempo_total=` e colunas `id\tgrau\t...`
- Cada pacote em `solucoes/` é independente, sem dependência entre si (arquitetura para desenvolvimento paralelo)

## Test quirks

- Os testes usam paths relativos para os dados, cada um do seu pacote: `../../data/caso1_jantar_5.txt` (nos 3 pacotes de solução) ou `../data/...` (no `core/`)
- **Chandy-Misra**: o teste reduz `Unidade` (variável de pacote, padrão `time.Second`) para `2ms` para execução rápida; isso é essencial para o teste não estourar timeout
- Chandy-Misra testa rounds **diferentes** por caso (4, 4, 3), não os mesmos 6/6/3 dos casos reais
- Backoff testa só caso 1 com 2 rounds
- `make race` testa tudo com `-count=20` implícito? **Não** — `make race` executa `go test -race ./...` uma vez. `go test -race -count=20 ./solucoes/chandy_misra/` é o comando manual para estressar

## Output format

TSV com cabeçalho `#`:
```
# solucao=Árbitro (Garçom)
# tempo_total=18.01s
id\tgrau\ttranquilo\tcom_sede\tbebendo\tbebidas
0\t2\t7.00\t2.00\t6.00\t6
...
# espera media (com sede) por grau:
# grau=2 media=2.80s (n=5)
```

## Verifying correctness

- Ausência de deadlock: simulacao conclui (ou teste não estoura timeout de 30s)
- Ausência de data race: `go test -race ./...` passa
- Fairness: esperas médias equilibradas entre filósofos de **mesmo grau** (verificar a saída `# espera media por grau:`)
