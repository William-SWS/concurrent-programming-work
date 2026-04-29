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

This project implements **two synchronization solutions** for comparison:

### 1. Resource Ordering (Bottle Numbering)
Each edge is assigned a unique number. Philosophers always acquire bottles in ascending order, preventing circular wait conditions and ensuring deadlock freedom.

### 2. Arbiter Solution (Waiter)
A central arbiter (waiter) controls resource access. Philosophers must request permission from the waiter before attempting to acquire any bottles. The waiter ensures at most N-1 philosophers can be active simultaneously.

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
├── README.md               # Project overview and documentation
├── go.mod                  # Go module definition and dependencies
├── Makefile                # Build, run and test automation targets
├── .gitignore              # Files and directories ignored by Git
│
├── cmd/                    # Application entry points
│   ├── main.go             # Main executable: runs the simulation for all test cases
│   └── benchmark/
│       └── main.go         # Benchmark executable: measures performance across solutions
│
├── pkg/                    # Public, reusable packages
│   ├── graph/              # Graph data structure (adjacency matrix representation)
│   │   ├── graph.go        # Graph type definition and core operations
│   │   ├── adjacency.go    # Adjacency matrix parsing and helpers
│   │   └── graph_test.go   # Unit tests for graph package
│   │
│   ├── philosopher/        # Philosopher abstraction shared across solutions
│   │   ├── philosopher.go  # Philosopher struct, lifecycle and timing logic
│   │   ├── states.go       # Philosopher state definitions (THINKING/THIRSTY/DRINKING)
│   │   └── philosopher_test.go # Unit tests for philosopher package
│   │
│   └── bottle/             # Bottle (shared resource / edge) abstraction
│       ├── bottle.go       # Bottle struct and mutex-based locking logic
│       └── bottle_test.go  # Unit tests for bottle package
│
├── internal/               # Private solution implementations
│   ├── solution1_ordering/ # Solution 1: Resource ordering (bottle numbering)
│   │   ├── solver.go       # Orchestrates the simulation for this solution
│   │   ├── philosopher.go  # Philosopher behaviour specific to ordering strategy
│   │   └── solver_test.go  # Integration tests for solution 1
│   │
│   ├── solution2_arbiter/  # Solution 2: Central arbiter (waiter)
│   │   ├── solver.go       # Orchestrates the simulation for this solution
│   │   ├── arbiter.go      # Arbiter goroutine that grants/denies resource access
│   │   ├── philosopher.go  # Philosopher behaviour specific to arbiter strategy
│   │   └── solver_test.go  # Integration tests for solution 2
│   │
│   ├── solution3_chandy_misra/ # Solution 3: Chandy-Misra token-passing algorithm
│   │   ├── solver.go       # Orchestrates the simulation for this solution
│   │   ├── token.go        # Token/fork state and passing logic
│   │   ├── philosopher.go  # Philosopher behaviour specific to Chandy-Misra strategy
│   │   └── solver_test.go  # Integration tests for solution 3
│   │
│   └── solution4_backoff/  # Solution 4: Randomised exponential backoff
│       ├── solver.go       # Orchestrates the simulation for this solution
│       ├── philosopher.go  # Philosopher behaviour specific to backoff strategy
│       └── solver_test.go  # Integration tests for solution 4
│
├── data/                   # Input graph definitions (adjacency matrices)
│   ├── caso1_jantar_5.txt  # Case 1: Classic dining philosophers (5-node cycle)
│   ├── caso2_bar_6.txt     # Case 2: Bar scenario with low connectivity (6 nodes)
│   └── caso3_bar_12.txt    # Case 3: Bar scenario with high connectivity (12 nodes)
│
├── results/                # Output directory for simulation results
│   ├── caso1/              # Timing statistics for case 1
│   ├── caso2/              # Timing statistics for case 2
│   └── caso3/              # Timing statistics for case 3
│
├── scripts/                # Utility scripts for running and analysing experiments
│   ├── run_all.sh          # Shell script to execute all test cases for every solution
│   └── compare_results.py  # Python script to compare and plot results across solutions
│
└── docs/                   # Additional project documentation
    ├── especificacao.md    # Problem specification and requirements
    ├── solucoes.md         # Description and analysis of each implemented solution
    └── resultados.md       # Experimental results and performance comparison
```

## License

This project was developed as an academic assignment for the Concurrent Programming course.
