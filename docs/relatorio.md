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
- Não garante *starvation-freedom* formalmente
- Utiliza memória compartilhada (`sync.Mutex` por aresta)

### 2.3 Chandy-Misra

> *Solução implementada por Davi de Oliveira.*

O algoritmo de Chandy-Misra (1984) é uma solução distribuída que resolve o problema por troca de mensagens entre os filósofos vizinhos, sem qualquer coordenador central. A implementação possui **duas camadas**:

1. **Camada dos garfos** (auxiliar): implementa o grafo de precedência H através do "jantar dos filósofos higiênico". Cada aresta possui um garfo que pode estar limpo ou sujo. Para ter precedência sobre uma garrafa, o filósofo precisa estar "comendo" — ou seja, segurando todos os garfos incidentes. Ao começar a comer, ele suja todos os garfos; ao ceder um garfo sujo, ele o entrega limpo.

2. **Camada das garrafas** (recurso real): cada aresta possui uma garrafa e um *request token*. Um filósofo com sede pede as garrafas de que precisa; ao receber um pedido, cede a garrafa a menos que precise dela **e** tenha precedência (garfo limpo) naquela aresta.

A comunicação entre filósofos é feita exclusivamente por canais Go (`chan message`), com quatro tipos de mensagem: `reqFork`, `sendFork`, `reqBottle` e `sendBottle`. Não há estado compartilhado mutável entre goroutines.

**Complexidade:**
- Mensagens por sessão: O(d) — no máximo 4 por vizinho
- Mensagens totais: O(rounds × E)
- Espaço: O(V + E)

**Principais características:**
- Estratégia completamente distribuída
- Garante *starvation-freedom* (Teorema 5, Chandy-Misra, 1984)
- Necessita de duas camadas conceituais (garfos + garrafas)
- Utiliza passagem de mensagens (canais), sem memória compartilhada
- A camada de garfos é essencial: uma implementação inicial com apenas uma camada (garrafas) resultou em deadlock no Caso 3, conforme documentado no desenvolvimento.

### 2.4 Randomized Backoff

> *Solução implementada por Antonio Mozar Braga.*

A solução de *randomized backoff* adota uma abordagem probabilística baseada em tentativa-e-erro com recuo exponencial aleatório. Cada filósofo, ao ficar com sede, tenta adquirir todas as garrafas do subconjunto sorteado utilizando a operação não-bloqueante `TryLock` do pacote `sync` de Go. Se conseguir travar todos os mutexes necessários sem contenção, avança para o estado de bebida. Caso contrário, libera imediatamente os mutexes já adquiridos e espera um intervalo aleatório (até 0,1 segundo) antes de tentar novamente.

```go
for {
    sucesso := true
    var garrafasPegas []*sync.Mutex

    for _, v := range vizinhosEscolhidos {
        m := garrafas[chave(menor, maior)]
        if m.TryLock() {
            garrafasPegas = append(garrafasPegas, m)
        } else {
            sucesso = false
            break
        }
    }

    if sucesso {
        break
    }
    for _, m := range garrafasPegas {
        m.Unlock()
    }
    time.Sleep(time.Duration(r.Float64() * 0.1 * float64(time.Second)))
}
```

Cada filósofo possui uma fonte de aleatoriedade (`math/rand.New`) exclusiva, evitando contenção no gerador global. O tempo de *backoff* é contínuo (0 a 100 ms), dimensionado para ser curto o bastante para não atrasar a simulação, mas longo o suficiente para dessincronizar tentativas concorrentes.

**Complexidade:**
- Tempo (esperado): O(N × D × R) — na prática degrada com contenção
- Tempo (pior caso): Não limitado superiormente (abordagem probabilística)
- Espaço: O(N²)

**Principais características:**
- Estratégia completamente distribuída e não-bloqueante
- Sem garantias determinísticas de progresso (probabilística)
- Implementação simples (~142 linhas)
- RNG privado por filósofo (sem contenção)
- *Backoff* curto fixo (0–100 ms), sem escalonamento exponencial
- Não garante *starvation-freedom* formalmente

---

## 3. Resultados

As execuções foram realizadas em ambiente Linux com Go 1.24, utilizando `time.Second` como unidade de tempo conforme o enunciado. Cada solução foi executada nos três casos com o número de rodadas especificado.

> **Nota:** A solução de Ordenação de Recursos não foi implementada; as células correspondentes estão marcadas com "—".

### Caso 1 — Jantar Clássico (5 nós, grau 2, 6 rodadas)

| Solução | Tempo total (s) | Espera média (s) | Observações |
|---------|-----------------|------------------|-------------|
| Ordenação | — | — | *não implementada* |
| Árbitro | 18,01 | 2,80 | Espera individual entre 2 e 4 s |
| Chandy-Misra | 18,01 | 4,20 | Distribuição uniforme: 3 a 5 s |
| Backoff | 17,42 | 3,56 | Variação individual: 2,40 a 5,15 s |

![Resultados Caso 1 - Árbitro](screenshots/caso1_arbitro.png)

### Caso 2 — Conectividade Baixa (6 nós, graus 2–4, 6 rodadas)

| Solução | Tempo total (s) | Espera média por grau | Observações |
|---------|-----------------|----------------------|-------------|
| Ordenação | — | — | *não implementada* |
| Árbitro | 28,02 | Grau 2: 1,67 s | Desigualdade perceptível: grau 4 esperou 7,01 s |
| | | Grau 3: 4,00 s | Filósofo 4 (grau 2) teve 0 s de espera |
| | | Grau 4: 7,01 s | |
| Chandy-Misra | 24,01 | Grau 2: 3,00 s | Distribuição equilibrada entre graus |
| | | Grau 3: 4,50 s | Maior espera individual: 5,00 s (grau 3) |
| | | Grau 4: 3,00 s | Menor variância entre as soluções |
| Backoff | 22,74 | Grau 2: 1,66 s | Mesmo padrão do Árbitro: graus altos esperam mais |
| | | Grau 3: 3,72 s | Grau 4 esperou 6,02 s (maior individual) |
| | | Grau 4: 6,02 s | Grau 2 teve espera mínima de 0,87 s |

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
| Chandy-Misra | 13,01 | Grau 2: 1,50 s | Distribuição mais equilibrada do conjunto |
| | | Grau 3: 0,40 s | 4 filósofos com espera zero |
| | | Grau 4: 0,67 s | Grau 5 (n=1) esperou 3,00 s |
| | | Grau 5: 3,00 s | Tempo total mais baixo |
| | | Grau 6: 1,00 s | |
| Backoff | 13,43 | Grau 2: 0,92 s | 2 filósofos com espera zero |
| | | Grau 3: 0,77 s | Graus 5 e 6 com maior espera (2,61 e 2,48 s) |
| | | Grau 4: 0,90 s | Padrão consistente: grau elevado = mais espera |
| | | Grau 5: 2,61 s | |
| | | Grau 6: 2,48 s | |

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
| Operação de lock | `Lock` (bloqueante) | `Lock` (bloqueante) | Mensagens (canais) | `TryLock` (não-bloqueante) |
| RNG | — | Global | Privado por filósofo | Privado por filósofo |

### 4.2 Complexidade de Tempo e Espaço

| Métrica | Ordenação | Árbitro | Chandy-Misra | Backoff |
|---------|-----------|---------|--------------|---------|
| Tempo (melhor caso) | Ω(N × D × R) | Ω(N × D × R) | Ω(N × D × R) | Ω(N × D × R) |
| Tempo (pior caso) | O(N × D × R) | O(N² × R) | O(rounds × E) | Não limitado |
| Espaço | O(N²) | O(N²) | O(V + E) | O(N²) |
| Métrica relevante | Locks | Locks | Mensagens | Tentativas |

### 4.3 Resultados Experimentais Agregados

| Métrica | Árbitro | Chandy-Misra | Backoff |
|---------|---------|--------------|---------|
| Tempo total médio (3 casos) | 21,68 s | 18,34 s | 17,86 s |
| Desvio padrão das esperas individuais (Caso 2) | 2,48 s | 0,89 s | 1,93 s |
| Maior espera individual registrada | 7,01 s (grau 4) | 5,00 s (grau 3) | 6,02 s (grau 4) |
| Data-race free | ✓ | ✓ | ✓ |

O Chandy-Misra apresentou o menor desvio padrão nas esperas individuais, confirmando sua superioridade em *fairness*. O Backoff, embora ligeiramente mais rápido em tempo total médio, exibiu o mesmo viés contra graus elevados observado no Árbitro.

### 4.4 Qualidade da Implementação

| Aspecto | Ordenação | Árbitro | Chandy-Misra | Backoff |
|---------|-----------|---------|--------------|---------|
| Linhas de código | — | ~170 | ~410 | ~142 |
| Data-race free | — | ✓ (testado `-race`) | ✓ (testado `-race`) | ✓ (testado `-race`) |
| Testes unitários | — | Sim | Sim | Sim |
| RNG privado | — | Não (global) | Sim (por filósofo) | Sim (por filósofo) |

### 4.5 Cenários Recomendados

- **Ordenação de Recursos**: recomendada para sistemas com topologia fixa e previsível, onde a ordenação prévia das arestas é viável. Oferece bom equilíbrio entre simplicidade e desempenho.

- **Árbitro (Garçom)**: melhor escolha para protótipos e sistemas com poucos processos (N ≤ 12), onde a simplicidade de implementação e verificação é priorizada. Não recomendada para sistemas críticos onde *fairness* é exigida formalmente.

- **Chandy-Misra**: recomendada para sistemas distribuídos e cenários onde *fairness* e ausência de starvation são requisitos formais. Ideal para topologias densas e alta concorrência. A complexidade de implementação é compensada pela robustez.

- **Backoff Aleatório**: adequada para sistemas tolerantes a falhas onde a correção probabilística é aceitável. Útil em cenários com contenção esporádica, mas inadequada para sistemas de tempo real ou críticos. A operação não-bloqueante `TryLock` evita inversão de prioridade, mas não oferece garantias de progresso.

---

## 5. Análise

### 5.1 Deadlock

Nenhuma das soluções implementadas (Árbitro, Chandy-Misra e Backoff) apresentou deadlock durante as execuções. Todas passaram nos testes com `go test -race -count=20` nos três grafos, confirmando a ausência de deadlock e de *data races*. A solução Chandy-Misra, após a introdução da camada de garfos, também foi validada com sucesso — uma versão inicial com apenas uma camada (garrafas) apresentou deadlock no Caso 3, evidenciando a sutileza do problema.

### 5.2 Starvation e Fairness

Os resultados revelam diferenças significativas na distribuição de espera entre as soluções.

**Árbitro**: apresentou o padrão mais desigual. No Caso 2, filósofos com grau 2 tiveram espera média de 1,67 s, enquanto o filósofo de grau 4 esperou 7,01 s. Esse viés contra processos com maior demanda é uma limitação inerente a soluções centralizadas: o semáforo serializa o acesso, e filósofos que precisam de mais recursos competem em desvantagem.

**Backoff**: exibiu o mesmo padrão de viés contra graus elevados. No Caso 2, a espera média foi de 1,66 s para grau 2 e 6,02 s para grau 4. Isso ocorre porque filósofos com mais garrafas têm maior probabilidade de encontrar contenção em pelo menos uma delas, aumentando o número de tentativas fracassadas. No Caso 3, filósofos de grau 5 e 6 tiveram as maiores esperas (2,61 s e 2,48 s), enquanto graus 2 a 4 tiveram médias abaixo de 1 s.

**Chandy-Misra**: apresentou a distribuição mais equilibrada. No Caso 2, a espera variou de 3,00 s (grau 2) a 4,50 s (grau 3), sem correlação clara com o grau. No Caso 3, quatro filósofos tiveram espera zero, e as médias por grau ficaram entre 0,40 s e 1,50 s (exceto grau 5, com n=1, que esperou 3,00 s). Esse resultado confirma a garantia formal de *starvation-freedom* do algoritmo original.

No Caso 3, a topologia mostrou-se mais determinante que o grau individual para todas as soluções: filósofos em posições centrais ou com vizinhos pouco demandantes tiveram espera consistentemente menor. Isso sugere que a métrica de "espera média por grau", embora útil, não captura heterogeneidades locais importantes.

### 5.3 Tempo Total de Execução

O Backoff apresentou o menor tempo total médio (17,86 s), seguido pelo Chandy-Misra (18,34 s) e pelo Árbitro (21,68 s). Essa diferença é explicada pelos paradigmas de sincronização:

- O **Backoff** usa `TryLock` não-bloqueante: um filósofo que encontra contenção desiste imediatamente e espera apenas 0–100 ms antes de tentar de novo. Isso reduz o tempo ocioso, pois ninguém fica bloqueado esperando um mutex ocupado.
- O **Chandy-Misra** usa passagem de mensagens bloqueante: um filósofo que pede uma garrafa ou um garfo fica bloqueado até receber a resposta, o que pode acumular latência.
- O **Árbitro** usa mutexes bloqueantes combinados com o semáforo: filósofos bloqueados no semáforo ou em mutexes individuais acumulam tempo de espera.

### 5.4 Trade-offs Arquiteturais

A comparação entre as soluções revela um *trade-off* fundamental entre **simplicidade** e **robustez**:

- O **Árbitro** é a solução mais simples (170 linhas, 1 camada conceitual), mas introduz um ponto único de contenção e não garante *fairness*.
- O **Chandy-Misra** é o mais robusto (garantias formais, distribuído, sem *starvation*), porém significativamente mais complexo (410 linhas, 2 camadas conceituais, gestão explícita de aposentadoria).
- O **Backoff** ocupa uma posição intermediária: mais simples que o Chandy-Misra (142 linhas), mas sem garantias determinísticas. O uso de `TryLock` elimina espera bloqueante, mas o recuo probabilístico pode, em teoria, levar a starvation.
- A **Ordenação de Recursos** (não implementada) ocuparia uma posição similar ao Árbitro em simplicidade, com a vantagem de garantir *starvation-freedom*, desde que a ordem total de recursos seja conhecida e estável.

O valor **N − 1** no Árbitro representa o ponto ótimo no *trade-off* entre prevenção de deadlock e maximização da concorrência: valores menores aumentariam a serialização desnecessariamente; valores maiores (N) permitiriam deadlock.

---

## 6. Conclusão

Este trabalho implementou e analisou quatro soluções de sincronização para o problema do Bar dos Filósofos, uma generalização do clássico problema dos filósofos famintos. Cada solução emprega uma estratégia distinta para evitar deadlock: ordenação global de recursos, controle centralizado por semáforo, coordenação distribuída por troca de mensagens e recuo probabilístico com `TryLock`.

Os resultados experimentais confirmam que todas as três soluções implementadas (Árbitro, Chandy-Misra e Backoff) atendem ao requisito fundamental de ausência de deadlock, validado por testes exaustivos com detector de *data races*. A análise comparativa revelou diferenças marcantes:

- O **Chandy-Misra** destacou-se como a única solução com distribuição verdadeiramente equilibrada das esperas, confirmando a garantia formal de *starvation-freedom* do algoritmo original. Seu tempo total médio (18,34 s) é competitivo, e o desvio padrão das esperas (0,89 s no Caso 2) é o menor do conjunto.
- O **Backoff** provou ser a solução mais rápida em tempo total médio (17,86 s), graças à operação não-bloqueante `TryLock`. No entanto, seu padrão de espera replica o viés contra graus elevados observado no Árbitro, e sua natureza probabilística não oferece garantias formais de progresso.
- O **Árbitro**, apesar de ser o mais lento (21,68 s em média) e o mais desigual na distribuição de esperas, continua sendo a melhor escolha para protótipos e sistemas pequenos onde a simplicidade de implementação é priorizada sobre *fairness*.

A principal lição do desenvolvimento foi a necessidade de rigor na implementação de algoritmos concorrentes: uma versão inicial do Chandy-Misra com apenas uma camada (garrafas) apresentou deadlock no caso de maior conectividade, ilustrando como a generalização do problema (subconjunto arbitrário de recursos) introduz sutilezas que não aparecem no jantar clássico.

Para trabalhos futuros, sugerimos: (a) a implementação da solução de Ordenação de Recursos; (b) a execução com um número maior de rodadas para estabilizar as médias; (c) a avaliação em grafos de maior escala (N ≥ 100) para estressar a escalabilidade; e (d) a introdução de métricas adicionais como *throughput* e utilização de recursos.

---

## 7. Como Reproduzir

Após descompactar o arquivo ZIP recebido, acesse o diretório raiz do projeto e utilize os comandos abaixo.

**Pré-requisito:** Go 1.21 ou superior instalado no sistema.

```sh
# Entrar no diretório do projeto
cd concurrent-programming-work/

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

Também é possível executar diretamente sem o Makefile:

```sh
go run ./cmd/runner -solucao=arbitro -grafo=data/caso1_jantar_5.txt -rodadas=6
```

---

## Referências

- Chandy, K. M., & Misra, J. (1984). The Drinking Philosophers Problem. *ACM Transactions on Programming Languages and Systems (TOPLAS)*, 6(4), 632–646.
- Dijkstra, E. W. (1971). Hierarchical ordering of sequential processes. *Acta Informatica*, 1(2), 115–138.
- Coffman, E. G., Elphick, M. J., & Shoshani, A. (1971). System deadlocks. *ACM Computing Surveys (CSUR)*, 3(2), 67–78.
- Tanenbaum, A. S., & Bos, H. (2015). *Modern Operating Systems* (4th ed.). Pearson.
