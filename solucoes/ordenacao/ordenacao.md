# Solução 1 — Ordenação de Recursos

**Responsável:** Antonio Bezerra de Morais Neto

**Pacote:** `solucoes/ordenacao/`

**Referência:** Solução clássica por ordenação hierárquica de recursos (_hierarchical resource allocation_, Havender, 1968), que atribui uma ordem total a todos os recursos do sistema e exige que cada processo os adquira nessa ordem, eliminando a condição de espera circular necessária para o deadlock (Coffman, Elphick & Shoshani, 1971).

---

## 1. Como rodar

Todos os comandos a partir da raiz do projeto (`concurrent-programming-work/`).

### Execução real

```sh
# caso 1 — jantar clássico (5 nós, 6 bebidas por filósofo)
go run ./cmd/runner -solucao=ordenacao -grafo=data/caso1_jantar_5.txt -rodadas=6

# caso 2 — bar, baixa conectividade (6 nós, 6 bebidas)
go run ./cmd/runner -solucao=ordenacao -grafo=data/caso2_bar_6.txt -rodadas=6

# caso 3 — bar, alta conectividade (12 nós, 3 bebidas)
go run ./cmd/runner -solucao=ordenacao -grafo=data/caso3_bar_12.txt -rodadas=3
```

Ou via Makefile:

```sh
make run SOL=ordenacao GRAFO=data/caso2_bar_6.txt RODADAS=6
```

### Verificação de data race

```sh
go test -race -count=20 ./solucoes/ordenacao/
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

A solução por ordenação de recursos — também conhecida como _hierarchical resource allocation_ ou _lock ordering_ — atribui um número único a cada aresta (garrafa) do grafo. A numeração é arbitrária, mas global e imutável durante toda a execução. Cada filósofo, ao precisar de um subconjunto de garrafas, adquire-as **estritamente em ordem crescente de numeração**, independentemente de sua identidade ou posição no grafo.

A prova de correção é direta: se todos os processos adquirem recursos seguindo a mesma ordem global, o grafo de alocação de recursos nunca pode conter ciclos. Suponha, por absurdo, que exista um ciclo P₁ → R₁ → P₂ → R₂ → ... → Pₖ → Rₖ → P₁, onde P₁ detém Rₖ e espera por R₁. Pela ordem global, R₁ < R₂ < ... < Rₖ. Mas P₁ detém Rₖ e espera R₁, o que implicaria Rₖ < R₁ — contradição. Portanto, o ciclo não pode existir, e o deadlock é impossível.

### 3.2 Estruturas de Dados

- **Garrafas**: cada aresta do grafo é protegida por um `sync.Mutex`, armazenado em um mapa indexado pelo par ordenado `{menorID, maiorID}`.
- **Ordem global**: um mapa auxiliar `map[[2]int]int` associa cada aresta a um número de ordem único, atribuído na sequência em que `g.Edges()` as enumera.
- **RNG privado**: cada filósofo possui seu próprio `*rand.Rand` criado com `rand.New(rand.NewSource(...))`, eliminando contenção no gerador global de números aleatórios.
- **Filósofos**: executados concorrentemente como goroutines, cada um seguindo o ciclo `TRANQUILO → COM_SEDE → BEBENDO`.

### 3.3 Pseudocódigo

```
# Atribuição global: numerar cada aresta com um ID único
para cada aresta e em g.Edges():
    order[e] = id_unico

para cada filósofo i, em paralelo:
    rng := nova fonte aleatória privada
    para cada rodada r em 1..rounds:
        # Estado TRANQUILO
        pensar por tempo aleatório 0..grau(i) segundos
        registrar tempo em Metrics.Tranquilo

        # Estado COM_SEDE
        escolher k garrafas aleatórias       # k ∈ [2, grau(i)]
        ordernar garrafas escolhidas por order[] crescente
        para cada garrafa escolhida (ordem crescente):
            mutex[garrafa].Lock()            # adquire na ordem
        registrar tempo em Metrics.ComSede

        # Estado BEBENDO
        dormir 1 segundo
        registrar tempo em Metrics.Bebendo

        # Liberação (ordem não importa)
        para cada garrafa escolhida:
            mutex[garrafa].Unlock()
```

### 3.4 Prova de Correção

#### 3.4.1 Livre de Deadlock

**Teorema**: A solução por ordenação de recursos é livre de deadlock para qualquer grafo.

**Demonstração**: Seja `f: E → ℕ` uma função injetora que atribui um número natural único a cada aresta do grafo. Suponha, por absurdo, que haja um deadlock. Nessa situação, existe um ciclo de espera C = P₁, R₁, P₂, R₂, ..., Pₖ, Rₖ, P₁ onde cada Pⱼ detém Rⱼ₋₁ e aguarda Rⱼ. Pela regra de aquisição, se Pⱼ aguarda Rⱼ, então `f(Rⱼ) > f(Rⱼ₋₁)` (pois Pⱼ está tentando adquirir garrafas em ordem crescente e ainda não conseguiu Rⱼ). Percorrendo o ciclo, obtemos `f(R₁) > f(Rₖ) > f(Rₖ₋₁) > ... > f(R₁)`, uma contradição. Logo, o ciclo não pode existir e o deadlock é impossível.

**Corolário**: Diferentemente do Árbitro (que impõe N − 1 ativos), a ordenação de recursos elimina a condição de _circular wait_ sem reduzir a concorrência máxima. Todos os N filósofos podem estar ativos simultaneamente, desde que sigam a ordem global.

#### 3.4.2 Starvation Freedom

A solução garante _starvation-freedom_ na aquisição individual de garrafas: como todos os filósofos adquirem recursos na mesma ordem, não há inversão de prioridade. No entanto, o _starvation_ global (um filósofo nunca consegue todas as garrafas simultaneamente) ainda é possível se houver alta contenção em uma aresta específica. O cenário é menos grave que no Árbitro, porém não há garantia formal equivalente ao Teorema 5 de Chandy-Misra.

#### 3.4.3 Exclusão Mútua

Cada garrafa (aresta) é protegida por um `sync.Mutex`. Apenas um filósofo por vez pode segurar uma dada garrafa, garantindo que nenhum par de vizinhos beba simultaneamente usando o mesmo recurso.

---

## 4. A armadilha: lock ordering com shuffle prévio

### O problema

Uma sutileza importante na implementação é a **ordem das operações**: primeiro sorteia-se o subconjunto de garrafas, depois ordena-se esse subconjunto globalmente antes de adquiri-las. Se a ordenação fosse feita antes do sorteio — ou se o sorteio já produzisse uma lista inerentemente ordenada — o comportamento seria diferente.

O código faz:

```go
chosen := chooseNeighbors(r, neighbors)         # subconjunto aleatório
sort.Slice(chosen, func(i, j int) bool {         # ordena globalmente
    return order[lockOrder(p.ID, chosen[i])] < order[lockOrder(p.ID, chosen[j])]
})
for _, nb := range chosen {
    bottles[lockOrder(p.ID, nb)].Lock()
}
```

A função `chooseNeighbors` já retorna um slice, mas ele está na ordem do shuffle, não na ordem global. O `sort.Slice` subsequente reordena para a sequência crescente de numeração. Se esse `sort` fosse omitido, a solução degeneraria para aquisição em ordem arbitrária — equivalente a locks sem ordenação, que **não** garante deadlock freedom.

### Impacto na starvation

A ordenação global introduz um viés sutil: filósofos que disputam garrafas de **baixo número** tendem a travá-las primeiro (porque todos começam por elas), enquanto garrafas de **alto número** ficam menos disputadas. Isso pode fazer com que filósofos cujo subconjunto inclui apenas garrafas de alto número tenham espera menor que a média — um efeito observado no Caso 3, onde filósofos de grau 6 tiveram espera zero.

### Lição

A correção da solução depende criticamente da **ordem de aquisição**. Um erro comum é assumir que "cada filósofo adquire recursos em ordem crescente" sem garantir que o subconjunto esteja ordenado antes do loop de locks. A implementação deve separar explicitamente as etapas de (a) escolha do subconjunto, (b) ordenação global e (c) aquisição sequencial.

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
| Escolha aleatória de garrafas | O(D) | Shuffle parcial (Fisher-Yates) |
| Ordenação do subconjunto | O(k log k) ⊆ O(D log D) | `sort.Slice` sobre k ≤ D elementos |
| Aquisição de garrafas | O(k) ⊆ O(D) | Lock sequencial na ordem global |
| Bebida (sleep) | O(1) | 1 segundo fixo |
| Liberação de garrafas | O(k) ⊆ O(D) | Unlock sequencial (qualquer ordem) |

**Custo por filósofo por rodada**: O(D log D) — dominado pela ordenação do subconjunto.

**Custo total sem contenção (melhor caso)**: Ω(N × D log D × R)

No melhor cenário, cada filósofo adquire garrafas sem esperar por locks já travados.

**Custo total com contenção (pior caso)**: O(N × D log D × R)

Diferentemente do Árbitro (que tem fator quadrático O(N²) devido ao semáforo), a ordenação de recursos **não introduz serialização adicional**. A contenção apenas alonga o tempo de espera nos locks individuais, mas não multiplica o trabalho por N. O pior caso real é limitado pelo tempo de bebida (1 s por rodada) somado à espera nos locks.

### 5.2 Complexidade de Espaço

| Componente | Complexidade | Detalhamento |
|------------|-------------|--------------|
| Matriz de adjacência | **O(N²)** | `[][]bool` carregada do arquivo |
| Mapa de mutexes das garrafas | **O(E)** ⊆ O(N²) | Um ponteiro `*sync.Mutex` por aresta |
| Mapa de ordem global | **O(E)** ⊆ O(N²) | Um inteiro por aresta |
| Slice de filósofos | **O(N)** | N structs `Philosopher` |
| RNGs privados | **O(N)** | N instâncias de `*rand.Rand` |
| Stack das goroutines | **O(N)** | Uma goroutine por filósofo (~8 KB cada) |
| **Total** | **O(N²)** | Dominado pela matriz de adjacência |

---

## 6. Interpretação dos Resultados Experimentais

### 6.1 Caso 1 — Jantar Clássico (Ciclo de 5 Nós, Grau 2)

```
go run ./cmd/runner -solucao=ordenacao -grafo=data/caso1_jantar_5.txt -rodadas=6
```

| Parâmetro | Valor |
|-----------|-------|
| Tempo total | 19,01 s |
| Espera média | 3,60 s |
| Todos os filósofos | Grau 2 |

| Filósofo | Grau | Tranquilo (s) | Com Sede (s) | Bebendo (s) | Bebidas |
|----------|------|---------------|--------------|-------------|---------|
| 0 | 2 | 2,00 | 3,00 | 6,00 | 6 |
| 1 | 2 | 4,00 | 5,01 | 6,00 | 6 |
| 2 | 2 | 4,00 | 2,00 | 6,00 | 6 |
| 3 | 2 | 6,00 | 3,00 | 6,00 | 6 |
| 4 | 2 | 8,00 | 5,00 | 6,01 | 6 |

**Análise**: No grafo cíclico, os 5 filósofos possuem grau 2. A espera média de 3,60 s é comparável à do Árbitro (2,80 s). Os tempos individuais variam de 2 s a 5 s. Como todos os filósofos disputam as mesmas 5 garrafas (organizadas em ordem global), a contenção é moderada. A ordenação evita deadlock, mas não elimina a competição pelos recursos — dois filósofos adjacentes ainda não podem beber simultaneamente se compartilham uma garrafa.

### 6.2 Caso 2 — Conectividade Baixa (6 Nós, Graus 2–4)

```
go run ./cmd/runner -solucao=ordenacao -grafo=data/caso2_bar_6.txt -rodadas=6
```

| Parâmetro | Valor |
|-----------|-------|
| Tempo total | 22,01 s |

| Filósofo | Grau | Tranquilo (s) | Com Sede (s) | Bebendo (s) | Bebidas |
|----------|------|---------------|--------------|-------------|---------|
| 0 | 2 | 5,00 | 4,00 | 6,00 | 6 |
| 1 | 2 | 5,00 | 3,00 | 6,00 | 6 |
| 2 | 3 | 12,00 | 2,00 | 6,00 | 6 |
| 3 | 3 | 10,00 | 3,01 | 6,00 | 6 |
| 4 | 2 | 2,00 | 1,00 | 6,00 | 6 |
| 5 | 4 | 14,00 | 2,00 | 6,01 | 6 |

**Espera média por grau:**

| Grau | Média (s) | n |
|------|-----------|---|
| 2 | 2,67 | 3 |
| 3 | 2,50 | 2 |
| 4 | 2,00 | 1 |

**Análise**: Este caso revela um comportamento **oposto ao observado no Árbitro e no Backoff**. Enquanto aquelas soluções penalizavam filósofos de maior grau, a Ordenação de Recursos apresenta espera **decrescente** com o grau: grau 4 (2,00 s) esperou menos que grau 2 (2,67 s). 

A explicação está na ordem global: filósofos de maior grau tendem a participar de mais arestas, e algumas dessas arestas têm numeração baixa — adquiridas primeiro por todos. Isso significa que um filósofo de grau 4 frequentemente consegue garrafas de baixo número antes que seus vizinhos de menor grau as requisitem. Em outras palavras, a ordenação global nivela a competição: o grau deixa de ser o principal fator determinante da espera.

Outra observação importante: as esperas individuais são muito mais homogêneas que no Árbitro. O maior tempo individual foi 4,00 s (grau 2) contra 7,01 s (grau 4) no Árbitro. Isso sugere melhor distribuição de _fairness_.

### 6.3 Caso 3 — Conectividade Alta (12 Nós, Graus 2–6)

```
go run ./cmd/runner -solucao=ordenacao -grafo=data/caso3_bar_12.txt -rodadas=3
```

| Parâmetro | Valor |
|-----------|-------|
| Tempo total | 17,01 s |

| Filósofo | Grau | Tranquilo (s) | Com Sede (s) | Bebendo (s) | Bebidas |
|----------|------|---------------|--------------|-------------|---------|
| 0 | 3 | 3,00 | 1,00 | 3,00 | 3 |
| 1 | 3 | 5,00 | 1,00 | 3,00 | 3 |
| 2 | 2 | 4,00 | 1,00 | 3,00 | 3 |
| 3 | 4 | 5,00 | 0,00 | 3,00 | 3 |
| 4 | 4 | 8,00 | 2,00 | 3,00 | 3 |
| 5 | 3 | 4,00 | 0,00 | 3,00 | 3 |
| 6 | 5 | 12,00 | 1,00 | 3,00 | 3 |
| 7 | 4 | 2,00 | 0,00 | 3,00 | 3 |
| 8 | 3 | 4,00 | 1,00 | 3,00 | 3 |
| 9 | 6 | 14,00 | 0,00 | 3,00 | 3 |
| 10 | 2 | 3,00 | 1,00 | 3,00 | 3 |
| 11 | 3 | 3,00 | 1,00 | 3,00 | 3 |

**Espera média por grau:**

| Grau | Média (s) | n |
|------|-----------|---|
| 2 | 1,00 | 2 |
| 3 | 0,80 | 5 |
| 4 | 0,67 | 3 |
| 5 | 1,00 | 1 |
| 6 | 0,00 | 1 |

**Análise**: Com 12 filósofos e apenas 3 rodadas, o tempo total foi de 17,01 s — o mais baixo entre as soluções implementadas para este caso (Árbitro: 19,01 s; Backoff: 13,43 s; Chandy-Misra: 13,01 s). A espera média de 0,75 s é a segunda menor (atrás do Chandy-Misra com 0,67 s).

O filósofo de grau 6 (nó 9) teve **espera zero** — em todas as rodadas, conseguiu todas as garrafas imediatamente. Isso ocorre porque, com a ordenação global, o nó 9 está ligado a arestas de numeração favorável: como todos começam adquirindo garrafas de baixo número, e o nó 9 depende majoritariamente de garrafas de numeração intermediária, ele raramente encontra contenção.

**Starvation outlier**: O filósofo 4 (grau 4) esperou 2,00 s — 3 vezes a média do grupo (0,67 s). Embora seja um outlier, o valor absoluto de 2 s para 3 rodadas é baixo (menor que o maior outlier de qualquer outra solução no mesmo caso). Isso demonstra que, mesmo com uma distribuição geral excelente, a ordenação global pode ocasionalmente prejudicar filósofos cujo subconjunto de garrafas coincide sistematicamente com as mais disputadas.

---

## 7. Observações e Discussão

### 7.1 Ordenação global inverte a correlação grau-espera

Um dos resultados mais interessantes deste trabalho é que a Ordenação de Recursos **inverte** a correlação entre grau e tempo de espera observada nas outras soluções. Enquanto Árbitro e Backoff penalizam graus altos, a Ordenação tende a beneficiá-los — não por _fairness_ intrínseca, mas porque a ordem global redistribui a contenção de forma diferente. Um filósofo de grau 4 pode ter espera menor que um de grau 2, dependendo de quais arestas cada um precisa.

### 7.2 Ausência de ponto único de contenção

A principal vantagem arquitetural sobre o Árbitro é a **ausência de um gargalo centralizado**. Todos os N filósofos podem estar ativos simultaneamente; a única serialização é nos mutexes individuais das garrafas. Isso elimina a multiplicação O(N²) do Árbitro e torna a solução inerentemente mais escalável.

### 7.3 Custo da ordenação

O `sort.Slice` com O(k log k) por rodada é barato para k ≤ D ≤ 6 nos grafos testados, mas tornaria-se o gargalo para grafos com grau muito alto (D ≥ 100). Nesses casos, uma abordagem alternativa seria pré-calcular a ordem de aquisição para cada filósofo, eliminando a ordenação em tempo de execução.

### 7.4 Comparação com outras soluções

- **vs. Árbitro**: mais escalável (sem semáforo N − 1), porém requer ordenação prévia das arestas. Distribuição de espera mais uniforme entre graus.
- **vs. Chandy-Misra**: mais simples (~108 linhas contra ~410), mas sem garantia formal de _starvation-freedom_. Competitivo em tempo total.
- **vs. Backoff**: determinístico (garante deadlock freedom), enquanto o Backoff é probabilístico. Implementação similar em complexidade.

---

## 8. Análise de Escalabilidade

### 8.1 Custo computacional vs. N

O custo por rodada escala com **O(N × D log D)**. Diferentemente do Árbitro, não há fator quadrático devido a serialização centralizada. Para grafos densos (D ≈ N), o custo se aproxima de O(N² log N), mas esse é o custo obrigatório para percorrer as arestas — não uma ineficiência introduzida pela solução.

### 8.2 Custo computacional vs. R

Cada rodada adicional adiciona O(D log D) mais 1 segundo de bebida. O tempo total escala linearmente com R. O tempo médio de espera tende a se estabilizar para R grande, convergindo para um valor dependente da topologia e da ordem global.

### 8.3 Uso de memória vs. N

O(N²), dominado pela matriz de adjacência. O mapa de ordem global adiciona O(E) inteiros, que no pior caso (grafo completo) é O(N²) — mas cada inteiro são apenas 8 bytes, contra 8 bytes do ponteiro de mutex mais o custo do mutex em si. O overhead adicional da ordenação é desprezível.

---

## 9. Verificação

| Verificação | Resultado |
|-------------|-----------|
| `go build ./...` / `go vet ./...` | OK |
| `go test -race -count=20 ./solucoes/ordenacao/` | PASS |
| Caso 1 real (6 bebidas) | Concluiu, sem deadlock |
| Caso 2 real (6 bebidas) | Concluiu, sem deadlock |
| Caso 3 real (3 bebidas) | Concluiu, sem deadlock |

Os testes com `-race` validam simultaneamente a ausência de _data races_ e, pelo timeout, a ausência de deadlock.

---

## 10. Limitações e Riscos

1. **Requer topologia estática**: a ordem global deve ser definida antes da execução e não pode mudar dinamicamente. Em sistemas onde as arestas são adicionadas ou removidas em tempo de execução, a reordenação é custosa e propensa a erros.
2. **Ordenação é O(k log k)**: para grafos com grau muito alto (D ≥ 100), ordenar o subconjunto a cada rodada torna-se o gargalo computacional. Uma otimização possível é pré-calcular a ordem para cada par possível.
3. **Starvation não eliminado formalmente**: embora a ordenação previna deadlock deterministicamente, não há garantia de _starvation-freedom_ equivalente ao Teorema 5 de Chandy-Misra. Um outlier de 3× a média foi observado no Caso 3.
4. **Granularidade dos tempos**: o tempo de pensamento é inteiro (0..n segundos), podendo ser 0 segundos. Isso aumenta a contenção em relação a tempos contínuos (usados no Backoff).
5. **Inversão da correlação grau-espera**: embora seja uma característica interessante, a inversão observada (maior grau, menor espera) não é uma propriedade desejável universalmente — depende da aplicação se espera-se que processos com mais recursos tenham prioridade ou não.

---

## 11. Conclusão

A solução por Ordenação de Recursos oferece uma abordagem determinística e elegantemente simples para o problema dos Filósofos Bêbados. Ao atribuir uma ordem global às arestas e exigir aquisição nessa ordem, elimina-se a condição de espera circular sem reduzir a concorrência máxima e sem introduzir coordenadores centrais.

Os resultados experimentais confirmam sua correção funcional (deadlock-free em todos os 3 grafos) e revelam um padrão de espera qualitativamente diferente das demais soluções: a ordenação global **inverte** a correlação entre grau e tempo de espera, beneficiando filósofos de maior grau — ao contrário do Árbitro e do Backoff, que os penalizam. O tempo total de execução é competitivo (17,01 s no Caso 3, o mais baixo entre as 3 soluções determinísticas).

Com 108 linhas de código e sem necessidade de coordenador central, a Ordenação de Recursos representa um ponto ótimo no espectro simplicidade-robustez: mais simples que o Chandy-Misra, mais justa que o Árbitro e mais previsível que o Backoff. Sua principal limitação é a exigência de topologia estática e a ausência de garantia formal de _starvation-freedom_.

---

## 12. Arquivos

| Arquivo | Conteúdo |
|---------|----------|
| `solucoes/ordenacao/solver.go` | Implementação completa do solver: ordem global, mutexes, laço dos filósofos |
| `solucoes/ordenacao/solver_test.go` | Teste acelerado com `-race` e timeout |
| `solucoes/ordenacao/ordenacao.md` | Este documento |
