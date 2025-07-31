# Benchmark de Ingestão – Go vs PySpark

## 1. Objetivo

Este benchmark foi desenvolvido como parte de uma iniciativa experimental para explorar alternativas de ingestão de dados no data lake, buscando:

* Reduzir custos operacionais
* Melhorar o uso de recursos computacionais
* Aumentar a performance em cenários de ingestão em larga escala

A proposta foi testar uma abordagem fora do comum para ambientes analíticos (Go) contra o padrão tradicional (PySpark), com foco em resultados concretos de desempenho.

---

## 2. Metodologia

Utilizou-se um arquivo CSV de reviews contendo mais de **1,5 milhão de registros**, o mesmo dataset usado no case principal do projeto.

Foram medidos:

* Tempo total de orquestração (execução end-to-end)
* Tempo de leitura do CSV
* Tempo de escrita em Parquet
* Uso de memória (quando disponível)

---

## 3. Implementações Testadas

| Abordagem               | Detalhes Técnicos                                                                  |
| ----------------------- | ---------------------------------------------------------------------------------- |
| **Go em ECS (Fargate)** | `parquet-go`, concorrência, escrita direta em Parquet                              |
| **PySpark no Glue Job** | Spark 3.x, escrita via `.write.parquet()`                                          |

---

## 4. Provisionamento e Ambiente

### Go (ECS Fargate)

* **2 vCPUs**, **4 GB RAM**
* Escrita concorrente com streaming para Parquet
* Baixíssimo uso de memória e tempo de execução

### PySpark (Glue Job)

* **Worker G.1X** _(menor instância do Glue)_
* 4 vCPUs, 16 GB RAM, 94 GB disco
* Ambiente gerenciado, alta latência inicial

---

## 5. Resultados

### Go (ECS)

```text
=== Benchmark Result ===
Implementation      : go
Orchestration Time  : 53.7s
Task Duration       : 4.54s
CSV Read Time       : 121.7ms
Parquet Write Time  : 3.79s
Memory Used (MB)    : 157.83
```

### PySpark (Glue Job)

```text
=== Benchmark Result ===
Implementation      : python
Orchestration Time  : 1m29.7s
Task Duration       : 37.57s
CSV Read Time       : 8.07s
Parquet Write Time  : 25.79s
Memory Used (MB)    : [não disponível – ver observações]
```

### Observações sobre uso de memória no benchmark Python

Durante a execução do benchmark com **AWS Glue Job (Spark/PySpark)**, **não foi possível capturar métricas confiáveis de memória**, devido a limitações do ambiente:

* APIs como `SparkStatusStore` ou `ExecutorMemoryStatus` **não estão disponíveis em PySpark no Glue**
* `SparkContext.getExecutorMemoryStatus()` retorna erro de atributo
* Bibliotecas como `psutil` **não podem ser instaladas no ambiente gerenciado**

> Por isso, a métrica de memória foi omitida na versão PySpark para evitar comparações injustas.

---

## 6. Reprodução

O benchmark pode ser reproduzido facilmente através da CLI, que provisiona toda a infraestrutura necessária no ambiente AWS de forma automatizada. Para mais detalhes sobre os recursos criados, consulte a documentação sobre as [stacks](stacks.md).

```bash
>>> deploy --stack benchmark
```

Para executar o benchmark com Go (ECS):

```bash
>>> benchmark --run ecs
```

Para executar com PySpark (Glue Job):

```bash
>>> benchmark --run glue
```

Os resultados são armazenados no bucket:

```
s3://dm-benchmark-<account_id>/
```

Incluem:

* Arquivo Parquet gerado
* JSON com todas as métricas coletadas por execução

---

## 7. Conclusões

A comparação entre as abordagens trouxe resultados expressivos:

| Comparativo        | Go (ECS)      | PySpark (Glue Job)     |
| ------------------ | ------------- | ---------------------- |
| Orquestração total | **\~53s**     | \~**1m29s**            |
| Duração da task    | **\~4.5s**    | \~**37.5s**            |
| Leitura CSV        | **121ms**     | \~**8s**               |
| Escrita Parquet    | **\~3.8s**    | \~**25.8s**            |
| Memória usada      | **157 MB**    | Indisponível           |
| Infraestrutura     | 2 vCPU / 4 GB | 4 vCPU / 16 GB (1 DPU) |

* O **Go com ECS** apresentou desempenho **quase 10x superior** na etapa de escrita e **uso de memória drasticamente menor**.
* O mesmo código em Go pode ser reutilizado em funções **AWS Lambda** para cenários com menor volume de dados (até \~100 mil linhas), eliminando o tempo de provisionamento do ECS e tornando a solução ainda mais atrativa para cargas pequenas ou frequentes.
* Usando **Go**, há **controle total sobre os recursos de infraestrutura e execução**, permitindo decisões precisas sobre paralelismo, alocação de memória e estratégias de escrita. Embora exija desenvolvimento próprio, **o código pode ser continuamente otimizado e evoluído**, sem limitações impostas por engines como Spark ou ambientes gerenciados.
* Já o **Glue Job** sofreu com latência maior de inicialização e maior custo computacional, apesar de oferecer um ambiente gerenciado e simplificado.

> **Conclusão:** Para cenários de ingestão massiva na **camada bronze**, com estrutura bem definida e sem dependência de transformações complexas, a abordagem com Go puro — escrevendo diretamente em Parquet — é altamente recomendada por sua **eficiência, simplicidade e custo-benefício**.
