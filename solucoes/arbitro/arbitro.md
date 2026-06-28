# Solução 2 — Árbitro (Garçom)

**Responsável:** Samuel William Silva Almeida

**Pacote:** `solucoes/arbitro/`

**Referência:** Clássica solução do garçom (_waiter_) para o problema dos filósofos famintos, adaptada para o problema dos Filósofos Bêbados (_Drinking Philosophers Problem_, Dijkstra, 1971).

---

## 1. Como rodar

Todos os comandos a partir da raiz do projeto (`concurrent-programming-work/`).

### Execução real

```sh
# caso 1 — jantar clássico (5 nós, 6 bebidas por filósofo)
go run ./cmd/runner -solucao=arbitro -grafo=data/caso1_jantar_5.txt -rodadas=6

# caso 2 — bar, baixa conectividade (6 nós, 6 bebidas)
go run ./cmd/runner -solucao=arbitro -grafo=data/caso2_bar_6.txt -rodadas=6

# caso 3 — bar, alta conectividade (12 nós, 3 bebidas)
go run ./cmd/runner -solucao=arbitro -grafo=data/caso3_bar_12.txt -rodadas=3
```

Ou via Makefile:

```sh
make run SOL=arbitro GRAFO=data/caso2_bar_6.txt RODADAS=6
```

### Verificação de data race

```sh
# teste acelerado com detector de data race e múltiplas repetições
go test -race -count=20 ./solucoes/arbitro/
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

A solução do árbitro central, também conhecida como solução do garçom (_waiter_), introduz um coordenador global responsável por serializar o acesso aos recursos compartilhados. O princípio é simples, porém elegantemente eficaz: o árbitro mantém um semáforo de capacidade **N − 1**, onde N é o número total de filósofos. Antes de qualquer filósofo tentar adquirir garrafas, ele deve obter uma autorização do árbitro. Caso o número máximo de filósofos simultâneos já tenha sido atingido, o filósofo é bloqueado até que outro termine de beber e libere sua vaga.

### 3.2 Estruturas de Dados

- **Árbitro**: implementado como um canal Go bufferizado com capacidade N − 1. O envio de um token (`chan <- struct{}{}`) representa a aquisição de permissão (operação P do semáforo); o recebimento (`<-chan`) representa a liberação (operação V).
- **Garrafas**: cada aresta do grafo é protegida por um `sync.Mutex` independente, armazenado em um mapa indexado pelo par ordenado `{menorID, maiorID}`. O uso de `map[[2]int]*sync.Mutex` garante que ambos os lados da aresta acessem o mesmo mutex.
- **Filósofos**: executados concorrentemente como goroutines, cada um seguindo o ciclo `TRANQUILO → COM_SEDE → BEBENDO`.

### 3.3 Pseudocódigo

```
para cada filósofo i, em paralelo:
    para cada rodada r em 1..rounds:
        # Estado TRANQUILO
        pensar por tempo aleatório 0..grau(i) segundos
        registrar tempo em Metrics.Tranquilo

        # Estado COM_SEDE
        arbiter.acquire()                  # semáforo P(): bloqueia se N-1 ativos
        escolher k garrafas aleatórias     # k ∈ [2, grau(i)]
        para cada garrafa escolhida:
            mutex[garrafa].Lock()
        registrar tempo em Metrics.ComSede

        # Estado BEBENDO
        dormir 1 segundo
        registrar tempo em Metrics.Bebendo

        # Liberação
        para cada garrafa escolhida:
            mutex[garrafa].Unlock()
        arbiter.release()                  # semáforo V(): libera vaga
```

### 3.4 Prova de Correção

#### 3.4.1 Livre de Deadlock

**Teorema**: A solução do árbitro com limite N − 1 é livre de deadlock.

**Demonstração**: Considere N filósofos e um semáforo de capacidade N − 1. No máximo N − 1 filósofos podem estar simultaneamente no estado COM_SEDE ou BEBENDO. Portanto, em qualquer instante, **pelo menos um filósofo está no estado TRANQUILO**.

Suponha, por absurdo, que ocorra um deadlock. Isso significa que todos os filósofos ativos estão retendo algumas garrafas e aguardando por outras — uma espera circular. No entanto, o filósofo que está TRANQUILO não retém nenhuma garrafa. Quando ele acordar e tentar entrar em COM_SEDE, encontrará o semáforo com pelo menos uma vaga disponível (pois no máximo N − 1 estão ocupados). Como ele não retém recurso algum, ele conseguirá adquirir todas as garrafas de que precisa — mesmo que todos os N − 1 ativos estejam segurando uma garrafa cada, ainda sobra pelo menos uma garrafa livre. Logo, o progresso é sempre possível, e o deadlock é impossível.

#### 3.4.2 Livre de Starvation (Fairness)

A solução não garante _starvation-freedom_ estrita. O semáforo do árbitro é do tipo _não-fair_ (por padrão, canais Go são _fair_ apenas na seleção entre goroutines prontas, mas não há fila de prioridade explícita). Na prática, para os casos testados, todos os filósofos conseguiram beber o número esperado de rodadas, indicando ausência de starvation no cenário experimental. Contudo, em teoria, um filósofo poderia ser repetidamente preterido se o árbitro sempre favorecesse outros — situação que se torna improvável com tempos de pensamento aleatórios.

#### 3.4.3 Exclusão Mútua

Cada garrafa (aresta) é protegida por um `sync.Mutex`. Apenas um filósofo por vez pode segurar uma dada garrafa, garantindo que nenhum par de vizinhos beba simultaneamente usando o mesmo recurso compartilhado.

---

## 4. A armadilha: RNG global vs. RNG privado

> Esta seção documenta um problema real de desempenho identificado durante o desenvolvimento e como foi diagnosticado.

### O problema

A primeira versão da implementação utilizava o gerador de números aleatórios global do Go, `math/rand.Intn()`, para sortear o tempo de pensamento:

```go
thinkingTime := time.Duration(rand.Intn(n+1)) * time.Second
```

O pacote `math/rand` do Go utiliza um gerador global protegido por um `sync.Mutex` interno. Isso significa que, sempre que um filósofo chama `rand.Intn()`, ele compete com todos os outros filósofos pelo **mesmo lock**. Com N goroutines chamando `rand.Intn()` intensamente (a cada rodada, para o tempo de pensamento), esse lock global se torna um gargalo de contenção oculto.

### Impacto medido

Para ilustrar o impacto, considere um cenário com N goroutines disputando o mesmo `rand.Intn()`:

```
N filósofos chamando rand.Intn() em paralelo
         │
         ▼
    ┌────────────────────┐
    │  lock global do    │  ← contenção: apenas 1 passa por vez
    │  math/rand         │
    └────────────────────┘
         │
    ┌────┴────┬────┬────┐
    │  G0     │ G1 │ G2 │ ...  ← N-1 goroutines bloqueadas no lock
    └─────────┴────┴────┘
```

Com N = 12 e rodadas rápidas, a contenção no RNG global poderia degradar o desempenho visivelmente — cada chamada `rand.Intn()` serializa todas as goroutines.

### A correção

Criar um RNG **privado por filósofo** usando `rand.New(rand.NewSource(...))`:

```go
rng := rand.New(rand.NewSource(int64(id)*7919 + 1))
```

No entanto, ao contrário da solução Chandy-Misra que adotou RNG privado, a solução do Árbitro **manteve o RNG global** intencionalmente, pelas seguintes razões:

1. **Simplicidade**: o RNG global é a abordagem mais simples e direta para um código que precisa ser didático.
2. **Impacto irrelevante para N ≤ 12**: com poucos filósofos e tempos de pensamento de 0 a 6 segundos, o lock do RNG global raramente é disputado — as goroutines passam a maior parte do tempo dormindo.
3. **Foco arquitetural**: a lição central da solução do Árbitro é o semáforo e a prevenção de deadlock, não a micro-otimização do gerador aleatório.

### Lição

RNG global com mutex interno é uma **fonte silenciosa de contenção** em programas concorrentes. Em aplicações com muitas threads e chamadas frequentes a `rand`, um RNG privado por thread (ou `sync.Pool` de geradores) elimina esse gargalo. A partir de Go 1.20, a função `math/rand/v2` amenizou parte do problema com locks mais granulares, mas a recomendação para código concorrente intensivo permanece: **cada goroutine com seu próprio `*rand.Rand`**.

---

## 5. Análise de Complexidade

### 5.1 Complexidade de Tempo

Sejam:
- **N**: número de filósofos (vértices)
- **E**: número de arestas (garrafas)
- **D**: grau médio do grafo (≈ 2E / N)
- **R**: número de rodadas por filósofo

#### Operações internas por rodada:

| Operação | Custo | Descrição |
|----------|-------|-----------|
| Escolha aleatória de garrafas | O(D) | _shuffle_ de Fisher-Yates parcial |
| Aquisição de garrafas | O(k) ⊆ O(D) | Lock sequencial de k mutexes |
| Bebida (sleep) | O(1) | 1 segundo fixo (delay real) |
| Liberação de garrafas | O(k) ⊆ O(D) | Unlock sequencial |
| Semáforo do árbitro | O(1) | Operação em canal |

**Custo por filósofo por rodada**: O(D) operações, mais 1 segundo de sono.

**Custo total sem contenção (melhor caso)**: Ω(N × D × R)

No melhor cenário — filósofos com tempos de pensamento longos e baixa sobreposição — poucos disputam recursos simultaneamente, e o semáforo raramente bloqueia.

**Custo total com contenção máxima (pior caso)**: O(N² × R)

No pior cenário, todos os N − 1 ativos competem intensamente. Cada acquire do último filósofo pode esperar que todos os N − 2 anteriores completem seu ciclo (incluindo 1 segundo de bebida cada). Isso introduz um fator multiplicativo O(N) de espera sobre as operações internas, resultando em O(N² × R).

### 5.2 Complexidade de Espaço

| Componente | Complexidade | Detalhamento |
|------------|-------------|--------------|
| Matriz de adjacência | **O(N²)** | `[][]bool` carregada do arquivo |
| Mapa de mutexes das garrafas | **O(E)** ⊆ O(N²) | Um ponteiro `*sync.Mutex` por aresta |
| Canal do semáforo | **O(N)** | Buffer de tamanho N − 1 |
| Slice de filósofos | **O(N)** | N structs `Philosopher` |
| Stack das goroutines | **O(N)** | Uma goroutine por filósofo (~8 KB cada) |
| **Total** | **O(N²)** | Dominado pela matriz de adjacência |

**Observação**: O mapa de mutexes gera O(E) entradas, mas no pior caso (grafo completo) E = N(N−1)/2, ou seja, O(N²). Logo, o espaço total permanece O(N²).

---

## 6. Interpretação dos Resultados Experimentais

### 6.1 Caso 1 — Jantar Clássico (Ciclo de 5 Nós, Grau 2)

```
go run ./cmd/runner -solucao=arbitro -grafo=data/caso1_jantar_5.txt -rodadas=6
```

| Parâmetro | Valor |
|-----------|-------|
| Tempo total | 18,01 s |
| Espera média | 2,80 s |
| Todos os filósofos | Grau 2 |

| Filósofo | Grau | Tranquilo (s) | Com Sede (s) | Bebendo (s) | Bebidas |
|----------|------|---------------|--------------|-------------|---------|
| 0 | 2 | 7,00 | 2,00 | 6,00 | 6 |
| 1 | 2 | 6,00 | 4,00 | 6,00 | 6 |
| 2 | 2 | 10,00 | 2,00 | 6,00 | 6 |
| 3 | 2 | 6,00 | 4,00 | 6,00 | 6 |
| 4 | 2 | 9,00 | 2,00 | 6,00 | 6 |

**Análise**: No grafo cíclico clássico, todos os 5 filósofos possuem exatamente 2 vizinhos. O limite do árbitro é 4 ativos simultâneos. A espera média de 2,80 s mostra que houve contenção moderada, mas todos completaram as 6 rodadas dentro do esperado. Os tempos de espera individuais variaram de 2 s a 4 s, indicando alternância razoável entre os filósofos — não houve monopolização do árbitro.

**Fato relevante**: Neste caso específico, poderíamos ter 4 filósofos ativos e 1 tranquilo. Como todos têm grau 2, cada filósofo precisa de 2 garrafas. Mesmo que 4 ativos seguem 4 das 5 garrafas, o filósofo tranquilo, ao acordar, consegue as 2 garrafas restantes. O deadlock é estruturalmente impossível.

### 6.2 Caso 2 — Conectividade Baixa (6 Nós, Graus 2–4)

```
go run ./cmd/runner -solucao=arbitro -grafo=data/caso2_bar_6.txt -rodadas=6
```

| Parâmetro | Valor |
|-----------|-------|
| Tempo total | 28,02 s |

| Filósofo | Grau | Tranquilo (s) | Com Sede (s) | Bebendo (s) | Bebidas |
|----------|------|---------------|--------------|-------------|---------|
| 0 | 2 | 5,00 | 4,00 | 6,00 | 6 |
| 1 | 2 | 5,00 | 1,00 | 6,00 | 6 |
| 2 | 3 | 12,00 | 5,00 | 6,00 | 6 |
| 3 | 3 | 9,00 | 3,00 | 6,00 | 6 |
| 4 | 2 | 3,00 | 0,00 | 6,00 | 6 |
| 5 | 4 | 15,00 | 7,01 | 6,00 | 6 |

**Espera média por grau:**

| Grau | Média (s) | n |
|------|-----------|---|
| 2 | 1,67 | 3 |
| 3 | 4,00 | 2 |
| 4 | 7,01 | 1 |

**Análise**: Este caso revela um **padrão de desigualdade por grau**. Filósofos com grau 2 (nós 0, 1, 4) tiveram espera média de apenas 1,67 s, enquanto o filósofo de grau 4 (nó 5) esperou 7,01 s — mais de 4 vezes maior. A explicação é direta: filósofos de maior grau precisam adquirir mais garrafas (k entre 2 e grau), aumentando a probabilidade de contenção. Além disso, ao segurar mais recursos, eles bloqueiam mais vizinhos, que por sua vez demoram mais para liberar.

**Fato relevante**: O filósofo 4 (grau 2) teve **0 segundos de espera** — ele conseguiu todas as garrafas imediatamente em todas as rodadas. Isso ocorre porque seus vizinhos (de maior grau) demoravam mais para adquirir recursos, deixando as garrafas do nó 4 frequentemente disponíveis. Note que, como mostram os resultados do Caso 2, os tempos de pensamento não são idênticos entre filosofos, mas sao determinados pelo grau de cada um — variando de O a n segundos. Isso influencia diretamente os padroes de concorrencia.

### 6.3 Caso 3 — Conectividade Alta (12 Nós, Graus 2–6)

```
go run ./cmd/runner -solucao=arbitro -grafo=data/caso3_bar_12.txt -rodadas=3
```

| Parâmetro | Valor |
|-----------|-------|
| Tempo total | 19,01 s |

| Filósofo | Grau | Tranquilo (s) | Com Sede (s) | Bebendo (s) | Bebidas |
|----------|------|---------------|--------------|-------------|---------|
| 0 | 3 | 5,00 | 1,00 | 3,00 | 3 |
| 1 | 3 | 3,00 | 0,00 | 3,00 | 3 |
| 2 | 2 | 1,00 | 1,00 | 3,00 | 3 |
| 3 | 4 | 11,00 | 2,00 | 3,00 | 3 |
| 4 | 4 | 10,00 | 2,00 | 3,00 | 3 |
| 5 | 3 | 1,00 | 0,00 | 3,00 | 3 |
| 6 | 5 | 14,00 | 2,00 | 3,00 | 3 |
| 7 | 4 | 3,00 | 3,00 | 3,00 | 3 |
| 8 | 3 | 5,00 | 0,00 | 3,00 | 3 |
| 9 | 6 | 14,00 | 1,00 | 3,00 | 3 |
| 10 | 2 | 4,00 | 3,00 | 3,00 | 3 |
| 11 | 3 | 4,00 | 1,00 | 3,00 | 3 |

**Espera média por grau:**

| Grau | Média (s) | n |
|------|-----------|---|
| 2 | 2,00 | 2 |
| 3 | 0,40 | 5 |
| 4 | 2,34 | 3 |
| 5 | 2,00 | 1 |
| 6 | 1,00 | 1 |

**Análise**: Com 12 filósofos e apenas 3 rodadas cada, o tempo total foi de 19,01 s — coerente com 36 bebidas de 1 s cada (36 s de bebida pura) mais tempos de pensamento, porém com sobreposição substancial (até 11 simultâneos). A espera média por grau não segue uma relação monotônica: o grau 3 teve a menor espera (0,40 s), enquanto o grau 2 teve 2,00 s e o grau 4 teve 2,34 s. Isso ocorre porque a topologia do grafo (quem é vizinho de quem) importa tanto quanto o grau individual. Um filósofo de grau 3 cercado por vizinhos de baixo grau sofre menos contenção que um de grau 2 cercado por vizinhos de alto grau.

**Fato relevante**: Os filósofos 1, 5 e 8 (todos de grau 3) tiveram **0 segundos de espera**. Isso sugere que eles estão posicionados em regiões do grafo com baixa contenção — possivelmente conectados a filósofos que passam mais tempo pensando ou que competem por garrafas diferentes.

---

## 7. Observações e Discussão

### 7.1 A topologia importa mais que o grau

Os resultados do Caso 3 demonstram que a **posição no grafo** é tão determinante quanto o grau para o tempo de espera. O nó 5 (grau 3, espera 0 s) e o nó 3 (grau 4, espera 2 s) exemplificam que a distribuição dos vizinhos e seus respectivos padrões de escolha de garrafas influenciam fortemente a contenção. A métrica de "espera média por grau" agrega demais e esconde heterogeneidades locais.

### 7.2 Efeito de serialização do árbitro

O semáforo de capacidade N − 1 resolve o deadlock, mas introduz um **gargalo de serialização**. Quando muitos filósofos ficam com sede simultaneamente, o canal do árbitro funciona como um _ponto de contenção único_ (_single point of contention_). Todos os filósofos disputam o mesmo canal, o que, em sistemas com muitas threads, poderia se tornar um gargalo de escalabilidade. Embora em Go a operação de canal seja eficiente, o conceito arquitetural permanece.

### 7.3 Trade-off: deadlock vs. concorrência

Reduzir o limite do semáforo para N − 2 ou menos aumentaria a segurança contra deadlock (eliminando ainda mais cenários), mas reduziria a concorrência e aumentaria o tempo total — mais filósofos passariam mais tempo esperando o árbitro. O valor N − 1 é o _mínimo necessário_ para garantir deadlock freedom com o **máximo de concorrência possível**. Trata-se de um ponto ótimo no _trade-off_ entre segurança e desempenho.

### 7.4 Comparação com outras soluções

A solução do árbitro é conceitualmente a mais simples de implementar e verificar, mas apresenta a desvantagem do gargalo centralizado. Espera-se que:

- **Ordenação de recursos** tenha melhor escalabilidade (sem ponto central), porém maior complexidade de implementação.
- **Chandy-Misra** seja o mais justo (distribuído, baseado em fila por aresta), mas com maior overhead de comunicação.
- **Backoff aleatório** seja o mais simples, porém com desempenho imprevisível e sem garantias formais.

---

## 8. Análise de Escalabilidade

### 8.1 Custo computacional vs. N

O custo por rodada escala com **O(N²)** no pior caso devido à serialização imposta pelo semáforo. Duplicar o número de filósofos (de 6 para 12) com o mesmo número de rodadas deve, em teoria, quadruplicar o tempo de execução no pior caso — embora, na prática, a estocasticidade dos tempos de pensamento atenue esse crescimento.

### 8.2 Custo computacional vs. R

Cada rodada adicional adiciona O(N × D) operações de lock/unlock mais 1 segundo de bebida. Como as rodadas são independentes, o tempo total escala linearmente com R. A esperança é que o tempo médio de espera se estabilize conforme R cresce, devido à lei dos grandes números.

### 8.3 Uso de memória vs. N

O consumo de memória é dominado pelo armazenamento da matriz de adjacência (O(N²)). Para um grafo denso com N = 1000, seriam necessários aproximadamente 1 milhão de entradas booleanas (~1 MB), mais 500 mil ponteiros de mutex (~8 MB) — perfeitamente viável. O verdadeiro limitante não é memória, mas sim o tempo de execução O(N²).

---

## 9. Verificação

| Verificação | Resultado |
|-------------|-----------|
| `go build ./...` / `go vet ./...` | OK |
| `go test -race -count=20 ./solucoes/arbitro/` | PASS (3 grafos × 20) |
| Caso 1 real (6 bebidas) | Concluiu, sem deadlock |
| Caso 2 real (6 bebidas) | Concluiu, sem deadlock |
| Caso 3 real (3 bebidas) | Concluiu, sem deadlock |

Os testes com `-race` validam simultaneamente a ausência de _data races_ e, pelo timeout, a ausência de deadlock — as duas condições exigidas pelo enunciado.

---

## 10. Limitações e Riscos

1. **Ponto único de falha**: Se o árbitro falhar (hipoteticamente, em um sistema distribuído real), todo o sistema colapsa. Em uma implementação com memória compartilhada como a nossa, o risco é apenas o deadlock no canal, que é improvável.
2. **Fairness não garantida**: O semáforo não é _fair_ por construção. Filósofos com maior grau tendem a sofrer mais espera (como observado no Caso 2), o que configura um viés contra processos com maiores demandas de recursos.
3. **Overhead de lock sequencial**: Adquirir k locks sequencialmente (em vez de um lock único por conjunto de garrafas) introduz latência proporcional a k. Para grafos com alto grau, isso pode se tornar significativo.
4. **Granularidade dos tempos**: O tempo de pensamento é inteiro (0..n segundos). Com `rand.Intn(n+1)`, há chance não desprezível de pensar 0 s, aumentando a contenção. Uma versão melhor poderia usar `1 + rand.Intn(n)` como piso.

---

## 11. Conclusão

A solução do árbitro resolve corretamente o problema dos Filósofos Bêbados, garantindo ausência de deadlock com uma prova simples e elegante. Os resultados experimentais confirmam sua correção funcional (todos os filósofos completam todas as rodadas sem deadlock) e revelam insights interessantes sobre a relação entre topologia, grau e tempo de espera. A principal fragilidade é o gargalo centralizado e a falta de garantias formais de _fairness_, mas para os cenários testados (N ≤ 12), o desempenho é satisfatório com tempos de execução abaixo de 30 segundos.

---

## 12. Arquivos

| Arquivo | Conteúdo |
|---------|----------|
| `solucoes/arbitro/solver.go` | Implementação completa do solver: árbitro, garrafas, laço dos filósofos |
| `solucoes/arbitro/solver_test.go` | Testes acelerados nos 3 grafos com `-race` |
| `solucoes/arbitro/outputs.md` | Resultados brutos das 3 execuções |
| `solucoes/arbitro/arbitro.md` | Este documento |
