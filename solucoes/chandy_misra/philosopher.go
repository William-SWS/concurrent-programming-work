package chandy_misra

import (
	"math/rand"
	"time"

	"github.com/William-SWS/concurrent-programming-work/core"
)

type msgType int

const (
	reqFork    msgType = iota // pede o garfo (envia o request token do garfo)
	sendFork                  // entrega o garfo (chega limpo)
	reqBottle                 // pede a garrafa (envia o request token da garrafa)
	sendBottle                // entrega a garrafa
)

type message struct {
	typ  msgType
	from int // id do remetente (identifica a aresta)
}

// Estado do jantar (camada auxiliar dos garfos / precedência H).
type diningState int

const (
	thinking diningState = iota
	hungry
	eating
)

// Estado da bebida (camada do recurso real)
type drinkingState int

const (
	tranquilo drinkingState = iota
	comSede
	bebendo
)

// edge é o estado local de uma aresta visto por um filósofo. Pra cada aresta,
// exatamente um lado detém o garfo (o outro o request token do garfo) e,
// independentemente, um lado detém a garrafa (o outro o request token dela).
type edge struct {
	neighbor int

	// camada garfo (precedência)
	fork  bool // detém o garfo
	reqf  bool // detém o request token do garfo
	dirty bool // garfo sujo

	// camada garrafa (recurso real)
	bot  bool // detém a garrafa
	reqb bool // detém o request token da garrafa
	need bool // precisa da garrafa nessa sessão
}

type phil struct {
	id     int
	degree int
	sim    *simulacao
	core   *core.Philosopher
	rng    *rand.Rand

	edges map[int]*edge // chaveado pelo id do vizinho
	in    chan message

	dining   diningState
	drinking drinkingState
	drinks   int
	retired  bool

	phaseStart time.Time // início do estado de bebida atual (pra cronometrar)
}

func newPhil(id int, g *core.Graph, sim *simulacao) *phil {
	p := &phil{
		id:       id,
		degree:   g.Degree(id),
		sim:      sim,
		core:     core.NewPhilosopher(id, g),
		rng:      rand.New(rand.NewSource(int64(id)*7919 + 1)),
		edges:    map[int]*edge{},
		in:       sim.inbox[id],
		dining:   thinking,
		drinking: tranquilo,
	}
	for _, j := range g.Neighbors(id) {
		p.edges[j] = &edge{neighbor: j}
	}
	return p
}

func (p *phil) send(to int, m message) { p.sim.inbox[to] <- m }

// run é o laço de eventos do filósofo (uma goroutine). O timer marca o fim do
// estado "tranquilo" (vai ficar com sede) ou do estado "bebendo" (volta a
// tranquilo); as mensagens dos vizinhos movem as camadas de garfo e garrafa
func (p *phil) run() {
	defer p.sim.exited.Done()

	p.phaseStart = time.Now()
	timer := p.scheduleThink()

	for {
		// Cumpriu as rodadas: "se aposenta". Continua servindo garfos e garrafas
		// para os vizinhos terminarem, mas não fica mais com sede
		if p.drinks >= p.sim.rounds && !p.retired {
			p.retire(&timer)
		}

		var timerC <-chan time.Time
		if timer != nil {
			timerC = timer.C
		}

		select {
		case m := <-p.in:
			p.handle(m)
		case <-timerC:
			timer = nil
			switch p.drinking {
			case tranquilo:
				p.becomeThirsty()
			case bebendo:
				p.finishDrinking()
				timer = p.scheduleThink()
			}
		case <-p.sim.stop:
			return
		}

		p.react()
		if t := p.startDrinkIfReady(); t != nil {
			timer = t
			p.react() // deixa o jantar reagir (eating -> thinking) já que bebe
		}
	}
}

func (p *phil) retire(timer **time.Timer) {
	p.retired = true
	p.drinking = tranquilo
	p.dining = thinking
	for _, e := range p.edges {
		e.need = false
	}
	if *timer != nil {
		(*timer).Stop()
		*timer = nil
	}
	p.sim.completed.Done()
}

// handle aplica as regras de recebimento.
func (p *phil) handle(m message) {
	e := p.edges[m.from]
	switch m.typ {
	case reqFork:
		e.reqf = true
	case sendFork:
		e.fork = true
		e.dirty = false // chega limpo
	case reqBottle:
		e.reqb = true
	case sendBottle:
		e.bot = true
	}
}

// react aplica todas as regras locais ate atingir um ponto fixo (uma regra pode
// habilitar outra: receber o último garfo permite "comer", o que suja os garfos
// e libera pedidos pendentes, etc.).
func (p *phil) react() {
	for p.step() {
	}
}

// step executa uma varredura das regras e devolve true se algo mudou
func (p *phil) step() bool {
	changed := false

	// transições do jantar (auxiliar)
	if p.dining == thinking && p.drinking == comSede { // D1
		p.dining = hungry
		changed = true
	}
	if p.dining == eating && p.drinking != comSede { // D2
		p.dining = thinking
		changed = true
	}
	if p.dining == hungry && p.holdsAllForks() {
		p.dining = eating
		for _, e := range p.edges {
			e.dirty = true // ao comer, suja todos os garfos
		}
		changed = true
	}

	// regras dos garfos
	for _, e := range p.edges {
		// R1f: faminto, tenho o token e não tenho o garfo -> peço o garfo.
		if p.dining == hungry && e.reqf && !e.fork {
			p.send(e.neighbor, message{reqFork, p.id})
			e.reqf = false
			changed = true
		}
		// R2f: não estou comendo e há pedido de um garfo sujo -> entrego
		if p.dining != eating && e.reqf && e.fork && e.dirty {
			p.send(e.neighbor, message{sendFork, p.id})
			e.fork = false
			e.dirty = false
			changed = true
		}
	}

	//  regras das garrafas
	for _, e := range p.edges {
		// R1b: com sede, preciso, tenho o token e não tenho a garrafa -> peço.
		if p.drinking == comSede && e.need && e.reqb && !e.bot {
			p.send(e.neighbor, message{reqBottle, p.id})
			e.reqb = false
			changed = true
		}
		// R2b: há pedido da garrafa -> entrego, a menos que eu precise dela E
		// (esteja bebendo ou tenha o garfo dessa aresta, ou seja, precedência)
		if e.reqb && e.bot && !(e.need && (p.drinking == bebendo || e.fork)) {
			p.send(e.neighbor, message{sendBottle, p.id})
			e.bot = false
			changed = true
		}
	}

	return changed
}

func (p *phil) holdsAllForks() bool {
	for _, e := range p.edges {
		if !e.fork {
			return false
		}
	}
	return true
}

// becomeThirsty: tranquilo -> com sede. Sorteia de 2 a n garrafas.
func (p *phil) becomeThirsty() {
	p.core.Record(core.Tranquilo, time.Since(p.phaseStart))
	p.drinking = comSede
	p.phaseStart = time.Now()

	vizinhos := make([]int, 0, len(p.edges))
	for j, e := range p.edges {
		e.need = false
		vizinhos = append(vizinhos, j)
	}

	k := 0
	if p.degree > 0 {
		lo := 2
		if p.degree < lo {
			lo = p.degree
		}
		k = lo + p.rng.Intn(p.degree-lo+1)
	}
	p.rng.Shuffle(len(vizinhos), func(a, b int) {
		vizinhos[a], vizinhos[b] = vizinhos[b], vizinhos[a]
	})
	for _, j := range vizinhos[:k] {
		p.edges[j].need = true
	}
}

// startDrinkIfReady: com sede -> bebendo, quando detem todas as garrafas
// necessárias. Devolve o timer de 1 unidade de bebida
func (p *phil) startDrinkIfReady() *time.Timer {
	if p.drinking != comSede {
		return nil
	}
	for _, e := range p.edges {
		if e.need && !e.bot {
			return nil
		}
	}
	p.core.Record(core.ComSede, time.Since(p.phaseStart))
	p.drinking = bebendo
	p.phaseStart = time.Now()
	return time.NewTimer(Unidade)
}

// finishDrinking: bebendo -> tranquilo. Libera o subconjunto necessário
func (p *phil) finishDrinking() {
	p.core.Record(core.Bebendo, time.Since(p.phaseStart))
	p.drinks++
	for _, e := range p.edges {
		e.need = false
	}
	p.drinking = tranquilo
	p.phaseStart = time.Now()
}

// scheduleThink cria o timer do tempo tranquilo: aleatório de 0 a n segundos
func (p *phil) scheduleThink() *time.Timer {
	var d time.Duration
	if p.degree > 0 {
		d = time.Duration(p.rng.Intn(p.degree+1)) * Unidade
	}
	return time.NewTimer(d)
}
