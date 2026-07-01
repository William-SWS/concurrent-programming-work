# Relatório — Bar dos Filósofos: Simulação de Alocação Concorrente de Recursos

**Disciplina:** Programação Concorrente

**Equipe:**
- Samuel William Silva Almeida
- Davi de Oliveira
- Antonio Mozar Braga
- Antonio Bezerra de Morais Neto

---

## 1. Introdução

O problema dos filósofos famintos (_Dining Philosophers Problem_), proposto por Dijkstra em 1971, é um clássico da computação concorrente que ilustra os desafios de sincronização entre processos que compartilham recursos limitados. Neste trabalho, abordamos uma generalização desse problema conhecida como **Bar dos Filósofos** (_Drinking Philosophers Problem_), formalizada por Chandy e Misra em 1984.

Diferentemente do problema clássico, onde os filósofos estão dispostos em um círculo e cada um precisa de exatamente dois garfos para comer, no Bar dos Filósofos temos as seguintes generalizações:

- **Topologia arbitrária**: os filósofos são vértices de um grafo qualquer, e as garrafas (recursos compartilhados) são as arestas.
- **Demanda variável**: cada filósofo, ao ficar com sede, sorteia aleatoriamente um subconjunto de 2 a *n* garrafas para solicitar, onde *n* é o grau do vértice (número de vizinhos).
- **Seleção dinâmica**: a cada rodada, o subconjunto de garrafas desejado pode ser diferente.

Cada filósofo percorre ciclicamente três estados: **tranquilo** (pensando por um tempo aleatório de 0 a *n* segundos), **com sede** (tentando adquirir todas as garrafas do subconjunto sorteado) e **bebendo** (seção crítica de 1 segundo). O tempo gasto no estado "com sede" representa o tempo de espera e é a principal métrica para avaliar *fairness* entre os filósofos.

O objetivo deste projeto foi implementar e comparar quatro soluções de sincronização para este problema, cada uma utilizando uma estratégia diferente para garantir ausência de deadlock. As soluções foram testadas em três cenários (grafos) de complexidade crescente, variando o número de vértices e a conectividade.

---

## 2. Estrutura do Projeto

O projeto foi estruturado em Go 1.24, com um pacote compartilhado `core/` contendo as definições comuns e quatro pacotes independentes em `solucoes/`, cada um implementando uma solução. A arquitetura permitiu o desenvolvimento paralelo sem conflitos.

### 2.1 Pacote Core (`core/`)

| Arquivo | Responsabilidade |
|---------|-----------------|
| `graph.go` | Estrutura `Graph` (matriz de adjacência), `LoadGraph()`, `Neighbors()`, `Degree()`, `Edges()` |
| `states.go` | Enum `State`: `Tranquilo`, `ComSede`, `Bebendo` |
| `philosopher.go` | Struct `Philosopher` (ID, Degree, Metrics), método `Record(state, duration)` |
| `metrics.go` | Struct `Metrics`, função `Report()` para saída TSV |
| `analyze.go` | Funções de análise de starvation (outliers por grau, cv, max/mean) |
| `solver.go` | Interface `Solver: Name() + Run(g *Graph, rounds int) []*Philosopher` |

### 2.2 Interface Central

```go
type Solver interface {
    Name() string
    Run(g *Graph, rounds int) []*Philosopher
}
```

### 2.3 Casos de Teste

| Caso | Arquivo | Descrição | Nós | Topologia | Grau máximo | Rodadas |
|------|---------|-----------|-----|-----------|-------------|---------|
| 1 | `data/caso1_jantar_5.txt` | Jantar clássico | 5 | Ciclo | 2 | 6 |
| 2 | `data/caso2_bar_6.txt` | Bar conectividade baixa | 6 | Esparsa | 4 | 6 |
| 3 | `data/caso3_bar_12.txt` | Bar conectividade alta | 12 | Densa | 6 | 3 |

### 2.4 Pipeline de Execução

```sh
# Execução única
go run ./cmd/runner -solucao=<nome> -grafo=<arquivo> -rodadas=<N>

# Todas as soluções em todos os grafos
make all          # → results/caso<N>/<solucao>.txt

# Tabela comparativa
make compare      # → python3 scripts/compare_results.py

# Validação com detector de data races
make race         # → go test -race ./...
```

---

## 3. Soluções Implementadas

### 3.1 Ordenação de Recursos

**Responsável:** Antonio Bezerra de Morais Neto  
**Pacote:** `solucoes/ordenacao/` (108 linhas)

#### Algoritmo

Cada aresta (garrafa) recebe um número de ordem global único. Os filósofos adquirem as garrafas **estritamente em ordem crescente** dessa numeração, eliminando a condição de espera circular (_circular wait_) — uma das quatro condições necessárias para o deadlock (Coffman, Elphick & Shoshani, 1971).

```go
chosen := chooseNeighbors(r, neighbors)       # subconjunto aleatório
sort.Slice(chosen, func(i, j int) bool {       # ordena pela numeração global
    return order[lockOrder(p.ID, chosen[i])] < order[lockOrder(p.ID, chosen[j])]
})
for _, nb := range chosen {
    bottles[lockOrder(p.ID, nb)].Lock()
}
```

#### Prova de Correção

Se todos os processos adquirem recursos na mesma ordem global, o grafo de alocação nunca pode conter ciclos. Suponha, por absurdo, um ciclo P₁ → R₁ → P₂ → R₂ → ... → Pₖ → Rₖ → P₁. Pela regra de aquisição, se Pⱼ aguarda Rⱼ, então `num(Rⱼ) > num(Rⱼ₋₁)`. Percorrendo o ciclo: `num(R₁) > num(Rₖ) > ... > num(R₁)` — contradição. Logo, deadlock é impossível.

#### Complexidade

| Cenário | Complexidade | Justificativa |
|---------|-------------|---------------|
| Tempo (melhor caso) | Ω(N × D log D × R) | Sem contenção: shuffle + sort + locks |
| Tempo (pior caso) | O(N × D log D × R) | Sem fator quadrático (diferente do Árbitro) |
| Espaço | O(N²) | Matriz de adjacência + mapas de mutex e ordem |

#### Armadilha: ordenação antes da aquisição

A correção depende de ordenar o subconjunto **antes** do loop de locks. Se o `sort.Slice` fosse omitido, a solução degeneraria para locks sem ordenação — que **não** garante deadlock freedom. O código separa explicitamente as etapas de escolha, ordenação e aquisição.

#### Padrão de espera

A ordenação global **inverte** a correlação grau-espera observada nas outras soluções: filósofos de maior grau podem ter espera **menor** que os de menor grau, porque a numeração global nivela a competição de forma distinta.

---

### 3.2 Árbitro (Garçom)

**Responsável:** Samuel William Silva Almeida  
**Pacote:** `solucoes/arbitro/` (171 linhas)

#### Algoritmo

Um coordenador central mantém um semáforo de capacidade **N − 1** implementado como canal Go bufferizado. Cada filósofo, ao ficar com sede, deve obter permissão do árbitro antes de tentar adquirir garrafas. Com N filósofos e no máximo N − 1 ativos, pelo menos um está sempre tranquilo — impossibilitando o deadlock.

```go
type arbiter struct {
    sem chan struct{} // capacidade N-1
}
func (a *arbiter) acquire() { a.sem <- struct{}{} }
func (a *arbiter) release() { <-a.sem }
```

#### Complexidade

| Cenário | Complexidade | Justificativa |
|---------|-------------|---------------|
| Tempo (melhor caso) | Ω(N × D × R) | Baixa contenção |
| Tempo (pior caso) | O(N² × R) | Semáforo serializa N − 1 ativos |
| Espaço | O(N²) | Matriz + mapa de mutexes + canal |

#### Armadilha: RNG global como gargalo oculto

A primeira versão usava `math/rand.Intn()` global, que possui um `sync.Mutex` interno. Com N goroutines chamando `rand` intensamente, esse lock se torna um gargalo de contenção. A solução manteve o RNG global intencionalmente para N ≤ 12, onde o impacto é irrelevante. Soluções com muitas threads devem usar RNG privado por goroutine (`rand.New(rand.NewSource(...))`).

#### Padrão de espera

Apresenta viés contra filósofos de maior grau: no Caso 2, grau 2 esperou 3,34 s enquanto grau 3 esperou 8,51 s. O semáforo serializa o acesso, e filósofos que precisam de mais recursos competem em desvantagem.

---

### 3.3 Chandy-Misra

**Responsável:** Davi de Oliveira  
**Pacote:** `solucoes/chandy_misra/` (~410 linhas)

#### Algoritmo

Solução completamente distribuída por troca de mensagens (canais Go) com duas camadas:

1. **Camada dos garfos** (auxiliar): implementa o grafo de precedência H via "jantar dos filósofos higiênico". Garfos limpos/sujos determinam precedência sobre cada aresta.
2. **Camada das garrafas** (recurso real): cada aresta possui uma garrafa e um _request token_. Pedidos são servidos na ordem de chegada.

A comunicação usa quatro tipos de mensagem: `reqFork`, `sendFork`, `reqBottle`, `sendBottle`. Não há estado compartilhado mutável entre goroutines.

#### Prova de Correção

O Teorema 5 de Chandy & Misra (1984) garante _deadlock-freedom_ e _starvation-freedom_: a distribuição inicial (garfo sujo no filósofo de menor ID) mantém o grafo de precedência H acíclico, e o sistema de mensagens em ordem de chegada (_mailbox_) impede starvation.

#### Complexidade

| Cenário | Complexidade | Justificativa |
|---------|-------------|---------------|
| Mensagens por sessão | O(d) | No máximo 4 mensagens por vizinho |
| Mensagens totais | O(rounds × E) | Por rodada, por aresta |
| Espaço | O(V + E) | Canais + arestas (sem matriz explícita duplicada) |

#### Armadilha: necessidade da camada de garfos

Uma implementação inicial com apenas a camada de garrafas (sem o grafo de precedência H) apresentou deadlock no Caso 3. A camada de garfos é essencial para manter H acíclico quando filósofos sorteiam subconjuntos arbitrários de garrafas — uma sutileza que não aparece no jantar clássico (onde cada filósofo precisa de exatamente 2 recursos).

#### Padrão de espera

Distribuição mais equilibrada entre todas as soluções. No Caso 2, as esperas variaram de 2,00 s (grau 4) a 3,34 s (grau 2) — sem correlação clara com o grau. Apresentou a menor variância do conjunto.

---

### 3.4 Randomized Backoff

**Responsável:** Antonio Mozar Braga  
**Pacote:** `solucoes/backoff/` (142 linhas)

#### Algoritmo

Abordagem probabilística: cada filósofo tenta adquirir as garrafas com `TryLock()` (não bloqueante). Se falhar em qualquer uma, libera as já adquiridas e espera um tempo aleatório (0–100 ms) antes de tentar novamente. O uso de `TryLock` elimina espera bloqueante — ninguém fica preso esperando um mutex ocupado.

```go
for {
    sucesso := true
    for _, v := range vizinhosEscolhidos {
        if m.TryLock() {
            garrafasPegas = append(garrafasPegas, m)
        } else {
            sucesso = false; break
        }
    }
    if sucesso { break }
    for _, m := range garrafasPegas { m.Unlock() }
    time.Sleep(time.Duration(r.Float64() * 0.1 * float64(time.Second)))
}
```

Cada filósofo possui RNG privado, eliminando contenção no gerador global. O tempo de pensamento é **contínuo** (`Float64`), não discreto (`Intn`), reduzindo a probabilidade de dois filósofos acordarem no mesmo instante.

#### Complexidade

| Cenário | Complexidade | Justificativa |
|---------|-------------|---------------|
| Tempo (esperado) | O(N × D × R) | Tentaivas bem-sucedidas na primeira iteração |
| Tempo (pior caso) | **Não limitado** | Livelock teórico possível |
| Espaço | O(N²) | Matriz + mapa de mutexes + RNGs privados |

#### Armadilha: TryLock não é _fair_

O método `TryLock()` do `sync.Mutex` em Go não é _fair_. O escalonamento interno pode favorecer certas goroutines, contribuindo para starvation. O backoff fixo (0–100 ms) mitiga parcialmente, mas a solução não oferece garantias formais. Backoff exponencial (_double on each failure_) seria mais robusto para cenários de alta contenção.

#### Padrão de espera

Exibe o mesmo viés contra graus elevados observado no Árbitro. No Caso 2, grau 4 esperou 7,47 s contra 2,58 s do grau 2. O tempo total médio é o mais baixo do conjunto (21,84 s), graças ao `TryLock` não bloqueante.

---

## 4. Resultados Experimentais

As execuções foram realizadas em ambiente Linux com Go 1.24, utilizando `time.Second` como unidade de tempo. Cada célula foi executada com `make all`, que roda cada uma das 4 soluções × 3 casos. A saída completa inclui a análise de starvation gerada pelo pacote `core/analyze.go`.

### 4.1 Caso 1 — Jantar Clássico (5 nós, grau 2, 6 rodadas)

| Solução | Tempo total (s) | Espera média (s) | Maior espera (s) | Outliers |
|---------|----------------|-----------------|-----------------|----------|
| Ordenação | 21,01 | 4,60 | 6,01 (filo 0) | nenhum |
| Árbitro | 18,01 | 2,80 | 6,00 (filo 1) | filo 1 (214% da média) |
| Chandy-Misra | 18,01 | 3,20 | 4,00 (filo 1,2,4) | nenhum |
| Backoff | 16,99 | 3,58 | 4,69 (filo 2) | nenhum |

**Detalhamento por filósofo (Ordenação):**

| ID | Grau | Tranquilo (s) | Com Sede (s) | Bebendo (s) | Bebidas |
|----|------|---------------|--------------|-------------|---------|
| 0 | 2 | 4,00 | 6,01 | 6,00 | 6 |
| 1 | 2 | 10,00 | 5,01 | 6,00 | 6 |
| 2 | 2 | 3,00 | 5,00 | 6,00 | 6 |
| 3 | 2 | 3,00 | 6,01 | 6,00 | 6 |
| 4 | 2 | 3,00 | 1,00 | 6,00 | 6 |

No grafo cíclico, todos os 5 filósofos possuem grau 2. O Backoff foi o mais rápido (16,99 s) por usar `TryLock` sem espera bloqueante. O Árbitro e o Chandy-Misra empataram em tempo total (18,01 s), mas o Chandy-Misra não apresentou outliers. A Ordenação teve o maior tempo total (21,01 s) e a maior espera média (4,60 s), embora sem outliers.

### 4.2 Caso 2 — Conectividade Baixa (6 nós, graus 2–4, 6 rodadas)

| Solução | Tempo total (s) | Espera média (s) | Grau 2 (n=3) | Grau 3 (n=2) | Grau 4 (n=1) | Outliers |
|---------|----------------|-----------------|-------------|-------------|-------------|----------|
| Ordenação | 27,02 | 5,17 | 6,01 | 3,01 | 7,01 | nenhum |
| Árbitro | 25,02 | 5,34 | 3,34 | 8,51 | 5,01 | nenhum |
| Chandy-Misra | 22,01 | 2,84 | 3,34 | 2,50 | 2,00 | nenhum |
| Backoff | 31,16 | 3,01 | 2,58 | 1,42 | 7,47 | filo 1 (211%) |

**Detalhamento por filósofo (Caso 2):**

**Ordenação:**
| ID | Grau | Tranquilo (s) | Com Sede (s) | Bebendo (s) | Bebidas |
|----|------|---------------|--------------|-------------|---------|
| 0 | 2 | 5,00 | 7,01 | 6,01 | 6 |
| 1 | 2 | 7,00 | 7,01 | 6,00 | 6 |
| 2 | 3 | 5,00 | 3,00 | 6,01 | 6 |
| 3 | 3 | 7,00 | 3,01 | 6,00 | 6 |
| 4 | 2 | 5,00 | 4,00 | 6,01 | 6 |
| 5 | 4 | 14,00 | 7,01 | 6,01 | 6 |

**Árbitro:**
| ID | Grau | Tranquilo (s) | Com Sede (s) | Bebendo (s) | Bebidas |
|----|------|---------------|--------------|-------------|---------|
| 0 | 2 | 5,00 | 1,00 | 6,01 | 6 |
| 1 | 2 | 8,00 | 5,01 | 6,01 | 6 |
| 2 | 3 | 10,01 | 9,01 | 6,01 | 6 |
| 3 | 3 | 5,00 | 8,01 | 6,01 | 6 |
| 4 | 2 | 4,00 | 4,01 | 6,00 | 6 |
| 5 | 4 | 13,00 | 5,01 | 6,00 | 6 |

**Chandy-Misra:**
| ID | Grau | Tranquilo (s) | Com Sede (s) | Bebendo (s) | Bebidas |
|----|------|---------------|--------------|-------------|---------|
| 0 | 2 | 7,00 | 3,00 | 6,00 | 6 |
| 1 | 2 | 8,01 | 3,00 | 6,00 | 6 |
| 2 | 3 | 14,00 | 2,00 | 6,01 | 6 |
| 3 | 3 | 8,00 | 3,00 | 6,01 | 6 |
| 4 | 2 | 5,00 | 4,00 | 6,00 | 6 |
| 5 | 4 | 13,01 | 2,00 | 6,00 | 6 |

**Backoff:**
| ID | Grau | Tranquilo (s) | Com Sede (s) | Bebendo (s) | Bebidas |
|----|------|---------------|--------------|-------------|---------|
| 0 | 2 | 7,68 | 1,19 | 6,00 | 6 |
| 1 | 2 | 7,42 | 5,45 | 6,00 | 6 |
| 2 | 3 | 11,35 | 0,47 | 6,00 | 6 |
| 3 | 3 | 12,74 | 2,37 | 6,00 | 6 |
| 4 | 2 | 2,98 | 1,11 | 6,00 | 6 |
| 5 | 4 | 17,69 | 7,47 | 6,00 | 6 |

O Caso 2 é o mais revelador para comparar as soluções. O Chandy-Misra foi o mais rápido (22,01 s) e o mais equilibrado (esperas de 2,00 a 4,00 s, sem correlação com grau). O Árbitro exibiu o maior viés: grau 3 esperou 8,51 s (média) contra 3,34 s do grau 2. O Backoff foi o mais lento (31,16 s) — a alta contenção do Caso 2 gerou múltiplas tentativas frustradas e backoffs acumulados. A Ordenação apresentou espera média alta (5,17 s) mas sem outliers: a ordenação global distribuiu a contenção de forma diferente, com grau 3 tendo a menor espera (3,01 s).

### 4.3 Caso 3 — Conectividade Alta (12 nós, graus 2–6, 3 rodadas)

| Solução | Tempo total (s) | Espera média (s) | Grau 2 | Grau 3 | Grau 4 | Grau 5 | Grau 6 | Outliers |
|---------|----------------|-----------------|--------|--------|--------|--------|--------|----------|
| Ordenação | 16,01 | 0,75 | 0,50 | 0,80 | 1,33 | 0,00 | 0,00 | filo 4 (225%), filo 5 (250%) |
| Árbitro | 13,01 | 1,17 | 1,00 | 1,60 | 1,00 | 0,00 | 1,00 | nenhum |
| Chandy-Misra | 13,01 | 0,58 | 1,50 | 0,20 | 0,67 | 0,00 | 1,00 | filo 7 (300%), filo 8 (**499%** CRÍTICO) |
| Backoff | 17,36 | 1,17 | 1,07 | 1,27 | 1,38 | 0,29 | 1,08 | nenhum |

**Detalhamento por filósofo (Caso 3):**

**Ordenação:**
| ID | Grau | Espera (s) | ID | Grau | Espera (s) |
|----|------|-----------|----|------|-----------|
| 0 | 3 | 1,00 | 6 | 5 | 0,00 |
| 1 | 3 | 0,00 | 7 | 4 | 0,00 |
| 2 | 2 | 0,00 | 8 | 3 | 0,00 |
| 3 | 4 | 1,00 | 9 | 6 | 0,00 |
| 4 | 4 | 3,00 | 10 | 2 | 1,00 |
| 5 | 3 | 2,00 | 11 | 3 | 1,00 |

**Chandy-Misra:**
| ID | Grau | Espera (s) | ID | Grau | Espera (s) |
|----|------|-----------|----|------|-----------|
| 0 | 3 | 0,00 | 6 | 5 | 0,00 |
| 1 | 3 | 0,00 | 7 | 4 | **2,00** |
| 2 | 2 | 1,00 | 8 | 3 | **1,00** |
| 3 | 4 | 0,00 | 9 | 6 | 1,00 |
| 4 | 4 | 0,00 | 10 | 2 | 2,00 |
| 5 | 3 | 0,00 | 11 | 3 | 0,00 |

O Caso 3, com 12 filósofos e 3 rodadas, apresentou as menores esperas. O Chandy-Misra teve a menor espera média (0,58 s) e o menor tempo total (13,01 s, empatado com o Árbitro), embora tenha gerado o outlier mais severo: filósofo 8 (grau 3) esperou 1,00 s — **499%** da média do seu grupo (0,20 s), classificado como **CRÍTICO** pelo sistema de análise. Este outlier ocorre porque, com apenas 3 rodadas e médias muito baixas, o desvio relativo é amplificado. Em valores absolutos, 1 s de espera para 3 rodadas é aceitável.

A Ordenação se destacou com 6 filósofos com espera zero, incluindo os de maior grau (5 e 6). Os outliers (filo 4 com 3,00 s e filo 5 com 2,00 s) são explicados pelo alinhamento desfavorável entre seus subconjuntos de garrafas e a ordem global.

---

## 5. Comparativo entre as Soluções

### 5.1 Abordagem de Sincronização

| Dimensão | Ordenação de Recursos | Árbitro (Garçom) | Chandy-Misra | Backoff Aleatório |
|----------|----------------------|------------------|--------------|-------------------|
| Paradigma | Memória compartilhada | Memória compartilhada | Passagem de mensagens | Memória compartilhada |
| Coordenação | Distribuída (implícita) | Centralizada | Distribuída (explícita) | Distribuída (probabilística) |
| Garantia de deadlock | Determinística (ordem total) | Determinística (N−1) | Determinística (H acíclico) | Probabilística |
| Starvation freedom | Sim (fairness por ordem) | Não garantido | Sim (Teorema 5) | Não |
| Operação de lock | `Lock` (bloqueante) | `Lock` (bloqueante) | Mensagens (canais) | `TryLock` (não-bloqueante) |
| RNG | Privado por filósofo | Global | Privado por filósofo | Privado por filósofo |
| Linhas de código | ~108 | ~171 | ~410 | ~142 |
| Camadas conceituais | 1 (garrafas) | 1 (árbitro + garrafas) | 2 (garfos + garrafas) | 1 (garrafas) |

### 5.2 Complexidade Assintótica

| Métrica | Ordenação | Árbitro | Chandy-Misra | Backoff |
|---------|-----------|---------|--------------|---------|
| Tempo (melhor caso) | Ω(N × D log D × R) | Ω(N × D × R) | Ω(N × D × R) | Ω(N × D × R) |
| Tempo (pior caso) | O(N × D log D × R) | **O(N² × R)** | O(rounds × E) | **Não limitado** |
| Espaço | O(N²) | O(N²) | O(V + E) | O(N²) |
| Gargalo principal | `sort.Slice` | Semáforo N−1 | Canais de mensagem | Tentativas T |

### 5.3 Resultados Agregados

| Métrica | Ordenação | Árbitro | Chandy-Misra | Backoff |
|---------|-----------|---------|--------------|---------|
| Tempo total médio (3 casos) | 21,35 s | 18,68 s | 17,68 s | 21,84 s |
| Desvio padrão das esperas (Caso 2) | 1,86 s | 2,63 s | **0,69 s** | 2,57 s |
| Maior espera individual | 7,01 s (grau 2/4) | 9,01 s (grau 3) | **4,00 s (grau 2)** | 7,47 s (grau 4) |
| Outliers detectados (3 casos) | 2 (leves) | 1 (leve) | 2 (1 crítico) | 1 (leve) |
| Coeficiente de variação médio | 0,51 | 0,53 | **0,30** | 0,56 |
| Data-race free | ✓ | ✓ | ✓ | ✓ |

O Chandy-Misra obteve o melhor equilíbrio geral: menor tempo total médio (17,68 s), menor desvio padrão (0,69 s no Caso 2) e menor coeficiente de variação médio (0,30). A Ordenação e o Backoff tiveram tempos totais médios mais altos (21,35 s e 21,84 s). O Árbitro, embora rápido em alguns casos (13,01 s no Caso 3), apresentou a maior espera individual (9,01 s) e o maior desvio padrão (2,63 s).

### 5.4 Correlação Grau × Espera

| Solução | Padrão observado | Explicação |
|---------|-----------------|------------|
| Ordenação | **Invertido**: maior grau → menor espera (Caso 3) | Ordem global nivela competição |
| Árbitro | **Direto**: maior grau → maior espera (Caso 2) | Semáforo serializa acesso; mais recursos = mais contenção |
| Chandy-Misra | **Sem correlação**: espera independente do grau | Precedência por garfo elimina viés |
| Backoff | **Direto**: maior grau → maior espera (Casos 2 e 3) | Mais garrafas = maior probabilidade de falha no TryLock |

### 5.5 Análise de Starvation

O sistema de análise implementado em `core/analyze.go` detecta outliers por grau (espera > 2× a média do grupo). Os resultados mostram que:

- **Ordenação**: 2 outliers leves no Caso 3 (filo 4 com 225%, filo 5 com 250%). Ambos em valores absolutos baixos (2–3 s para 3 rodadas).
- **Árbitro**: 1 outlier leve no Caso 1 (filo 1 com 214%). Ausente nos Casos 2 e 3.
- **Chandy-Misra**: 2 outliers no Caso 3, incluindo 1 **CRÍTICO** (filo 8 com 499% — embora apenas 1,00 s absoluto). O coeficiente de variação médio mais baixo (0,30) confirma a melhor distribuição geral.
- **Backoff**: 1 outlier leve no Caso 2 (filo 1 com 211%). Ausente nos Casos 1 e 3.

Nenhuma solução apresentou starvation real (filósofos impossibilitados de beber). Todos os 48 filósofos (4 soluções × 12 nós) completaram o número esperado de rodadas em todos os casos.

### 5.6 Cenários Recomendados

- **Ordenação de Recursos**: ideal para sistemas com topologia fixa e previsível, onde a ordenação prévia das arestas é viável. Oferece o melhor equilíbrio entre simplicidade (~108 linhas) e desempenho determinístico, com a vantagem única de beneficiar filósofos de maior grau.

- **Árbitro (Garçom)**: melhor escolha para protótipos e sistemas com poucos processos (N ≤ 12), onde a simplicidade de implementação e verificação é priorizada. Não recomendada para sistemas críticos onde _fairness_ é exigida.

- **Chandy-Misra**: recomendada para sistemas distribuídos e cenários onde _fairness_ e ausência de starvation são requisitos formais. Ideal para topologias densas e alta concorrência. A complexidade de implementação (~410 linhas, 2 camadas) é compensada pela robustez.

- **Backoff Aleatório**: adequada para sistemas tolerantes a falhas onde a correção probabilística é aceitável. Útil em cenários com contenção esporádica. O `TryLock` não bloqueante evita inversão de prioridade, mas o tempo total pode ser imprevisível em alta contenção (31,16 s no Caso 2 contra 22,01 s do Chandy-Misra).

---

## 6. Análise

### 6.1 Deadlock

Nenhuma das quatro soluções implementadas apresentou deadlock durante as execuções. Todas passaram nos testes com `go test -race -count=20` nos três grafos, confirmando a ausência de deadlock e de _data races_. A solução Chandy-Misra, após a introdução da camada de garfos, foi validada com sucesso — uma versão inicial com apenas uma camada (garrafas) apresentou deadlock no Caso 3, evidenciando a sutileza do problema quando o subconjunto de recursos é arbitrário.

### 6.2 Starvation

Os resultados revelam diferenças significativas na distribuição de espera. Em termos de **coeficiente de variação** (cv = stddev/mean), que mede a dispersão relativa das esperas:

- **Chandy-Misra** (cv médio = 0,30) — distribuição mais equilibrada, sem correlação clara entre grau e espera. Confirma a garantia formal de _starvation-freedom_.
- **Ordenação** (cv médio = 0,51) — distribuição moderadamente equilibrada, com padrão invertido (maior grau, menor espera).
- **Árbitro** (cv médio = 0,53) — maior variabilidade, com viés contra graus elevados.
- **Backoff** (cv médio = 0,56) — maior dispersão relativa, com o mesmo viés do Árbitro.

No Caso 3, a topologia mostrou-se mais determinante que o grau individual para todas as soluções: filósofos em posições centrais ou com vizinhos pouco demandantes tiveram espera consistentemente menor.

### 6.3 Tempo Total de Execução

O tempo total médio nos 3 casos foi:

| Solução | Caso 1 | Caso 2 | Caso 3 | Média |
|---------|--------|--------|--------|-------|
| Ordenação | 21,01 s | 27,02 s | 16,01 s | 21,35 s |
| Árbitro | 18,01 s | 25,02 s | 13,01 s | 18,68 s |
| Chandy-Misra | 18,01 s | 22,01 s | 13,01 s | 17,68 s |
| Backoff | 16,99 s | 31,16 s | 17,36 s | 21,84 s |

O Chandy-Misra teve o menor tempo total médio (17,68 s), seguido pelo Árbitro (18,68 s). A Ordenação (21,35 s) e o Backoff (21,84 s) foram os mais lentos. O Backoff apresentou a maior variação entre casos (16,99 s a 31,16 s), refletindo a natureza probabilística do algoritmo.

### 6.4 Trade-offs Arquiteturais

A comparação entre as soluções revela um espectro que vai de **simplicidade** a **robustez**:

- **Ordenação de Recursos** (108 linhas, 1 camada): a mais simples entre as determinísticas. Sem coordenador central, sem gargalo de semáforo. Requer topologia estática.
- **Árbitro** (171 linhas, 1 camada + semáforo): simples, mas com ponto único de contenção. O semáforo N−1 é o ponto ótimo entre prevenção de deadlock e maximização da concorrência.
- **Backoff** (142 linhas, 1 camada): simples, probabilístico, sem garantias. O `TryLock` evita bloqueio, mas o tempo total varia muito com a contenção.
- **Chandy-Misra** (410 linhas, 2 camadas): o mais robusto, com garantias formais de _starvation-freedom_. A complexidade é compensada pela confiabilidade.

O valor **N − 1** no Árbitro representa o ponto ótimo no _trade-off_ entre segurança e concorrência: valores menores aumentariam a serialização; valores maiores (N) permitiriam deadlock.

---

## 7. Conclusão

Este trabalho implementou e analisou quatro soluções de sincronização para o problema do Bar dos Filósofos: **Ordenação de Recursos** (ordem total de aquisição), **Árbitro** (semáforo centralizado N−1), **Chandy-Misra** (passagem de mensagens distribuída) e **Randomized Backoff** (tentativa-e-erro com `TryLock`). Cada solução emprega uma estratégia distinta para evitar deadlock, cobrindo um espectro que vai da simplicidade determinística à robustez formal.

Os resultados experimentais confirmam que **todas as quatro soluções** atendem ao requisito fundamental de ausência de deadlock, validado por testes exaustivos com detector de _data races_ (`go test -race -count=20`). As diferenças estão na distribuição de _fairness_, no tempo total de execução e na complexidade de implementação:

- O **Chandy-Misra** destacou-se como a solução mais equilibrada: menor tempo total médio (17,68 s), menor desvio padrão das esperas (0,69 s) e coeficiente de variação médio mais baixo (0,30), confirmando a garantia formal de _starvation-freedom_.
- A **Ordenação de Recursos** revelou o padrão mais inesperado: ao contrário das demais, filósofos de maior grau tiveram espera **menor** que os de menor grau, uma consequência da ordem global de aquisição que nivela a competição.
- O **Árbitro**, apesar de ser o mais simples conceitualmente, exibiu o maior viés contra graus elevados e a maior espera individual (9,01 s).
- O **Backoff** apresentou a maior variação de desempenho (16,99 s a 31,16 s), refletindo sua natureza probabilística. O uso de `TryLock` elimina espera bloqueante, mas o algoritmo não oferece garantias formais.

A principal lição do desenvolvimento foi a necessidade de rigor na implementação de algoritmos concorrentes: uma versão inicial do Chandy-Misra com apenas uma camada (garrafas) apresentou deadlock no caso de maior conectividade, e a Ordenação de Recursos exige que o subconjunto seja ordenado **antes** do loop de locks — uma sutileza que comprometeria a correção se omitida.

Para trabalhos futuros, sugerimos: (a) execução com um número maior de rodadas (R ≥ 30) para estabilizar as médias; (b) avaliação em grafos de maior escala (N ≥ 100) para estressar a escalabilidade; (c) introdução de métricas adicionais como _throughput_ (bebidas por segundo) e eficiência de escalabilidade (_speedup_); e (d) implementação de _exponential backoff_ na solução de Backoff para melhorar o desempenho em alta contenção.

---

## 8. Como Reproduzir

**Pré-requisito:** Go 1.21 ou superior instalado.

```sh
# Descompactar e acessar
cd concurrent-programming-work/

# Executar uma solução em um grafo específico
make run SOL=ordenacao GRAFO=data/caso1_jantar_5.txt RODADAS=6

# Executar todas as 4 soluções × 3 grafos
make all

# Gerar tabela comparativa
make compare

# Validar data races e deadlock
make race

# Execução direta sem Makefile
go run ./cmd/runner -solucao=chandy_misra -grafo=data/caso3_bar_12.txt -rodadas=3
```

Os resultados são armazenados como TSV em `results/caso<N>/<solucao>.txt`, incluindo a seção `# starvation analysis:` gerada automaticamente pelo pacote `core/analyze.go`.

---

## Referências

- Chandy, K. M., & Misra, J. (1984). The Drinking Philosophers Problem. *ACM Transactions on Programming Languages and Systems (TOPLAS)*, 6(4), 632–646.
- Dijkstra, E. W. (1971). Hierarchical ordering of sequential processes. *Acta Informatica*, 1(2), 115–138.
- Coffman, E. G., Elphick, M. J., & Shoshani, A. (1971). System deadlocks. *ACM Computing Surveys (CSUR)*, 3(2), 67–78.
- Havender, J. W. (1968). Avoiding deadlock in multitasking systems. *IBM Systems Journal*, 7(2), 74–84.
- Tanenbaum, A. S., & Bos, H. (2015). *Modern Operating Systems* (4th ed.). Pearson.
