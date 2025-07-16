# Trade-offs e Decisões de Arquitetura

Nesta seção, explico as principais decisões técnicas e os trade-offs avaliados ao longo do projeto. As escolhas foram guiadas pelos princípios de **baixo custo**, **simplicidade operacional**, **escalabilidade** e **governança via código**.

---

## 1. CloudFormation vs Terraform
### Decisão: **CloudFormation**
#### Por que optei por CloudFormation?
- **Integração nativa com AWS**: facilita o provisionamento e acesso a recursos.
- **Zero custo adicional**: não requer serviços externos ou execução em pipelines.
- **Simplicidade operacional**: atende plenamente o escopo do projeto.

#### Alternativa Considerada: **Terraform**
- **Prós**:
  - Suporte multi-cloud.
  - Modularidade avançada.
- **Contras**:
  - Requer tooling e setup adicional.
  - Menor integração nativa com recursos AWS.

---

## 2. Aurora Serverless v2 vs RDS provisionado
### Decisão: **Aurora Serverless v2**
#### Por que optei por Aurora Serverless?
- Escala sob demanda, ideal para workloads intermitentes.
- Custo por segundo, otimizado para uso esporádico.
- Elimina a necessidade de configurar instâncias.

#### Alternativa Considerada: **RDS provisionado**
- **Prós**:
  - Estabilidade em cargas constantes.
  - Mais opções de engines e regiões.
- **Contras**:
  - Custo fixo mesmo em ociosidade.
  - Maior complexidade de dimensionamento.

---

## 3. CLI Catalog Sync vs Glue Crawlers
### Decisão: **CLI Catalog Sync**
#### Por que optei por não usar Crawlers?
- Total controle via código (sem inferência).
- Alinhado à filosofia de versionamento determinístico.
- Mapeamento automático de schema → database via CLI.

#### Alternativa Considerada: **Glue Crawlers**
- **Prós**:
  - Úteis quando o schema muda frequentemente.
  - Descoberta automática de dados no S3.
- **Contras**:
  - Pouco controle e versionamento.
  - Execução dependente de agendamento e permissões extras.

---

## 4. Go vs Spark (Camada Bronze)
### Decisão: **Go (parquet-go)**
#### Por que optei por Go?
- **Alto desempenho com baixo overhead**, ideal para Lambda e ECS.
- **Cold start mínimo**, útil em ambientes serverless.
- **Processamento streaming-friendly** com controle detalhado.
- **Independência de engine**: dispensa cluster ou framework externo.

#### Alternativa Considerada: **Spark (Glue)**
- **Prós**:
  - Suporte nativo no Glue.
  - Robusto para cargas em lote.
- **Contras**:
  - Overhead desnecessário para arquivos pequenos.
  - Latência de cold start elevada.

---

## 5. EMR Serverless vs Glue Jobs (Silver e Gold)
### Decisão: **EMR Serverless**
#### Por que optei por EMR Serverless?
- **Controle granular sobre recursos Spark** (memória, paralelismo).
- **Boa performance em joins e transformações complexas**.

#### Alternativa Inicial: **Glue Jobs**
- **Prós**:
  - Provisionamento simplificado.
  - Integração direta com AWS nativa.
- **Contras**:
  - Difícil de depurar: logs incompletos ou ausentes em falhas silenciosas.
  - Limitado para modularização e reaproveitamento de código entre etapas.

---

## 6. Iceberg vs Delta Lake
### Decisão: **Apache Iceberg**
#### Por que optei por Iceberg?
- Integração nativa com Glue Catalog e Athena.
- Metadados transacionais, com versionamento leve.
- Independente de engine (Spark, Trino, Athena, etc.).

#### Alternativa Considerada: **Delta Lake**
- **Prós**:
  - Ecossistema maduro em Spark/Databricks.
- **Contras**:
  - Fortemente acoplado ao Spark.
  - Integração com AWS exige workarounds.

---

## 7. Kinesis vs Apache Kafka
### Decisão: **Kinesis**
#### Por que optei por Kinesis?
- Totalmente gerenciado (serverless).
- Custo baseado em throughput.
- Integração nativa com Firehose e Pipes.

#### Alternativa Considerada: **Apache Kafka**
- **Prós**:
  - Open-source com alto controle.
  - Escalável para altíssimos volumes.
- **Contras**:
  - Complexidade operacional elevada.
  - Custo fixo mais alto (mesmo em versões gerenciadas).

---

## 8. Step Functions vs Apache Airflow
### Decisão: **Step Functions**
#### Por que optei por Step Functions?
- Orquestração serverless e com billing granular.
- Integração nativa com serviços AWS.
- Facilidade para versionar e visualizar workflows.

#### Alternativa Considerada: **Apache Airflow**
- **Prós**:
  - Extremamente flexível e poderoso.
- **Contras**:
  - Requer ambiente dedicado ou MWAA.
  - Integração com AWS exige operadores adicionais.

---

## 9. S3 + Athena vs Redshift
### Decisão: **S3 + Athena**
#### Por que optei por S3 + Athena?
- Custo zero de provisionamento.
- Escalabilidade praticamente ilimitada.
- Ideal para acesso esporádico com volume variável.

#### Alternativa Considerada: **Redshift**
- **Prós**:
  - Performance superior em BI intensivo.
- **Contras**:
  - Clusters requerem tuning e manutenção.
  - Custo fixo mesmo com pouco uso.

---

## 10. Managed Grafana vs QuickSight
### Decisão: **Managed Grafana**
#### Por que optei por Grafana?
- Suporte a Athena, CloudWatch, logs e métricas.
- Flexibilidade na construção de dashboards técnicos.
- Provisionável via CloudFormation.

#### Alternativa Considerada: **QuickSight**
- **Prós**:
  - Ótimo para dashboards executivos e BI simples.
- **Contras**:
  - Integração limitada com fontes técnicas.
  - Incompatível com IaC completa via CloudFormation.

---

## Conclusão

Todas as decisões foram tomadas com base no escopo e objetivos do projeto, priorizando:

- **Custo variável e previsível**
- **Operação simplificada e 100% reprodutível**
- **Aproveitamento de recursos nativos AWS**
- **Separação clara de responsabilidades por camada**

As escolhas entre Go e Spark, Glue e EMR, ou Iceberg e Delta foram feitas com foco em **ajuste fino por etapa**, e não com uma abordagem one-size-fits-all.

---

[Voltar para a página inicial](../README.md#documentação)