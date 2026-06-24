.PHONY: build test race run all compare clean

build:
	go build ./...

test:
	go test ./...

# Roda os testes com o detector de race (evidência de ausência de data races).
race:
	go test -race ./...

# Ex.: make run SOL=ordenacao GRAFO=data/caso2_bar_6.txt RODADAS=6
SOL ?= ordenacao
GRAFO ?= data/caso1_jantar_5.txt
RODADAS ?= 6
run:
	go run ./cmd/runner -solucao=$(SOL) -grafo=$(GRAFO) -rodadas=$(RODADAS)

# Roda todas as soluções em todos os grafos e salva em results/.
all:
	bash scripts/run_all.sh

# Tabela comparativa a partir de results/.
compare:
	python3 scripts/compare_results.py

clean:
	rm -rf results/caso*/
