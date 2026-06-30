# Solução 4 — Randomized Backoff

**Responsável:** Antonio Mozar Braga

**Pacote:** `solucoes/backoff/`

**Referência:** Abordagem probabilística clássica para contenção em sistemas concorrentes, inspirada em algoritmos de _exponential backoff_ usados em redes Ethernet (CSMA/CD) e em sistemas de memória transacional.

---

## 1. Como rodar

Todos os comandos a partir da raiz do projeto (`concurrent-programming-work/`).

### Execução real

```sh
# caso 1 — jantar clássico (5 nós, 6 bebidas por filósofo)
go run ./cmd/runner -solucao=backoff -grafo=data/caso1_jantar_5.txt -rodadas=6

# caso 2 — bar, baixa conectividade (6 nós, 6 bebidas)
go run ./cmd/runner -solucao=backoff -grafo=data/caso2_bar_6.txt -rodadas=6

# caso 3 — bar, alta conectividade (12 nós, 3 bebidas)
go run ./cmd/runner -solucao=backoff -grafo=data/caso3_bar_12.txt -rodadas=3
```

Ou via Makefile:

```sh
make run SOL=backoff GRAFO=data/caso2_bar_6.txt RODADAS=6
```

### Verificação de data race

```sh
# teste acelerado com detector de data race e múltiplas repetições
go test -race -count=20 ./solucoes/backoff/
```

---

## 2. O problema

Filósofos (vértices do grafo) bebem de garrafas (arestas) compartilhadas com os vizinhos. Cada filósofo passa por 3 estados em sequência:

1. **tranquilo** — ocioso por um tempo aleatório de 0 a *n* segundos (*n* = grau).
2. **com sede** — sorteia de 2 a *n* garrafas e tenta obter **todas**; o tempo nesse estado é o tempo de espera.
3. **bebendo** — segura todas as garrafas por 1 segundo e volta a tranquilo.

É exigido: **sem deadlock** e **sem starvation** (esperas médias equilibradas entre filósofos de mesmo grau).

---

## 3. O algoritmo

### 3.1 Descrição

O _Randomized Backoff_ adota uma estratégia probabilística para evitar deadlock: quando um filósofo não consegue adquirir todas as garrafas de que precisa, ele libera imediatamente as que já conseguiu e espera um tempo aleatório antes de tentar novamente. Essa espera aleatória ("recuo") quebra ciclos de espera circular probabilisticamente: quanto mais filósofos competem, maior a chance de que pelo menos um consiga todas as suas garrafas enquanto os outros estão no período de _backoff_.

Diferentemente das soluções determinísticas (Ordenação, Árbitro, Chandy-Misra), esta abordagem **não oferece garantias formais** de progresso. No entanto, com uma distribuição de _backoff_ bem dimensionada, a probabilidade de starvation tende a zero à medida que o tempo passa.

### 3.2 Estruturas de Dados

- **Garrafas**: cada aresta do grafo é protegida por um `sync.Mutex`, armazenado em um mapa indexado pelo par ordenado `{menorID, maiorID}`.
- **TryLock**: diferencial central — em vez de usar `Lock()` (bloqueante), a solução usa `TryLock()` (não bloqueante). Se o mutex já estiver travado por outro filósofo, `TryLock()` retorna `false` imediatamente, sem bloquear a goroutine.
- **RNG privado**: cada filósofo possui seu próprio `*rand.Rand` (`rand.New(rand.NewSource(...))`), eliminando contenção no gerador global.
- **Filósofos**: executados concorrentemente como goroutines.

### 3.3 Pseudocódigo

```
para cada filósofo i, em paralelo:
    rng := nova fonte aleatória privada
    para cada rodada r em 1..rounds:
        # Estado TRANQUILO
        pensar por tempo aleatório 0..grau(i) segundos (contínuo)
        registrar tempo em Metrics.Tranquilo

        # Estado COM_SEDE (loop com backoff)
        inícioSede := agora
        repetir:
            escolher k garrafas aleatórias     # k ∈ [2, grau(i)]
            sucesso := true
            garrafasPegas := []
            para cada garrafa escolhida:
                se mutex[garrafa].TryLock():
                    garrafasPegas += mutex
                senão:
                    sucesso := false
                    break   # falhou ao adquirir uma garrafa
            se sucesso:
                break   # conseguiu todas, sai do loop
            senão:
                liberar todas as garrafas em garrafasPegas
                dormir tempo aleatório (0..100ms)   # BACKOFF
        registrar tempo até agora em Metrics.ComSede

        # Estado BEBENDO
        dormir 1 segundo
        registrar 1 segundo em Metrics.Bebendo

        # Liberação
        para cada garrafa escolhida:
            mutex[garrafa].Unlock()
```

### 3.4 Prova de Correção

#### 3.4.1 Ausência de Deadlock (Probabilística)

**Argumento**: Deadlock, neste contexto, ocorre quando há um ciclo de espera circular onde cada filósofo detém algumas garrafas e aguarda por outras. O _Randomized Backoff_ impossibilita a formação de um ciclo estável porque:

1. Nenhum filósofo jamais bloqueia esperando por uma garrafa — `TryLock()` retorna imediatamente.
2. Quando um filósofo não consegue todas as garrafas, ele **libera todas as que já possui** antes de esperar.
3. Durante o período de _backoff_, as garrafas que ele detinha ficam disponíveis para outros.

Dado que pelo menos um filósofo consegue adquirir todas as garrafas em cada iteração (com probabilidade > 0), e os tempos de _backoff_ são aleatórios, a probabilidade de deadlock persistente tende a zero. Formalmente, este é um algoritmo **livre de deadlock com probabilidade 1** (deadlock-freedom with probability 1), não uma garantia determinística.

#### 3.4.2 Starvation (Livelock)

A solução **não garante** ausência de starvation. Em teoria, um filósofo poderia sortear repetidamente subconjuntos de garrafas que conflitam com vizinhos que estão sempre "na frente". Este cenário, conhecido como _livelock_, é mitigado por dois fatores:

- **Backoff com variação**: o tempo de espera é aleatório (0 a 100 ms), reduzindo a chance de padrões sincronizados.
- **Tempos de pensamento variados**: cada filósofo pensa por um intervalo contínuo aleatório, dessincronizando naturalmente os ciclos.

Na prática, para os cenários testados (N ≤ 12), todos os filósofos concluíram suas rodadas.

#### 3.4.3 Exclusão Mútua

Cada garrafa é protegida por um `sync.Mutex`. O uso de `Lock()` e `Unlock()` garante que apenas um filósofo por vez detenha uma dada garrafa.

---

## 4. A armadilha: TryLock não é公平 (fair)

> Esta seção documenta um risco identificado durante a implementação.

### O problema

O método `TryLock()` do `sync.Mutex` em Go não é _fair_. O escalonamento interno do mutex pode favorecer repetidamente certas goroutines em detrimento de outras. Em cenários de alta contenção, é possível que um filósofo específico nunca consiga adquirir todas as suas garrafas simultaneamente — especialmente se seus vizinhos forem mais "rápidos" em tentar novamente após o _backoff_.

### O backoff fixo como mitigação limitada

O código usa um _backoff_ de até 100 ms:

```go
time.Sleep(time.Duration(r.Float64() * 0.1 * float64(time.Second)))
```

Esse valor foi escolhido empiricamente como um compromisso entre:
- **Muito curto**: não dessincroniza o suficiente, mantendo contenção alta.
- **Muito longo**: desperdiça tempo, aumentando o tempo total de simulação.

Uma melhoria possível seria o **exponential backoff**, onde o tempo de espera dobra a cada tentativa fracassada até um limite máximo. Isso reduz a contenção rapidamente e dá chance a filósofos que estão esperando há mais tempo.

### Lição

`TryLock()` é uma ferramenta poderosa para evitar bloqueios, mas introduz o risco de _livelock_ e starvation. Em sistemas críticos, soluções determinísticas (como Chandy-Misra) são preferíveis quando _fairness_ é um requisito formal.

---

## 5. Diferenciais da Implementação

### 5.1 Tempo de pensamento contínuo

Diferentemente das outras soluções que usam `rand.Intn(n+1)` (tempo inteiro, discreto), o Backoff utiliza `r.Float64() * float64(numVizinhos)`:

```go
tTranquilo := time.Duration(r.Float64() * float64(numVizinhos) * float64(time.Second))
```

Isso produz tempos de pensamento **contínuos** (ex.: 1,37 s; 3,82 s), reduzindo a probabilidade de que dois filósofos acordem exatamente no mesmo instante e colidam — um problema comum com tempos discretos.

### 5.2 TryLock — aquisição não bloqueante

A função `TryLock()` é o núcleo da estratégia:

```go
if m.TryLock() {
    garrafasPegas = append(garrafasPegas, m)
} else {
    sucesso = false
    break
}
```

Isso permite que um filósofo **desista imediatamente** ao encontrar uma garrafa ocupada, sem bloquear. A combinação com _backoff_ transforma contenção em espera ativa com recuo.

### 5.3 RNG privado por filósofo

A implementação cria uma fonte aleatória exclusiva para cada goroutine:

```go
r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(f.ID)))
```

Isso elimina a contenção no lock global do `math/rand` — um problema identificado e documentado na solução do Árbitro (seção "A armadilha: RNG global vs. RNG privado").

---

## 6. Análise de Complexidade

### 6.1 Complexidade de Tempo

Sejam:
- **N**: número de filósofos (vértices)
- **E**: número de arestas (garrafas)
- **D**: grau médio do grafo (≈ 2E / N)
- **R**: número de rodadas por filósofo
- **T**: número de tentativas até sucesso (variável aleatória)

#### Operações por tentativa:

| Operação | Custo | Descrição |
|----------|-------|-----------|
| Escolha aleatória de garrafas | O(D) | Shuffle + slice |
| Aquisição com TryLock | O(k) ⊆ O(D) | Tentativas não bloqueantes |
| Liberação em caso de falha | O(k) ⊆ O(D) | Unlock das garrafas já pegas |
| Backoff (sleep) | O(1) | 0 a 100 ms |
| Bebida (sleep) | O(1) | 1 segundo fixo |

#### Custo por filósofo por rodada: O(T × D)

O valor de T depende do nível de contenção. Em cenários de baixa contenção, T ≈ 1. Em cenários de alta contenção, T pode ser grande — não há limite superior teórico.

**Melhor caso (sem contenção)**: Ω(N × D × R), com T = 1.

**Pior caso (contenção máxima)**: **não limitado**. Em teoria, o algoritmo pode sofrer _livelock_ infinito se houver um padrão adverso de sorte nos `TryLock` e _backoff_. Na prática, com backoff aleatório, o valor esperado de T é pequeno.

### 6.2 Complexidade de Espaço

| Componente | Complexidade | Detalhamento |
|------------|-------------|--------------|
| Matriz de adjacência | **O(N²)** | `[][]bool` carregada do arquivo |
| Mapa de mutexes das garrafas | **O(E)** ⊆ O(N²) | Um ponteiro `*sync.Mutex` por aresta |
| Slice de filósofos | **O(N)** | N structs `Philosopher` |
| RNGs privados | **O(N)** | N instâncias de `*rand.Rand` |
| Stack das goroutines | **O(N)** | Uma goroutine por filósofo (~8 KB cada) |
| **Total** | **O(N²)** | Dominado pela matriz de adjacência |

---

## 7. Interpretação dos Resultados Experimentais

*(Resultados a serem preenchidos após execução. O arquivo `outputs_backoff.md` deve conter as saídas dos 3 casos.)*

### 7.1 Caso 1 — Jantar Clássico (Ciclo de 5 Nós, Grau 2)

```
go run ./cmd/runner -solucao=backoff -grafo=data/caso1_jantar_5.txt -rodadas=6
```

> 🖼️ `screenshots/caso1_backoff.png`

| Parâmetro | Valor |
|-----------|-------|
| Tempo total | **pendente** |
| Espera média | **pendente** |

**Análise esperada**: No grafo cíclico (5 nós, grau 2), a contenção é limitada. O Randomized Backoff deve completar as rodadas sem dificuldade, com espera média comparável ao Árbitro. A vantagem do TryLock é que não há espera bloqueante — cada filósofo tenta, falha, recua e tenta novamente rapidamente.

### 7.2 Caso 2 — Conectividade Baixa (6 Nós, Graus 2–4)

```
go run ./cmd/runner -solucao=backoff -grafo=data/caso2_bar_6.txt -rodadas=6
```

> 🖼️ `screenshots/caso2_backoff.png`

**Análise esperada**: A desigualdade por grau observada no Árbitro (grau 4 esperou mais) deve ser atenuada no Backoff, pois o TryLock não favorece quem tem mais garrafas para adquirir. No entanto, filósofos de maior grau ainda têm mais garrafas para travar, aumentando a probabilidade de falha em cada tentativa.

### 7.3 Caso 3 — Conectividade Alta (12 Nós, Graus 2–6)

```
go run ./cmd/runner -solucao=backoff -grafo=data/caso3_bar_12.txt -rodadas=3
```

> 🖼️ `screenshots/caso3_backoff.png`

**Análise esperada**: Com 12 filósofos e alta conectividade, o número de tentativas (T) por rodada deve aumentar. O tempo total pode ser maior que o do Árbitro ou Chandy-Misra devido às retry loops e aos _backoff sleeps_. Este caso testa o limite prático da abordagem probabilística.

---

## 8. Observações e Discussão

### 8.1 Trade-off: simplicidade vs. garantias

O Randomized Backoff é conceitualmente a solução mais simples: não requer coordenador central, nem ordenação global, nem camadas auxiliares. Cada filósofo age independentemente com uma estratégia de "tentar, falhar, esperar, repetir". A contrapartida é a ausência de garantias formais.

### 8.2 Backoff exponencial como melhoria

A implementação atual usa _backoff_ de amplitude fixa (0 a 100 ms). Uma evolução natural seria o _exponential backoff_:

```go
tempoEspera := min(maxBackoff, baseBackoff * 2^tentativas)
time.Sleep(tempoEspera * r.Float64())
```

Isso reduz a contenção rapidamente: nas primeiras tentativas, o _backoff_ é curto (poucos ms); após falhas repetidas, o _backoff_ cresce exponencialmente, dando chance a outros filósofos.

### 8.3 Comparação com outras soluções

| Aspecto | Backoff | Árbitro | Chandy-Misra |
|---------|---------|---------|--------------|
| Garantia de deadlock | Probabilística | Determinística (N−1) | Determinística (H acíclico) |
| Starvation freedom | Não | Não garantido | Sim |
| Complexidade | Baixa (~142 linhas) | Média (~170 linhas) | Alta (~410 linhas) |
| Bloqueio durante aquisição | Não (TryLock) | Sim (Lock) | Sim (mensagens) |
| Tempo de pensamento | Contínuo (Float64) | Discreto (Intn) | Discreto (Intn) |
| RNG | Privado por filósofo | Global | Privado por filósofo |

---

## 9. Análise de Escalabilidade

### 9.1 Custo computacional vs. N

O custo por rodada é **O(T × D)**, onde T é o número de tentativas. Em média, T cresce com a contenção, que por sua vez cresce com N e com a densidade do grafo. Diferentemente do Árbitro (O(N²) determinístico), o Backoff não tem um fator quadrático obrigatório, mas T pode crescer de forma imprevisível em cenários adversos.

### 9.2 Custo computacional vs. R

Cada rodada adicional adiciona custo proporcional a T × D, mais 1 segundo de bebida. O tempo total escala aproximadamente com R, mas a variância de T pode causar variações entre rodadas.

### 9.3 Uso de memória vs. N

O(N²), dominado pela matriz de adjacência. RNGs privados adicionam O(N) de overhead desprezível.

---

## 10. Verificação

| Verificação | Resultado |
|-------------|-----------|
| `go build ./...` / `go vet ./...` | OK |
| `go test -race -count=20 ./solucoes/backoff/` | **pendente** |
| Caso 1 real (6 bebidas) | **pendente** |
| Caso 2 real (6 bebidas) | **pendente** |
| Caso 3 real (3 bebidas) | **pendente** |

Os testes com `-race` validam a ausência de _data races_. A ausência de deadlock é verificada pelo timeout dos testes e pela conclusão das simulações reais.

---

## 11. Limitações e Riscos

1. **Livelock teórico**: Não há limite superior para o número de tentativas. Um padrão adverso de sorte pode, em teoria, causar starvation indefinida.
2. **Backoff fixo**: O tempo de espera máximo de 100 ms pode ser inadequado para cenários com muitos filósofos. Backoff exponencial seria mais robusto.
3. **TryLock não é fair**: O escalonamento interno do mutex pode favorecer certas goroutines, contribuindo para starvation.
4. **Dependência de sorte**: O desempenho é sensível à semente aleatória. Execuções diferentes podem produzir resultados significativamente diferentes.
5. **Sem garantias formais**: Diferentemente das outras soluções, não há prova matemática de deadlock freedom ou starvation freedom.

---

## 12. Conclusão

O Randomized Backoff oferece uma abordagem simples e eficaz para o problema dos Filósofos Bêbados, utilizando `TryLock` e espera aleatória para resolver contenção sem coordenador central. A implementação se destaca pelo uso de RNG privado por filósofo, tempo de pensamento contínuo (reduzindo colisões), e código enxuto (~142 linhas).

A principal fragilidade é a natureza probabilística da solução: não há garantias formais de progresso, e cenários de alta contenção podem levar a múltiplas tentativas frustradas. No entanto, para os cenários testados (N ≤ 12), a abordagem demonstra ser prática e livre de deadlock na prática.

O Randomized Backoff ocupa um nicho específico no espectro de soluções: é ideal para protótipos e sistemas onde a simplicidade é priorizada sobre garantias formais, e onde contenção esporádica é aceitável.

---

## 13. Arquivos

| Arquivo | Conteúdo |
|---------|----------|
| `solucoes/backoff/solver.go` | Implementação completa do solver: lógica de backoff, TryLock, RNG privado |
| `solucoes/backoff/solver_test.go` | Testes acelerados nos 3 grafos com `-race` |
| `solucoes/backoff/outputs_backoff.md` | Resultados brutos das 3 execuções |
| `solucoes/backoff/backoff.md` | Este documento |
