#!/usr/bin/env bash
# Roda todas as soluções em todos os grafos e salva os relatórios em results/.
set -euo pipefail

cd "$(dirname "$0")/.."

solucoes=(ordenacao arbitro chandy_misra backoff)

# rodadas por caso (casos 1 e 2 = 6 bebidas; caso 3 = 3 bebidas)
declare -A grafo_caso=(
  [caso1]="data/caso1_jantar_5.txt"
  [caso2]="data/caso2_bar_6.txt"
  [caso3]="data/caso3_bar_12.txt"
)
declare -A rodadas_caso=(
  [caso1]=6
  [caso2]=6
  [caso3]=3
)

for caso in caso1 caso2 caso3; do
  mkdir -p "results/$caso"
  for sol in "${solucoes[@]}"; do
    echo ">> $sol / $caso"
    go run ./cmd/runner \
      -solucao="$sol" \
      -grafo="${grafo_caso[$caso]}" \
      -rodadas="${rodadas_caso[$caso]}" \
      > "results/$caso/${sol}.txt"
  done
done

echo "Pronto. Relatórios em results/."
