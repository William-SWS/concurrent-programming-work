# Solução 3 — Chandy-Misra (Drinking Philosophers)

**Responsável:** Davi de Oliveira

**Pacote:** `solucoes/chandy_misra/`

**Referência:** K. M. Chandy & J. Misra, *The Drinking Philosophers Problem*, ACM TOPLAS, 1984.

---

## 1. Como rodar

Todos os comandos a partir da raiz do projeto (`concurrent-programming-work/`).

### Execução real (1 segundo por unidade de tempo, ~minutos)

```sh
# caso 1 — jantar clássico (5 nós, 6 bebidas por filósofo)
go run ./cmd/runner -solucao=chandy_misra -grafo=data/caso1_jantar_5.txt -rodadas=6

# caso 2 — bar, baixa conectividade (6 nós, 6 bebidas)
go run ./cmd/runner -solucao=chandy_misra -grafo=data/caso2_bar_6.txt -rodadas=6

# caso 3 — bar, alta conectividade (12 nós, 3 bebidas)
go run ./cmd/runner -solucao=chandy_misra -grafo=data/caso3_bar_12.txt -rodadas=3
```

Ou via Makefile:

```sh
make run SOL=chandy_misra GRAFO=data/caso2_bar_6.txt RODADAS=6
```

### Verificação rápida (tempo acelerado)

```sh
# roda os 3 grafos com tempo reduzido; falha em timeout (deadlock) ou se algum
# filósofo não beber o número exigido de vezes
go test ./solucoes/chandy_misra/

# o mesmo, com detector de data race e 20 repetições para estressar o
# escalonamento (evidência forte de ausência de deadlock e de race)
go test -race -count=20 ./solucoes/chandy_misra/
```

### Saída

Ao final imprime, por filósofo: tempo em cada estado (`tranquilo`, `com sede`,
`bebendo`), número de bebidas; e ainda o tempo total e a **espera média (com
sede) por grau** (o número que mostra ausência de starvation).

```
# solucao=Chandy-Misra
# tempo_total=18.01s
id  grau  tranquilo  com_sede  bebendo  bebidas
0   2     7.00       4.00      6.00     6
...
# espera media (com sede) por grau:
# grau=2 media=4.00s (n=5)
```

---

## 2. O problema

Filósofos (vértices do grafo) bebem de garrafas (arestas) compartilhadas com os
vizinhos. Cada filósofo passa por 3 estados em sequência:

1. **tranquilo** - ocioso por um tempo aleatório de 0 a *n* segundos (*n* = grau).
2. **com sede** - sorteia de 2 a *n* garrafas e tenta obter **todas**; o tempo
   nesse estado é o tempo de espera.
3. **bebendo** - segura todas as garrafas por 1 segundo e volta a tranquilo.

O grafo é lido de uma matriz de adjacência (`data/*.txt`). É exigido: **sem
deadlock** e **sem starvation** (esperas médias equilibradas entre filósofos de
mesmo grau).

---

## 3. O algoritmo (duas camadas)

A solução do Chandy-Misra resolve conflitos por um **grafo de precedência H**:
quando dois filósofos disputam um recurso, vence o de maior precedência, e a
precedência muda ao longo do tempo para ninguém ficar sempre perdendo (= sem
starvation). A abordagem do artigo (a referência) está em manter H **acíclico** de forma
**distribuída**, só com mensagens, e por isso a solução tem **duas camadas**:

### Camada das garrafas (recurso real, o que o enunciado descreve)

Cada aresta tem uma garrafa e um *request token*. Regras:

- **com sede**: para cada garrafa necessária que não detém, envia um pedido ao
  vizinho (só pode pedir quem tem o *request token*).
- ao **receber um pedido**: cede a garrafa, **a menos que** precise dela **e**
  (esteja bebendo **ou** tenha precedência naquela aresta).
- ao **terminar de beber**: libera as garrafas usadas.
- pedidos concorrentes são servidos **na ordem de chegada** (garantido pelo
  *request token* + a fila de mensagens do filósofo).

### Camada dos garfos (auxiliar, implementa a precedência H)

Cada aresta também tem um **garfo**, que pode estar **limpo** ou **sujo**, e um
*request token* próprio. Os garfos implementam o "jantar dos filósofos
higiênico":

- para ter precedência, o filósofo fica **faminto** e precisa segurar **todos**
  os garfos incidentes (estado **comendo**).
- ao começar a comer, **suja** todos os garfos.
- um filósofo que **não** está comendo **cede** um garfo **sujo** quando pedido;
  retém os **limpos** (tem precedência neles).
- garfo recebido chega **limpo**.

O elo entre as camadas é a regra de cessão da garrafa: *"retenho a garrafa se
preciso dela E tenho o garfo daquela aresta"*. Ou seja, **o garfo (a
precedência) é quem decide o desempate sobre a garrafa**.

### Por que isso não trava nem causa starvation

- A orientação inicial dos garfos é **acíclica** (garfo sujo sempre com o
  filósofo de **menor id**), e toda mudança em H preserva a aciclicidade —
  porque um garfo só é sujo quando o filósofo está **comendo**, e para comer ele
  segura **todos** os garfos (todas as arestas apontam para ele, impossível
  fechar ciclo).
- Como H é sempre acíclico, sempre existe um filósofo de maior precedência que
  consegue comer e, comendo, puxa todas as garrafas de que precisa e bebe em
  tempo finito. Por indução, **todo filósofo com sede bebe em tempo finito** ->
  sem deadlock e sem starvation.

---

## 4. A armadilha: por que uma camada só NÃO basta

> Essa seção documenta um erro real cometido durante o desenvolvimento e como
> foi diagnosticado e corrigido.

### A versão errada

A primeira implementação tinha **uma camada só**: cada aresta era uma garrafa
limpa/suja (mapeando "cheia/vazia" do enunciado direto para "limpa/suja" do
artigo), sem a camada separada de garfos. A ideia era: a própria garrafa carrega
a precedência.

Isso **parece** equivalente, e de fato funciona no *jantar* dos filósofos
(quando o filósofo sempre precisa de **todas** as suas garrafas). Passou nos
testes acelerados e nos casos 1 e 2 reais.

### Por que estava errada

No *bar* dos filósofos, o filósofo bebe de um **subconjunto** sorteado das
garrafas. A garantia de aciclicidade de H depende de o filósofo segurar **todos**
os recursos incidentes ao "sujar" suas arestas. Bebendo de um subconjunto, ele
suja **apenas algumas** arestas sem segurar todas -> **H pode virar cíclico** ->
espera circular -> **deadlock**.

Em grafos esparsos (casos 1 e 2) a probabilidade é baixa e não apareceu. No caso
3 (alta conectividade, até 6 arestas por nó) apareceu:

```
fatal error: all goroutines are asleep - deadlock!

goroutine 1 [sync.WaitGroup.Wait]:
...
github.com/.../chandy_misra.(*Solver).Run(...)
        .../chandy_misra/solver.go:74
```

Todas as goroutines bloqueadas em `select` e a `main` parada no `WaitGroup`:
clássico estado de deadlock (um conjunto de filósofos com sede esperando
garrafas uns dos outros, em ciclo, sem ninguém poder ceder).

O teste acelerado tinha passado por **sorte de escalonamento** (com tempos de 2
ms o entrelaçamento das goroutines foi diferente e não atingiu o estado cíclico)
- um lembrete de que *teste concorrente que passa não prova ausência de
deadlock*, só aumenta a confiança.

### A correção

Reintroduzir a **camada de garfos** do artigo. As garrafas continuam exatamente
como o enunciado descreve; por baixo, os garfos implementam a precedência H via
o jantar higiênico (comer = segurar **todos** os garfos), restaurando a
aciclicidade que o texto do enunciado, ao pé da letra, omitia.

### Lição

O artigo separa garfos (auxiliares) de garrafas (recurso real) **exatamente** por
causa disso: é o que permite beber de subconjuntos arbitrários mantendo H
acíclico. Colapsar as duas camadas só é correto no caso particular do jantar
(sempre todas as garrafas).

---

## 5. Complexidade

Sendo **V** o número de filósofos (vértices), **E** o número de garrafas
(arestas), **d** o grau de um filósofo e **D** o grau máximo do grafo. Lembrando
que a soma dos graus é 2E.

### Mensagens (a métrica que importa num algoritmo distribuído)

O custo central aqui é o número de **mensagens trocadas**, não operações de CPU.
Por **sessão de bebida** de um filósofo, há no máximo **4 mensagens por vizinho**
(prova de *Economy* / Lemma 6 do artigo): 1 pedido de garfo, 1 garfo, e a
garrafa, que pode trafegar até 2 vezes antes de alguém beber.

| Escopo | Mensagens |
|--------|-----------|
| Por sessão de um filósofo de grau *d* | **O(d)** (≤ 4d) |
| Para *q* sessões em todo o sistema | **O(q · D)** por filósofo, **O(q · E)** no total |
| Simulação completa (`rounds` sessões por filósofo) | **O(rounds · E)** |

Um filósofo que fica **permanentemente tranquilo** não troca um número infinito
de mensagens - só responde a pedidos dos vizinhos (≤ d). Não há *busy-waiting*: o
filósofo dorme no `select` e só acorda com mensagem ou timer.

### Espaço

| Escopo | Espaço |
|--------|--------|
| Estado por filósofo | **O(d)** - 6 flags por aresta incidente (fork, reqf, dirty, bot, reqb, need) |
| Estado de todos os filósofos | **O(V + E)** |
| Matriz de adjacência lida do arquivo | **O(V²)** |
| Mensagens em trânsito entre um par, a qualquer instante | **O(1)** - no máximo 3 (*Boundedness*) |

Como o número de mensagens em trânsito por aresta é constante, bastariam buffers
de tamanho O(d) por filósofo; o código usa O(V) por folga e simplicidade.

### Tempo / espera (fairness)

O critério de starvation é **temporal**, não de passos de CPU. O artigo prova
(Teorema 5) que **todo filósofo com sede bebe em tempo finito**. O atraso até
beber é limitado pela **profundidade do filósofo no grafo de precedência H** - o
comprimento da cadeia de quem tem precedência sobre ele:

- pior caso teórico: profundidade ≤ V − 1, ou seja, espera proporcional a **O(V)**
  períodos de bebida numa cadeia longa;
- na prática observada (ver Seção 6), a espera acompanha o **grau local**, não V:
  filósofos de grau maior têm mais garfos/precedência e esperam menos.

### Custo computacional local (desta implementação)

Cada evento (mensagem recebida ou timer) dispara `react()`, que aplica as regras
até o ponto fixo. Cada varredura `step()` percorre as arestas incidentes em
**O(d)**, e o ponto fixo converge em O(d) varreduras, dando **O(d²) por evento** —
desprezível frente ao custo de coordenação por mensagens. São **V goroutines** e
**V channels** no total.

### Nos grafos do trabalho

| Caso | V | E | D (grau máx.) | Mensagens/sessão por filósofo (≤ 4D) |
|------|---|---|---------------|--------------------------------------|
| 1 | 5 | 5 | 2 | ≤ 8 |
| 2 | 6 | 8 | 4 | ≤ 16 |
| 3 | 12 | 21 | 6 | ≤ 24 |

São cargas pequenas: o gargalo da simulação é o **tempo de relógio** dos estados
(pensar 0..n s, beber 1 s), não o processamento nem o volume de mensagens.

---

## 6. Verificação

| Verificação | Resultado |
|-------------|-----------|
| `go build ./...` / `go vet ./...` | OK |
| `go test -race -count=20 ./solucoes/chandy_misra/` | PASS (3 grafos × 20) |
| Caso 1 real (6 bebidas) | conclui, sem deadlock; espera grau-2 = 4.00 s (equilibrada) |
| Caso 2 real (6 bebidas) | conclui, sem deadlock |
| Caso 3 real (3 bebidas) — *o que travava* | conclui, sem deadlock |

O `-race` cobre dois objetivos de uma vez: prova ausência de *data race* e, se
houvesse deadlock, o teste estouraria o timeout. É a evidência direta do
critério de correção do enunciado ("não apresentar deadlock durante a
execução").

As esperas médias saem equilibradas entre filósofos de mesmo grau, atendendo ao
critério de ausência de starvation.

---

## 7. Arquivos

| Arquivo | Conteúdo |
|---------|----------|
| `solucoes/chandy_misra/solver.go` | implementa `core.Solver`; cria os filósofos, faz a distribuição inicial acíclica de garfos e garrafas, orquestra as goroutines e coleta as métricas |
| `solucoes/chandy_misra/philosopher.go` | a goroutine de cada filósofo: estados, mailbox e as regras das duas camadas (garfo e garrafa) |
| `solucoes/chandy_misra/solver_test.go` | teste dos 3 grafos com tempo acelerado (timeout = possível deadlock) |

Comunicação é **só por mensagens** (channels), sem estado compartilhado mutável
entre goroutines: cada filósofo escreve apenas no próprio registro de métricas
(validado com `-race`).
