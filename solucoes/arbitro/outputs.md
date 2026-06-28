# Solução 2 — Árbitro (Garçom) — Resultados

## Caso 1: Jantar de 5 nós, 6 rodadas

```
go run ./cmd/runner -solucao=arbitro -grafo=data/caso1_jantar_5.txt -rodadas=6
```

| Parâmetro | Valor |
|-----------|-------|
| Tempo total | 18.01s |

| Filósofo | Grau | Tranquilo (s) | Com Sede (s) | Bebendo (s) | Bebidas |
|----------|------|---------------|--------------|-------------|---------|
| 0 | 2 | 7.00 | 2.00 | 6.00 | 6 |
| 1 | 2 | 6.00 | 4.00 | 6.00 | 6 |
| 2 | 2 | 10.00 | 2.00 | 6.00 | 6 |
| 3 | 2 | 6.00 | 4.00 | 6.00 | 6 |
| 4 | 2 | 9.00 | 2.00 | 6.00 | 6 |

**Espera média por grau:**

| Grau | Média (s) | n |
|------|-----------|---|
| 2 | 2.80 | 5 |

---

## Caso 2: 6 nós, conectividade baixa, 6 rodadas

```
go run ./cmd/runner -solucao=arbitro -grafo=data/caso2_bar_6.txt -rodadas=6
```

| Parâmetro | Valor |
|-----------|-------|
| Tempo total | 28.02s |

| Filósofo | Grau | Tranquilo (s) | Com Sede (s) | Bebendo (s) | Bebidas |
|----------|------|---------------|--------------|-------------|---------|
| 0 | 2 | 5.00 | 4.00 | 6.00 | 6 |
| 1 | 2 | 5.00 | 1.00 | 6.00 | 6 |
| 2 | 3 | 12.00 | 5.00 | 6.00 | 6 |
| 3 | 3 | 9.00 | 3.00 | 6.00 | 6 |
| 4 | 2 | 3.00 | 0.00 | 6.00 | 6 |
| 5 | 4 | 15.00 | 7.01 | 6.00 | 6 |

**Espera média por grau:**

| Grau | Média (s) | n |
|------|-----------|---|
| 2 | 1.67 | 3 |
| 3 | 4.00 | 2 |
| 4 | 7.01 | 1 |

---

## Caso 3: 12 nós, conectividade alta, 3 rodadas

```
go run ./cmd/runner -solucao=arbitro -grafo=data/caso3_bar_12.txt -rodadas=3
```

| Parâmetro | Valor |
|-----------|-------|
| Tempo total | 19.01s |

| Filósofo | Grau | Tranquilo (s) | Com Sede (s) | Bebendo (s) | Bebidas |
|----------|------|---------------|--------------|-------------|---------|
| 0 | 3 | 5.00 | 1.00 | 3.00 | 3 |
| 1 | 3 | 3.00 | 0.00 | 3.00 | 3 |
| 2 | 2 | 1.00 | 1.00 | 3.00 | 3 |
| 3 | 4 | 11.00 | 2.00 | 3.00 | 3 |
| 4 | 4 | 10.00 | 2.00 | 3.00 | 3 |
| 5 | 3 | 1.00 | 0.00 | 3.00 | 3 |
| 6 | 5 | 14.00 | 2.00 | 3.00 | 3 |
| 7 | 4 | 3.00 | 3.00 | 3.00 | 3 |
| 8 | 3 | 5.00 | 0.00 | 3.00 | 3 |
| 9 | 6 | 14.00 | 1.00 | 3.00 | 3 |
| 10 | 2 | 4.00 | 3.00 | 3.00 | 3 |
| 11 | 3 | 4.00 | 1.00 | 3.00 | 3 |

**Espera média por grau:**

| Grau | Média (s) | n |
|------|-----------|---|
| 2 | 2.00 | 2 |
| 3 | 0.40 | 5 |
| 4 | 2.34 | 3 |
| 5 | 2.00 | 1 |
| 6 | 1.00 | 1 |
