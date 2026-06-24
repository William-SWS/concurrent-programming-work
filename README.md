# Drinking Philosophers Problem - Concurrent Resource Allocation Simulation

## Overview

This project implements a simulation of the **Drinking Philosophers Problem**, a generalization of the classic Dining Philosophers Problem proposed by Chandy and Misra (1984). The problem models concurrent access to shared resources (bottles) in an arbitrary graph structure, where philosophers (processes) need to acquire multiple resources simultaneously to execute their critical section (drinking).

Unlike the circular dependency of the Dining Philosophers Problem, this generalization allows:
- Arbitrary graph topologies
- Variable resource requirements per philosopher
- Dynamic resource selection per drinking session

## Problem Description

Philosophers are represented as vertices in a graph, while bottles (shared resources) are edges. Each philosopher:
- Has **3 states**: `THINKING` → `THIRSTY` → `DRINKING`
- When transitioning from THINKING to THIRSTY, randomly selects **2 to n** bottles to request (where n = degree of the vertex)
- Must acquire **all** requested bottles before drinking
- Drinks for **1 second**, then returns to THINKING

### Timing Parameters
- **Thinking time**: 0 to n seconds (random, where n = number of adjacent edges)
- **Thirsty time**: Time spent waiting to acquire all requested bottles
- **Drinking time**: 1 second (fixed)

## Implemented Solutions

This project implements **four synchronization solutions** for comparison:

### 1. Resource Ordering (Bottle Numbering)
Each edge is assigned a unique number. Philosophers always acquire bottles in ascending order, preventing circular wait conditions and ensuring deadlock freedom.

### 2. Arbiter Solution (Waiter)
A central arbiter (waiter) controls resource access. Philosophers must request permission from the waiter before attempting to acquire any bottles. The waiter ensures at most N-1 philosophers can be active simultaneously.

### 3. Chandy-Misra
Each bottle has one of the following states: full or empty. Initially, all bottles are empty. When the philosopher gets thirsty, he sends requests for all the neighbours asking for the resource. After receiving the message, if drinking, awaits to end. If the philosopher isn't drinking and the bottle is empty, fulfills the bottle and gives it to the requester. When the philosopher ends drinking, mark the bottle as empty. If a philosopher receives more than one request, serves in order of arrival.

### 4. Randomized Backoff
If a philosopher cannot acquire all of its bottles, it releases the ones it
already holds and waits for a random amount of time before trying again,
breaking circular wait probabilistically.

## Features

- Reads adjacency matrices from text files
- Supports arbitrary graph configurations
- Each philosopher drinks a specified number of times:
  - Cases 1 & 2: 6 drinks per philosopher
  - Case 3: 3 drinks per philosopher
- Collects timing statistics for each state
- Deadlock-free guaranteed
- Fairness evaluation through average waiting times

## Test Cases

| Case | Description | Nodes | Max Edges per Node |
|------|-------------|-------|--------------------|
| 1 | Classic Dining Philosophers (cycle graph) | 5 | 2 |
| 2 | Low connectivity | 6 | 4 |
| 3 | High connectivity | 12 | 6 |

## Requirements
- Go compiler installed in your machine.
- (optional) Python 3 for the comparison script.

## How to Run

```sh
# one solution on one graph
make run SOL=ordenacao GRAFO=data/caso2_bar_6.txt RODADAS=6
# or directly:
go run ./cmd/runner -solucao=ordenacao -grafo=data/caso2_bar_6.txt -rodadas=6

# all four solutions on all three graphs -> results/
make all

# summary comparison table
make compare

# data-race check (evidence of no deadlock/race)
make race
```

Solution names: `ordenacao`, `arbitro`, `chandy_misra`, `backoff`.

## Expected Results

A correct implementation should:
1. **Never deadlock** during execution
2. Maintain **balanced average waiting times** for philosophers with the same degree (no starvation)
3. Complete within ~2 minutes for the provided test cases

## Team Members

- [Samuel William SIlva Almeida]
- [Davi de Oliveira]
- [Antonio Mozar Braga]
- [Antonio Bezerra de Morais Neto]


## Repository Structure

```
concurrent-programming-work/
│
├── README.md             # Project overview and documentation
├── go.mod                # Go module definition
├── Makefile              # build / test / race / run / all / compare targets
├── .gitignore
│
├── core/                 # SHARED CODE 
│   ├── graph.go          # adjacency-matrix parsing + neighbours/degree/edges
│   ├── states.go         # State enum: tranquilo / com sede / bebendo
│   ├── philosopher.go    # Philosopher struct + timing accumulation
│   ├── metrics.go        # per-state metrics + final report (stable TSV)
│   ├── solver.go         # Solver interface implemented by every solution
│   └── graph_test.go     # validates the 3 data files (loads + symmetric)
│
├── solucoes/             # ONE PACKAGE PER SOLUTION - one student each
│   ├── ordenacao/        # Solution 1: resource ordering
│   │   ├── solver.go
│   │   └── solver_test.go   # deadlock/race test (go test -race)
│   ├── arbitro/          # Solution 2: arbiter (waiter)
│   ├── chandy_misra/     # Solution 3: Chandy-Misra
│   └── backoff/          # Solution 4: randomized backoff
│
├── cmd/runner/main.go    # single entry point: -solucao -grafo -rodadas
│
├── data/                 # the 3 graphs from the assignment (adjacency matrices)
│   ├── caso1_jantar_5.txt
│   ├── caso2_bar_6.txt
│   └── caso3_bar_12.txt
│
├── results/              # generated reports (one file per solution/case)
│
├── scripts/
│   ├── run_all.sh        # run every solution on every graph -> results/
│   └── compare_results.py# summary table from results/
│
└── docs/
    └── relatorio.md      # comparison report template
```

### Division of work

Each solution is an independent package under `solucoes/` implementing the
`core.Solver` interface, so the four can be developed in parallel without
conflicts. The shared `core/` (graph parsing, philosopher, metrics) is the only
part that must be agreed on first.

## License

This project was developed as an academic assignment for the Concurrent Programming course.
