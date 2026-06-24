#!/usr/bin/env python3
"""Compara os relatórios gerados em results/ e imprime uma tabela resumo.

Lê os arquivos results/<caso>/<solucao>.txt produzidos por scripts/run_all.sh e
resume, por solução e por caso, o tempo total e a espera média (com sede) —
exatamente os números que o enunciado pede para comparar.
"""
import glob
import os
import re

RESULTS_DIR = os.path.join(os.path.dirname(__file__), "..", "results")


def parse(path):
    """Extrai tempo_total e a espera média geral de um relatório."""
    total = None
    esperas = []
    with open(path) as f:
        for line in f:
            if line.startswith("# tempo_total="):
                total = float(re.search(r"=([\d.]+)s", line).group(1))
            elif re.match(r"^\d+\t", line):  # linha de dados: id\tgrau\t...
                cols = line.split("\t")
                esperas.append(float(cols[3]))  # coluna com_sede
    media = sum(esperas) / len(esperas) if esperas else 0.0
    return total, media


def main():
    linhas = []
    for path in sorted(glob.glob(os.path.join(RESULTS_DIR, "*", "*.txt"))):
        caso = os.path.basename(os.path.dirname(path))
        solucao = os.path.splitext(os.path.basename(path))[0]
        total, media = parse(path)
        linhas.append((caso, solucao, total, media))

    if not linhas:
        print("Nenhum resultado encontrado. Rode scripts/run_all.sh primeiro.")
        return

    print(f"{'caso':8} {'solucao':14} {'total(s)':>9} {'espera_media(s)':>16}")
    print("-" * 50)
    for caso, solucao, total, media in linhas:
        t = f"{total:.2f}" if total is not None else "-"
        print(f"{caso:8} {solucao:14} {t:>9} {media:>16.2f}")


if __name__ == "__main__":
    main()
