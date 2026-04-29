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


## License

This project was developed as an academic assignment for the Concurrent Programming course.
