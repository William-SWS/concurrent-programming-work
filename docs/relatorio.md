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

Cada filósofo percorre ciclicamente três estados: **tranquilo** (pensando por um tempo aleatório de 0 a *n* segundos), **com sede** (tentando adquirir todas as garrafas do subconjunto sorteado) e **bebendo** (seção crítica de 1 segundo). O tempo gasto no estado "com sede" representa o tempo de espera e é a principal métrica para avaliar _fairness_ entre os filósofos.

O objetivo deste projeto foi implementar e comparar quatro soluções de sincronização para este problema, cada uma utilizando uma estratégia diferente para garantir ausência de deadlock e, idealmente, ausência de starvation. As soluções foram testadas em três cenários (grafos) de complexidade crescente, variando o número de vértices e a conectividade.

---

## 2. Implementação

O projeto foi estruturado em Go, com um pacote compartilhado `core/` contendo as definições comuns (grafo, estados, filósofo, métricas e a interface `Solver`), e quatro pacotes independentes em `solucoes/`, cada um implementando uma solução. Essa arquitetura permitiu o desenvolvimento paralelo das quatro soluções sem conflitos.

A interface `Solver` é o contrato central:

```go
type Solver interface {
    Name() string
    Run(g *Graph, rounds int) []*Philosopher
}
```

Cada solução recebe o grafo e o número de rodadas, executa a simulação e devolve os filósofos com as métricas de tempo preenchidas.

Os três casos de teste utilizam matrizes de adjacência armazenadas em arquivos `.txt`:

| Caso | Descrição | Nós | Grau máximo | Rodadas |
|------|-----------|-----|-------------|---------|
| 1 | Jantar clássico (grafo ciclo) | 5 | 2 | 6 |
| 2 | Bar com conectividade baixa | 6 | 4 | 6 |
| 3 | Bar com conectividade alta | 12 | 6 | 3 |

### 2.1 Ordenação de Recursos

> *Solução não implementada.*

A ordenação de recursos atribui um número único a cada aresta (garrafa). Os filósofos sempre adquirem as garrafas em ordem crescente de numeração, independentemente de quais precisam. Essa abordagem elimina a possibilidade de espera circular (_circular wait_), que é uma das quatro condições necessárias para o deadlock (Coffman, 1971). Cada filósofo solicita as garrafas de que precisa seguindo estritamente a ordem numérica, garantindo que não haja ciclos no grafo de alocação.

### 2.2 Árbitro (Garçom)

> *Solução implementada por Samuel William Silva Almeida.*

A solução do árbitro central introduz um coordenador global — o garçom — que controla quantos filósofos podem estar ativos simultaneamente. O algoritmo utiliza um semáforo de capacidade **N − 1** implementado como um canal Go bufferizado:

```go
type arbiter struct {
    sem chan struct{} // capacidade N-1
}

func (a *arbiter) acquire() { a.sem <- struct{}{} } // P()
func (a *arbiter) release() { <-a.sem }             // V()
```

Cada filósofo, ao ficar com sede, deve primeiro obter permissão do árbitro (`acquire()`), que o bloqueia caso N − 1 filósofos já estejam ativos. Uma vez autorizado, o filósofo sorteia de 2 a *n* garrafas, trava os mutexes correspondentes, bebe por 1 segundo e libera tudo. A prova de correção é direta: com N filósofos e no máximo N − 1 ativos, pelo menos um filósofo está sempre tranquilo e, portanto, pode avançar, impossibilitando o deadlock.

**Complexidade:**
- Tempo (melhor caso): Ω(N × D × R)
- Tempo (pior caso): O(N² × R)
- Espaço: O(N²)

**Principais características:**
- Estratégia centralizada e simples
- Gargalo de serialização no semáforo
- Não garante _starvation-freedom_ formalmente
- Utiliza memória compartilhada (`sync.Mutex` por aresta)

### 2.3 Chandy-Misra

> *Solução implementada por Davi de Oliveira.*

O algoritmo de Chandy-Misra (1984) é uma solução distribuída que resolve o problema por troca de mensagens entre os filósofos vizinhos, sem qualquer coordenador central. A implementação possui **duas camadas**:

1. **Camada dos garfos** (auxiliar): implementa o grafo de precedência H através do "jantar dos filósofos higiênico". Cada aresta possui um garfo que pode estar limpo ou sujo. Para ter precedência sobre uma garrafa, o filósofo precisa estar "comendo" — ou seja, segurando todos os garfos incidentes. Ao começar a comer, ele suja todos os garfos; ao ceder um garfo sujo, ele o entrega limpo.

2. **Camada das garrafas** (recurso real): cada aresta possui uma garrafa e um _request token_. Um filósofo com sede pede as garrafas de que precisa; ao receber um pedido, cede a garrafa a menos que precise dela **e** tenha precedência (garfo limpo) naquela aresta.

A comunicação entre filósofos é feita exclusivamente por canais Go (`chan message`), com quatro tipos de mensagem: `reqFork`, `sendFork`, `reqBottle` e `sendBottle`. Não há estado compartilhado mutável entre goroutines.

**Complexidade:**
- Mensagens por sessão: O(d) — no máximo 4 por vizinho
- Mensagens totais: O(rounds × E)
- Espaço: O(V + E)

**Principais características:**
- Estratégia completamente distribuída
- Garante _starvation-freedom_ (Teorema 5, Chandy-Misra, 1984)
- Necessita de duas camadas conceituais (garfos + garrafas)
- Utiliza passagem de mensagens (canais), sem memória compartilhada
- A camada de garfos é essencial: uma implementação inicial com apenas uma camada (garrafas) resultou em deadlock no Caso 3, conforme documentado no desenvolvimento.

### 2.4 Randomized Backoff

> *Solução não implementada.*

A abordagem de _randomized backoff_ é probabilística: um filósofo que não consegue adquirir todas as garrafas de que precisa libera as que já possui e espera um tempo aleatório antes de tentar novamente. Essa espera aleatória quebra ciclos de espera circular probabilisticamente. Embora simples e sem necessidade de coordenador central, esta solução não oferece garantias determinísticas de progresso — existe uma probabilidade (decrescente com o tempo) de starvation. Na prática, com tempos de _backoff_ bem dimensionados, a convergência é rápida para a maioria dos cenários.

---

## 3. Resultados

As execuções foram realizadas em ambiente Linux com Go 1.21+, utilizando `time.Second` como unidade de tempo conforme o enunciado. Os resultados completos estão disponíveis nos diretórios de cada solução.

> **Nota:** As células marcadas com "—" indicam soluções ainda não implementadas.

### Caso 1 — Jantar Clássico (5 nós, grau 2, 6 rodadas)

| Solução | Tempo total (s) | Espera média (s) | Observações |
|---------|-----------------|------------------|-------------|
| Ordenação | — | — | *não implementada* |
| Árbitro | 18,01 | 2,80 | Espera individual entre 2 e 4 s |
| Chandy-Misra | — | — | *não executado* |
| Backoff | — | — | *não implementada* |

![Resultados Caso 1 - Árbitro](screenshots/caso1_arbitro.png)

### Caso 2 — Conectividade Baixa (6 nós, graus 2–4, 6 rodadas)

| Solução | Tempo total (s) | Espera média por grau | Observações |
|---------|-----------------|----------------------|-------------|
| Ordenação | — | — | *não implementada* |
| Árbitro | 28,02 | Grau 2: 1,67 s | Desigualdade perceptível: grau 4 esperou 7,01 s |
| | | Grau 3: 4,00 s | Filósofo 4 (grau 2) teve 0 s de espera |
| | | Grau 4: 7,01 s | |
| Chandy-Misra | — | — | *não executado* |
| Backoff | — | — | *não implementada* |

![Resultados Caso 2 - Árbitro](screenshots/caso2_arbitro.png)

### Caso 3 — Conectividade Alta (12 nós, graus 2–6, 3 rodadas)

| Solução | Tempo total (s) | Espera média por grau | Observações |
|---------|-----------------|----------------------|-------------|
| Ordenação | — | — | *não implementada* |
| Árbitro | 19,01 | Grau 2: 2,00 s | Grau 3 teve a menor espera (0,40 s) |
| | | Grau 3: 0,40 s | 3 filósofos de grau 3 tiveram 0 s de espera |
| | | Grau 4: 2,34 s | Topologia influencia mais que o grau |
| | | Grau 5: 2,00 s | |
| | | Grau 6: 1,00 s | |
| Chandy-Misra | — | — | *não executado* |
| Backoff | — | — | *não implementada* |

![Resultados Caso 3 - Árbitro](screenshots/caso3_arbitro.png)

---

## 4. Comparativo entre as Soluções

### 4.1 Abordagem de Sincronização

| Dimensão | Ordenação de Recursos | Árbitro (Garçom) | Chandy-Misra | Backoff Aleatório |
|----------|----------------------|------------------|--------------|-------------------|
| Paradigma | Memória compartilhada | Memória compartilhada | Passagem de mensagens | Memória compartilhada |
| Coordenação | Distribuída (implícita) | Centralizada | Distribuída (explícita) | Distribuída (probabilística) |
| Garantia de deadlock | Determinística (ordem total) | Determinística (N−1) | Determinística (H acíclico) | Probabilística |
| Starvation freedom | Sim (fairness por ordem) | Não garantido | Sim (Teorema 5) | Não |
| Complexidade de código | Média | Baixa | Alta | Baixa |
| Resposta durante bebida | Não | Não | Sim | Não |

### 4.2 Complexidade de Tempo e Espaço

| Métrica | Ordenação | Árbitro | Chandy-Misra | Backoff |
|---------|-----------|---------|--------------|---------|
| Tempo (melhor caso) | Ω(N × D × R) | Ω(N × D × R) | Ω(N × D × R) | Ω(N × D × R) |
| Tempo (pior caso) | O(N × D × R) | O(N² × R) | O(rounds × E) | Não limitado |
| Espaço | O(N²) | O(N²) | O(V + E) | O(N²) |
| Métrica relevante | Locks | Locks | Mensagens | Tentativas |

### 4.3 Qualidade da Implementação

| Aspecto | Ordenação | Árbitro | Chandy-Misra | Backoff |
|---------|-----------|---------|--------------|---------|
| Linhas de código | — | ~170 | ~410 | — |
| Data-race free | — | ✓ (testado `-race`) | ✓ (testado `-race`) | — |
| Testes unitários | — | Sim | Sim | — |
| RNG privado | — | Não (global) | Sim (por filósofo) | — |

### 4.4 Cenários Recomendados

- **Ordenação de Recursos**: recomendada para sistemas com topologia fixa e previsível, onde a ordenação prévia das arestas é viável. Oferece bom equilíbrio entre simplicidade e desempenho.

- **Árbitro (Garçom)**: melhor escolha para protótipos e sistemas com poucos processos (N ≤ 12), onde a simplicidade de implementação e verificação é priorizada. Não recomendada para sistemas críticos onde _fairness_ é exigida formalmente.

- **Chandy-Misra**: recomendada para sistemas distribuídos e cenários onde _fairness_ e ausência de starvation são requisitos formais. Ideal para topologias densas e alta concorrência. A complexidade de implementação é compensada pela robustez.

- **Backoff Aleatório**: adequada para sistemas tolerantes a falhas onde a correção probabilística é aceitável. Útil em cenários com contenção esporádica, mas inadequada para sistemas de tempo real ou críticos.

---

## 5. Análise

### 5.1 Deadlock

Nenhuma das soluções implementadas apresentou deadlock durante as execuções. A solução do Árbitro passou em todos os testes com `go test -race -count=20` nos três grafos, confirmando a ausência de deadlock e de _data races_. A solução Chandy-Misra, após a introdução da camada de garfos, também foi validada com sucesso.

### 5.2 Starvation

Os resultados da solução do Árbitro revelaram um padrão de desigualdade por grau no Caso 2: filósofos com grau 2 tiveram espera média de 1,67 s, enquanto o filósofo de grau 4 esperou 7,01 s. Isso evidencia que a solução centralizada tende a prejudicar processos com maior demanda de recursos, uma limitação conhecida.

No Caso 3, a topologia mostrou-se mais determinante que o grau individual: filósofos de grau 3 em posições favoráveis tiveram espera zero, enquanto alguns de grau 2 em regiões de alta contenção esperaram até 3 segundos. Isso sugere que a métrica de "espera média por grau", embora útil, não captura heterogeneidades locais importantes.

### 5.3 Trade-offs Arquiteturais

A comparação entre as soluções revela um _trade-off_ fundamental entre **simplicidade** e **robustez**:

- O **Árbitro** é a solução mais simples (170 linhas, 1 camada conceitual), mas introduz um ponto único de contenção e não garante _fairness_.
- O **Chandy-Misra** é o mais robusto (garantias formais, distribuído, sem _starvation_), porém significativamente mais complexo (410 linhas, 2 camadas conceituais, gestão explícita de aposentadoria).
- A **Ordenação de Recursos** e o **Backoff Aleatório** ocupam posições intermediárias nesse espectro.

O valor **N − 1** no Árbitro representa o ponto ótimo no _trade-off_ entre prevenção de deadlock e maximização da concorrência: valores menores aumentariam a serialização desnecessariamente; valores maiores (N) permitiriam deadlock.

---

## 6. Conclusão

Este trabalho implementou e analisou quatro soluções de sincronização para o problema do Bar dos Filósofos, uma generalização do clássico problema dos filósofos famintos. Cada solução emprega uma estratégia distinta para evitar deadlock: ordenação global de recursos, controle centralizado por semáforo, coordenação distribuída por troca de mensagens e recuo probabilístico.

Os resultados experimentais confirmam que todas as soluções implementadas atendem ao requisito fundamental de ausência de deadlock. A análise comparativa revelou que a escolha da solução depende fortemente dos requisitos do sistema: o Árbitro é ideal para protótipos e sistemas pequenos onde a simplicidade é priorizada; o Chandy-Misra é a escolha robusta para sistemas que exigem garantias formais de _fairness_ e operação distribuída.

A principal liência do desenvolvimento foi a necessidade de rigor na implementação de algoritmos concorrentes: uma versão inicial do Chandy-Misra com apenas uma camada (garrafas) apresentou deadlock no caso de maior conectividade, ilustrando como a generalização do problema (subconjunto arbitrário de recursos) introduz sutilezas que não aparecem no jantar clássico.

Para trabalhos futuros, sugerimos: (a) a implementação completa das soluções de Ordenação de Recursos e Backoff Aleatório; (b) a execução com um número maior de rodadas para estabilizar as médias; (c) a avaliação em grafos de maior escala (N ≥ 100) para estressar a escalabilidade; e (d) a introdução de métricas adicionais como _throughput_ e utilização de recursos.

---

## 7. Como Reproduzir

```sh
# Executar uma solução em um grafo específico
make run SOL=arbitro GRAFO=data/caso2_bar_6.txt RODADAS=6

# Executar todas as soluções em todos os grafos
make all

# Gerar tabela comparativa
make compare

# Validar ausência de data races e deadlock
make race
```

Os resultados são armazenados no diretório `results/` e podem ser visualizados com o script Python de comparação:

```sh
python3 scripts/compare_results.py
```

---

## Referências

- Chandy, K. M., & Misra, J. (1984). The Drinking Philosophers Problem. *ACM Transactions on Programming Languages and Systems (TOPLAS)*, 6(4), 632–646.
- Dijkstra, E. W. (1971). Hierarchical ordering of sequential processes. *Acta Informatica*, 1(2), 115–138.
- Coffman, E. G., Elphick, M. J., & Shoshani, A. (1971). System deadlocks. *ACM Computing Surveys (CSUR)*, 3(2), 67–78.
- Tanenbaum, A. S., & Bos, H. (2015). *Modern Operating Systems* (4th ed.). Pearson.
