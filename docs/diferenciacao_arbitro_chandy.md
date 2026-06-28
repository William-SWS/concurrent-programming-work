# Diferenciação das Implementações: Árbitro vs. Chandy-Misra

## 1. Filosofia de Coordenação

| Aspecto | Árbitro (Garçom) | Chandy-Misra |
|---------|------------------|--------------|
| **Tipo** | Centralizado | Distribuído |
| **Controle** | Um coordenador global serializa o acesso | Coordenação ponto-a-ponto via troca de mensagens |
| **Metáfora** | Garçom de bar que libera mesas | Protocolo de cavalheiros trocando objetos |

---

## 2. Mecanismo de Exclusão Mútua

### Árbitro
- Usa **`sync.Mutex`** por aresta — lock/unlock direto no recurso compartilhado
- O mutex é uma variável de memória compartilhada; o filósofo dá Lock() e ninguém mais consegue usar aquela garrafa até o Unlock()
- O arbítro central (**`chan struct{}`** de capacidade N−1) funciona como semáforo contador: `acquire()` = `sem <- token`, `release()` = `<-sem`

### Chandy-Misra
- Usa **canais Go (`chan message`)** para troca de mensagens — zero memória compartilhada mutável entre goroutines
- Cada filósofo tem seu próprio `inbox` (`chan message`). Para adquirir uma garrafa, ele envia uma `message{reqBottle, ...}` ao vizinho, que responde com `message{sendBottle, ...}`
- A garrafa nunca é "travada" — ela é um token que está sempre de um lado da aresta. A posse é transferida por mensagem (`sendBottle`)

**Conclusão**: Árbitro usa **memória compartilhada** (mutex); Chandy-Misra usa **passagem de mensagens** (canais). Isso reflete os dois grandes paradigmas de concorrência.

---

## 3. Prevenção de Deadlock

| Aspecto | Árbitro | Chandy-Misra |
|---------|---------|--------------|
| **Estratégia** | Limitar concorrência a N−1 | Grafo de precedência H acíclico |
| **Como evita** | Semáforo garante que 1 filósofo sempre fique de fora | Garfos "sujos/limpos" + distribuição inicial no menor ID mantêm H acíclico |
| **Camada extra?** | Não | Sim: **garfos** (jantar higiênico) são uma camada auxiliar sobre as garrafas |
| **Prova** | Pigeonhole: N−1 slots, N filósofos → 1 sempre fora | Chandy-Misra 1984 prova que H permanece acíclico |

### Por que Chandy-Misra precisa de 2 camadas?
O problema de "beber" permite que um filósofo peça **apenas um subconjunto** dos recursos, diferentemente do jantar clássico (que precisa de 2 garfos fixos). Isso pode criar **ciclos no grafo de precedência H** se não houver controle. A camada dos **garfos** (fork state) implementa o "jantar dos filósofos higiênico": para ter precedência sobre uma garrafa, o filósofo precisa estar "comendo" (holding all forks). A distribuição inicial (garfo sujo sempre no filósofo de menor ID) garante H acíclico.

O árbitro simplesmente **evita o problema** limitando quantos competem — sacrificando concorrência máxima por simplicidade.

---

## 4. Estrutura de Dados

| Componente | Árbitro | Chandy-Misra |
|------------|---------|--------------|
| **Estado da aresta** | `sync.Mutex` (global, indexado por `[2]int`) | `edge` struct local por filósofo: `{fork, reqf, dirty, bot, reqb, need}` |
| **Comunicação** | Acesso direto ao mutex | Mensagens tipadas via `inbox` channel |
| **Árbitro** | `chan struct{}` bufferizado N−1 | Inexistente (não há coordenador) |
| **Timer** | `time.Sleep()` síncrono | `time.Timer` orientado a eventos (`select/case`) |
| **RNG** | `math/rand` global (compartilhado) | `rand.New(rand.NewSource(...))` — um RNG **privado por filósofo** |

### Por que RNG privado no Chandy-Misra?
Cada filósofo no Chandy-Misra tem seu próprio `*rand.Rand`:

```go
rng: rand.New(rand.NewSource(int64(id)*7919 + 1)),
```

Isso elimina contenção no gerador de números aleatórios. No Árbitro, `rand.Intn(n+1)` usa o `rand` global, que tem um mutex interno — **contenção escondida** que pode degradar desempenho com muitos filósofos.

---

## 5. Modelo de Execução (Goroutines)

| Aspecto | Árbitro | Chandy-Misra |
|---------|---------|--------------|
| **Goroutines** | N (um filósofo cada) | N (um filósofo cada) |
| **Loop** | Sequencial: `for drink { think → acquire locks → drink → unlock → release }` | Orientado a eventos: `select { case msg: handle; case timer: transition }` |
| **Bloqueio** | `arb.acquire()` + `mutex.Lock()` — bloqueia a goroutine | `p.in <- msg` (envio) + `<-timerC` (timer) — bloqueia a goroutine |
| **Aquisição** | Lock direto (imediato, sem negociação) | Troca de mensagens: request → response (pode levar várias iterações) |

No Árbitro, o fluxo de cada rodada é **imperativo e passo-a-passo**: pensar, pedir árbitro, lock, beber, unlock, release. Não há tratamento de eventos externos durante a execução — o filósofo só interage com os mutexes e o channel.

No Chandy-Misra, o fluxo é **orientado a eventos**: o filósofo reage a mensagens e timers. As funções `handle()` e `react()` (com seu `step()` em ponto fixo) processam assincronamente as transições de estado. Isso é mais complexo porém mais flexível — o filósofo pode processar pedidos de vizinhos mesmo enquanto está esperando recursos.

---

## 6. Complexidade das Transições de Estado

### Árbitro — Transições Simples

```
TRANQUILO  ──(timer)──►  COM_SEDE.acquire()
                              │
                              ▼
                          LOCK garrafas
                              │
                              ▼
                          BEBENDO ──(timer)──► UNLOCK + release()
                              │
                              └──────────────────► TRANQUILO
```

### Chandy-Misra — Transições com Duas Camadas

```
tranquilo ──(timer)──► comSede ──(D1)──► hungry
                                            │
                                     ──(R1f)──► pede garfo ao vizinho
                                            │
                                     ◇──(sendFork)──◇ holdsAllForks?
                                            │
                                        eating (suja garfos)
                                            │
                                   ──(R1b)──► pede garrafa ao vizinho
                                            │
                                     ◇──(sendBottle)──◇ hasAllNeeded?
                                            │
                                        bebendo ──(timer)──► finishDrinking
                                            │
                                        tranquilo (D2: eating→thinking)
```

O Chandy-Misra precisa de **6 estados conceituais** (tranquilo, comSede, hungry, eating, bebendo e as combinações com retired), enquanto o Árbitro usa apenas os **3 estados do enunciado** (tranquilo, comSede, bebendo).

---

## 7. Troca de Mensagens

### Árbitro
- **0 mensagens entre filósofos**. Toda comunicação é com o árbitro central (channel) ou com os mutexes.
- O canal do árbitro é único e compartilhado por todos.

### Chandy-Misra
- **4 tipos de mensagem** entre filósofos:
  - `reqFork` / `sendFork` — controle da camada de precedência (garfos)
  - `reqBottle` / `sendBottle` — controle do recurso real (garrafas)
- Cada aresta pode gerar múltiplas mensagens por rodada (request + response).
- Mensagens são assíncronas — ficam no `inbox` até o destino processá-las no próximo `select`.

---

## 8. Tratamento de "Aposentadoria" (Retirement)

### Árbitro
- **Inexistente**: todas as goroutines executam exatamente `rounds` iterações e terminam juntas via `sync.WaitGroup`. Ao terminar, simplesmente saem do `for`.

### Chandy-Misra
- **Explícito e complexo**: quando `p.drinks >= p.sim.rounds`, o filósofo "se aposenta" (`retire()`):
  - Marca `p.retired = true`
  - Zera `need` em todas as arestas (não precisa mais de garrafas)
  - Para o timer
  - Decrementa `sim.completed` (WaitGroup)
  - **Continua executando**: mesmo aposentado, o filósofo ainda precisa **servir os vizinhos** — entregando garfos e garrafas que eles pedirem
  - Só termina quando `sim.stop` é fechado (após todos terem completado)

Essa diferença é fundamental: no Chandy-Misra, um filósofo que já bebeu todas as rodadas **não pode simplesmente morrer**, pois os vizinhos ainda podem precisar de recursos que ele detém.

---

## 9. Granularidade Temporal

| Aspecto | Árbitro | Chandy-Misra |
|---------|---------|--------------|
| **Timer pensamento** | `time.Sleep()` — bloqueia a goroutine inteira | `time.NewTimer()` — permite processar mensagens enquanto espera |
| **Timer bebida** | `time.Sleep(1s)` — bloqueia tudo | `time.NewTimer(Unidade)` — permite atender pedidos enquanto bebe |
| **Unidade de tempo** | Fixo em `time.Second` (hardcoded) | Parametrizável via variável `var Unidade = time.Second` |

### Implicação
No **Árbitro**, enquanto um filósofo está bebendo (1 segundo de `time.Sleep`), ele **não processa absolutamente nada**. Se um vizinho tentar travar o mutex de uma garrafa que ele detém, ficará bloqueado até o `Sleep` terminar.

No **Chandy-Misra**, o `select` permite que, mesmo durante a bebida, o filósofo atenda mensagens dos vizinhos. A linha `case m := <-p.in:` está sempre ativa. Isso significa que um vizinho pode **enfileirar um pedido de garrafa** e o filósofo bêbado pode **responder imediatamente** (cedendo a garrafa se não tiver precedência). Maior capacidade de resposta.

---

## 10. Fairness

| Aspecto | Árbitro | Chandy-Misra |
|---------|---------|--------------|
| **Garantia formal** | Nenhuma — semáforo não é fair | Chandy-Misra garante ausência de starvation (1984) |
| **Na prática** | Filósofos de maior grau sofrem mais espera (evidenciado nos resultados) | Distribuição mais equilibrada por usar fila por aresta |
| **Mecanismo de fairness** | Canal FIFO (Go garante ~fairness entre produtores) | Pedidos servidos em ordem de chegada (request token + mailbox) |

---

## 11. Resumo Comparativo

| Dimensão | Árbitro | Chandy-Misra |
|----------|---------|--------------|
| Paradigma | Memória compartilhada (mutex) | Passagem de mensagens (canais) |
| Coordenação | Centralizada | Distribuída |
| Complexidade de código | ~170 linhas, 1 arquivo | ~410 linhas, 2 arquivos |
| Camadas conceituais | 1 (garrafas) | 2 (garfos + garrafas) |
| Estados internos | 3 | 6+ |
| Risco de data race | Baixo (mutex protegido) | Zero (sem estado compartilhado) |
| Deadlock freedom | Sim (N−1) | Sim (H acíclico) |
| Starvation freedom | Não garantido | Garantido |
| Resposta durante bebida | Não (Sleep bloqueante) | Sim (select orientado a eventos) |
| Aposentadoria | Simples (termina loop) | Complexa (precisa servir após terminar) |
| RNG | Global (com contenção) | Privado por filósofo |
| Unidade temporal | Hardcoded | Parametrizável |

---

## 12. Conclusão

O **Árbitro** é a solução **mais simples**: fácil de entender, implementar e verificar. Sacrifica concorrência máxima (N−1) e fairness por simplicidade. Ideal para cenários com poucos processos e onde a prova de deadlock freedom precisa ser óbvia.

O **Chandy-Misra** é a solução **mais robusta**: distribuída, livre de starvation por construção, com resposta a eventos mesmo durante a seção crítica. A contrapartida é a complexidade — duas camadas (garfos para precedência + garrafas para recurso), máquina de estados com ponto fixo, e gerenciamento explícito de aposentadoria.

Em termos de **engenharia de software**, o Árbitro é mais fácil de manter e estender; o Chandy-Misra é mais correto formalmente e mais eficiente em cenários de alta contenção. A escolha entre eles reflete o clássico _trade-off_ entre simplicidade e robustez em sistemas concorrentes.
