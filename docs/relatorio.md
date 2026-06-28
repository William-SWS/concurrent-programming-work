# Relatório — Bar dos Filósofos

> Template. Preencher com os números gerados por `make all` + `make compare`.

## 1. Soluções implementadas

| # | Solução | Responsável | Estratégia anti-deadlock |
|---|---------|-------------|--------------------------|
| 1 | Ordenação de Recursos | | adquire garrafas em ordem crescente de número |
| 2 | Árbitro (Garçom) | | no máx. N-1 filósofos ativos |
| 3 | Chandy-Misra | | garrafas com estado, troca de mensagens |
| 4 | Randomized Backoff | | solta tudo e espera tempo aleatório |

## 2. Resultados

Tabela por caso (preencher com `results/`):

### Caso 1 — Jantar clássico (5 nós, 6 bebidas)

| Solução | Tempo total (s) | Espera média (s) |
|---------|-----------------|------------------|
| ... | | |

### Caso 2 — Bar baixa conectividade (6 nós, 6 bebidas)

| Solução | Tempo total (s) | Espera média (s) |
|---------|-----------------|------------------|
| ... | | |

### Caso 3 — Bar alta conectividade (12 nós, 3 bebidas)

| Solução | Tempo total (s) | Espera média (s) |
|---------|-----------------|------------------|
| ... | | |

## 3. Análise

- **Deadlock:** nenhuma solução pode travar. Evidência: `make race` passa sem
  warnings de race e as simulações concluem.
- **Starvation:** comparar a espera média **entre filósofos de mesmo grau** — o
  enunciado exige que seja equilibrada. Comentar discrepâncias.
- **Comparação:** qual solução teve menor tempo total? Menor espera? Trade-offs
  (ex.: árbitro é simples mas serializa demais; Chandy-Misra escala melhor mas
  é mais complexo).

## 4. Como reproduzir

```sh
make all       # roda as 4 soluções nos 3 grafos -> results/
make compare   # tabela resumo
make race      # valida ausência de data races
```
