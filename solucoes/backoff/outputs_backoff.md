  go run ./cmd/runner -solucao=backoff -grafo=data/caso1_jantar_5.txt -rodadas=6

# solucao=backoff
# tempo_total=16.52s
id	grau	tranquilo	com_sede	bebendo	bebidas
0	2	3.80	5.70	6.00	6
1	2	6.67	3.84	6.00	6
2	2	7.10	1.91	6.00	6
3	2	7.14	2.93	6.00	6
4	2	4.66	2.10	6.00	6
# espera media (com sede) por grau:
# grau=2 media=3.30s (n=5)
  go run ./cmd/runner -solucao=backoff -grafo=data/caso2_bar_6.txt -rodadas=6

# solucao=backoff
# tempo_total=23.65s
id	grau	tranquilo	com_sede	bebendo	bebidas
0	2	6.87	3.59	6.00	6
1	2	4.57	1.95	6.00	6
2	3	9.04	2.96	6.00	6
3	3	10.69	5.58	6.00	6
4	2	5.62	0.60	6.00	6
5	4	10.64	7.00	6.00	6
# espera media (com sede) por grau:
# grau=2 media=2.05s (n=3)
# grau=3 media=4.27s (n=2)
# grau=4 media=7.00s (n=1)
  go run ./cmd/runner -solucao=backoff -grafo=data/caso3_bar_12.txt -rodadas=3

# solucao=backoff
# tempo_total=14.97s
id	grau	tranquilo	com_sede	bebendo	bebidas
0	3	3.13	2.16	3.00	3
1	3	1.44	2.82	3.00	3
2	2	4.15	2.17	3.00	3
3	4	8.64	2.32	3.00	3
4	4	6.12	0.00	3.00	3
5	3	4.88	0.29	3.00	3
6	5	8.53	3.43	3.00	3
7	4	4.67	1.65	3.00	3
8	3	5.00	0.31	3.00	3
9	6	11.23	0.00	3.00	3
10	2	1.73	0.89	3.00	3
11	3	3.46	1.39	3.00	3
# espera media (com sede) por grau:
# grau=2 media=1.53s (n=2)
# grau=3 media=1.39s (n=5)
# grau=4 media=1.32s (n=3)
# grau=5 media=3.43s (n=1)
# grau=6 media=0.00s (n=1)