# Análise dos Resultados — Solução 2: Árbitro (Garçom)

## 1. Descrição do Algoritmo

A solução do árbitro central, também conhecida como solução do garçom (_waiter_), introduz um coordenador global responsável por serializar o acesso aos recursos compartilhados. O princípio é simples, porém elegantemente eficaz: o árbitro mantém um semáforo de capacidade **N − 1**, onde N é o número total de filósofos. Antes de qualquer filósofo tentar adquirir garrafas, ele deve obter uma autorização do árbitro. Caso o número máximo de filósofos simultâneos já tenha sido atingido, o filósoho é bloqueado até que outro termine de beber e libere sua vaga.

### 1.1 Estruturas de Dados

- **Árbitro**: implementado como um canal Go bufferizado com capacidade N − 1. O envio de um token (`chan <- struct{}{}`) representa a aquisição de permissão; o recebimento (`<-chan`) representa a liberação.
- **Garrafas**: cada aresta do grafo é protegida por um `sync.Mutex` independente, armazenado em um dicionário indexado pelo par ordenado `{menorID, maiorID}`.
- **Filósofos**: executados concorrentemente como goroutines, cada um seguindo o ciclo `TRANQUILO → COM_SEDE → BEBENDO`.

### 1.2 Pseudocódigo

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

---

## 2. Prova de Correção

### 2.1 Livre de Deadlock

**Teorema**: A solução do árbitro com limite N − 1 é livre de deadlock.

**Demonstração**: Considere N filósofos e um semáforo de capacidade N − 1. No máximo N − 1 filósofos podem estar simultaneamente no estado COM_SEDE ou BEBENDO. Portanto, em qualquer instante, **pelo menos um filósofo está no estado TRANQUILO**.

Suponha, por absurdo, que ocorra um deadlock. Isso significa que todos os filósofos ativos estão retendo algumas garrafas e aguardando por outras — uma espera circular. No entanto, o filósofo que está TRANQUILO não retém nenhuma garrafa. Quando ele acordar e tentar entrar em COM_SEDE, encontrará o semáforo com pelo menos uma vaga disponível (pois no máximo N − 1 estão ocupados). Como ele não retém recurso algum, ele conseguirá adquirir todas as garrafas de que precisa — mesmo que todos os N − 1 ativos estejam segurando uma garrafa cada, ainda sobra pelo menos uma garrafa livre. Logo, o progresso é sempre possível, e o deadlock é impossível.

### 2.2 Livre de Starvation (Fairness)

A solução não garante _starvation-freedom_ estrita. O semáforo do árbitro é do tipo _não-fair_ (por padrão, canais Go são _fair_ apenas na seleção entre goroutines prontas, mas não há fila de prioridade explícita). Na prática, para os casos testados, todos os filósofos conseguiram beber o número esperado de rodadas, indicando ausência de starvation no cenário experimental. Contudo, em teoria, um filósofo poderia ser repetidamente preterido se o árbitro sempre favorecesse outros — situação que se torna improvável com tempos de pensamento aleatórios.

### 2.3 Exclusão Mútua

Cada garrafa (aresta) é protegida por um `sync.Mutex`. Apenas um filósofo por vez pode segurar uma dada garrafa, garantindo que nenhum par de vizinhos beba simultaneamente usando o mesmo recurso compartilhado. Isso está de acordo com a especificação do problema: garrafas são recursos não-compartilháveis.

---

## 3. Análise de Complexidade

### 3.1 Complexidade de Tempo

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

No pior cenário, todos os N − 1 ativos competem intensamente. Cada acquire do último filósofo pode esperar que todos os N − 2 anteriores completem seu ciclo (incluindo 1 segundo de bebida cada). Isso introduz um fator multiplicativo O(N) de espera sobre as operações internas, resultando em O(N² × R). O tempo total observado de 18,01 s para N = 5, 28,02 s para N = 6 e 19,01 s para N = 12 (com R = 3) é consistente com esta análise: o caso 3, com mais nós, teve menos rodadas (3 vs. 6), o que explica o tempo total comparável.

### 3.2 Complexidade de Espaço

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

## 4. Interpretação dos Resultados Experimentais

### 4.1 Caso 1 — Jantar Clássico (Ciclo de 5 Nós, Grau 2)

| Parâmetro | Valor |
|-----------|-------|
| Tempo total | 18,01 s |
| Espera média | 2,80 s |
| Todos os filósofos | Grau 2 |

**Análise**: No grafo cíclico clássico, todos os 5 filósofos possuem exatamente 2 vizinhos. O limite do árbitro é 4 ativos simultâneos. A espera média de 2,80 s mostra que houve contenção moderada, mas todos completaram as 6 rodadas dentro do esperado. Os tempos de espera individuais variaram de 2 s a 4 s, indicando alternância razoável entre os filósofos — não houve monopolização do árbitro.

**Fato relevante**: Neste caso específico, poderíamos ter 4 filósofos ativos e 1 tranquilo. Como todos têm grau 2, cada filósofo precisa de 2 garrafas. Mesmo que 4 ativos seguem 4 das 5 garrafas, o filósofo tranquilo, ao acordar, consegue as 2 garrafas restantes. O deadlock é estruturalmente impossível.

### 4.2 Caso 2 — Conectividade Baixa (6 Nós, Graus 2–4)

| Parâmetro | Valor |
|-----------|-------|
| Tempo total | 28,02 s |
| Espera média grau 2 | 1,67 s |
| Espera média grau 3 | 4,00 s |
| Espera média grau 4 | 7,01 s |

**Análise**: Este caso revela um **padrão de desigualdade por grau**. Filósofos com grau 2 (nós 0, 1, 4) tiveram espera média de apenas 1,67 s, enquanto o filósofo de grau 4 (nó 5) esperou 7,01 s — mais de 4 vezes maior. A explicação é direta: filósofos de maior grau precisam adquirir mais garrafas (k entre 2 e grau), aumentando a probabilidade de contenção. Além disso, ao segurar mais recursos, eles bloqueiam mais vizinhos, que por sua vez demoram mais para liberar.

**Fato relevante**: O filósofo 4 (grau 2) teve **0 segundos de espera** — ele conseguiu todas as garrafas imediatamente em todas as rodadas. Isso ocorre porque seus vizinhos (de maior grau) demoravam mais para adquirir recursos, deixando as garrafas do nó 4 frequentemente disponíveis.

### 4.3 Caso 3 — Conectividade Alta (12 Nós, Graus 2–6)

| Parâmetro | Valor |
|-----------|-------|
| Tempo total | 19,01 s |
| Espera média grau 2 | 2,00 s |
| Espera média grau 3 | 0,40 s |
| Espera média grau 4 | 2,34 s |
| Espera média grau 5 | 2,00 s |
| Espera média grau 6 | 1,00 s |

**Análise**: Com 12 filósofos e apenas 3 rodadas cada, o tempo total foi de 19,01 s — coerente com 36 bebidas de 1 s cada (36 s de bebida pura) mais tempos de pensamento, porém com sobreposição substancial (até 11 simultâneos). A espera média por grau não segue uma relação monotônica: o grau 3 teve a menor espera (0,40 s), enquanto o grau 2 teve 2,00 s e o grau 4 teve 2,34 s. Isso ocorre porque a topologia do grafo (quem é vizinho de quem) importa tanto quanto o grau individual. Um filósofo de grau 3 cercado por vizinhos de baixo grau sofre menos contenção que um de grau 2 cercado por vizinhos de alto grau.

**Fato relevante**: Os filósofos 1, 5 e 8 (todos de grau 3) tiveram **0 segundos de espera**. Isso sugere que eles estão posicionados em regiões do grafo com baixa contenção — possivelmente conectados a filósofos que passam mais tempo pensando ou que competem por garrafas diferentes.

---

## 5. Observações Interessantes

### 5.1 A topologia importa mais que o grau

Os resultados do Caso 3 demonstram que a **posição no grafo** é tão determinante quanto o grau para o tempo de espera. O nó 5 (grau 3, espera 0 s) e o nó 3 (grau 4, espera 2 s) exemplificam que a distribuição dos vizinhos e seus respectivos padrões de escolha de garrafas influenciam fortemente a contenção. A métrica de "espera média por grau" agrega demais e esconde heterogeneidades locais.

### 5.2 Efeito de serialização do árbitro

O semáforo de capacidade N − 1 resolve o deadlock, mas introduz um **gargalo de serialização**. Quando muitos filósofos ficam com sede simultaneamente, o canal do árbitro funciona como um _ponto de contenção único_ (_single point of contention_). Todos os filósofos disputam o mesmo canal, o que, em sistemas com muitas threads, poderia se tornar um gargalo de escalabilidade. Embora em Go a operação de canal seja eficiente, o conceito arquitetural permanece.

### 5.3 Trade-off: deadlock vs. concorrência

Reduzir o limite do semáforo para N − 2 ou menos aumentaria a segurança contra deadlock (eliminando ainda mais cenários), mas reduziria a concorrência e aumentaria o tempo total — mais filósofos passariam mais tempo esperando o árbitro. O valor N − 1 é o _mínimo necessário_ para garantir deadlock freedom com o **máximo de concorrência possível**. Trata-se de um ponto ótimo no _trade-off_ entre segurança e desempenho.

### 5.4 Comparação com outras soluções

A solução do árbitro é conceitualmente a mais simples de implementar e verificar, mas apresenta a desvantagem do gargalo centralizado. Espera-se que:

- **Ordenação de recursos** tenha melhor escalabilidade (sem ponto central), porém maior complexidade de implementação.
- **Chandy-Misra** seja o mais justo (distribuído, baseado em fila por aresta), mas com maior overhead de comunicação.
- **Backoff aleatório** seja o mais simples, porém com desempenho imprevisível e sem garantias formais.

---

## 6. Análise de Escalabilidade

### 6.1 Custo computacional vs. N

O custo por rodada escala com **O(N²)** no pior caso devido à serialização imposta pelo semáforo. Duplicar o número de filósofos (de 6 para 12) com o mesmo número de rodadas deve, em teoria, quadruplicar o tempo de execução no pior caso — embora, na prática, a estocasticidade dos tempos de pensamento atenue esse crescimento.

### 6.2 Custo computacional vs. R

Cada rodada adicional adiciona O(N × D) operações de lock/unlock mais 1 segundo de bebida. Como as rodadas são independentes, o tempo total escala linearmente com R. A esperança é que o tempo médio de espera se estabilize conforme R cresce, devido à lei dos grandes números.

### 6.3 Uso de memória vs. N

O consumo de memória é dominado pelo armazenamento da matriz de adjacência (O(N²)). Para um grafo denso com N = 1000, seriam necessários aproximadamente 1 milhão de entradas booleanas (~1 MB), mais 500 mil ponteiros de mutex (~8 MB) — perfeitamente viável. O verdadeiro limitante não é memória, mas sim o tempo de execução O(N²).

---

## 7. Limitações e Riscos

1. **Ponto único de falha**: Se o árbitro falhar (hipoteticamente, em um sistema distribuído real), todo o sistema colapsa. Em uma implementação com memória compartilhada como a nossa, o risco é apenas o deadlock no canal, que é improvável.
2. **Fairness não garantida**: O semáforo não é _fair_ por construção. Filósofos com maior grau tendem a sofrer mais espera (como observado no Caso 2), o que configura um viés contra processos com maiores demandas de recursos.
3. **Overhead de lock sequencial**: Adquirir k locks sequencialmente (em vez de um lock único por conjunto de garrafas) introduz latência proporcional a k. Para grafos com alto grau, isso pode se tornar significativo.
4. **Granularidade dos tempos**: O tempo de pensamento é inteiro (0..n segundos). Com `rand.Intn(n+1)`, há chance não desprezível de pensar 0 s, aumentando a contenção. Uma versão melhor poderia usar `1 + rand.Intn(n)` como piso.

---

## 8. Conclusão

A solução do árbitro resolve corretamente o problema dos Filósofos Bêbados, garantindo ausência de deadlock com uma prova simples e elegante. Os resultados experimentais confirmam sua correção funcional (todos os filósofos completam todas as rodadas sem deadlock) e revelam insights interessantes sobre a relação entre topologia, grau e tempo de espera. A principal fragilidade é o gargalo centralizado e a falta de garantias formais de _fairness_, mas para os cenários testados (N ≤ 12), o desempenho é satisfatório com tempos de execução abaixo de 30 segundos.
