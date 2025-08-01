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

## 3. DynamoDB vs MongoDB

### Decisão: **DynamoDB**

#### Por que optei por DynamoDB?
* **Totalmente gerenciado pela AWS**, com escalabilidade automática e billing por demanda.
* **Integração nativa com EventBridge**, permitindo arquitetura event-driven sem dependências externas.
* **Latência extremamente baixa** em queries por chave primária.
* **Suporte direto em SDKs AWS** para controle e rastreamento de objetos processados.

#### Alternativa Considerada: **MongoDB**
- **Prós**:
  * Modelo de documentos flexível e expressivo.
  * Consultas mais avançadas que o DynamoDB em estruturas aninhadas.
- **Contras**:
  * **Não possui integração direta com EventBridge** (exige solução intermediária ou polling).
  * A versão **gerenciada na AWS é limitada ao MongoDB Atlas**, que **não é nativa nem 100% integrada** ao ecossistema AWS.
  * **Maior esforço operacional** para setup de triggers, monitoramento e permissionamento granular.

---

## 4. CLI Catalog Sync vs Glue Crawlers

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

## 5. Go vs Spark (Camada Bronze)

### Decisão: **Go (parquet-go)**

#### Por que optei por Go?
- **Alto desempenho com baixo overhead**, ideal para Lambda e ECS.
- **Cold start mínimo**, útil em ambientes serverless.
- **Processamento streaming-friendly** com controle detalhado.
- **Independência de engine**: dispensa cluster ou framework externo.

#### Alternativa Considerada: **Spark (Glue)**
- **Prós**:
  - Suporte nativo no Glue (infraestrutura gerenciada).
  - Robusto para pipelines complexos e transformações estruturadas.
- **Contras**:
  - Tempo de execução significativamente maior.
  - Overhead desnecessário para arquivos pequenos.
  - Latência de cold start elevada.

### Justificativa baseada em benchmark

Para a etapa de ingestão da **camada bronze**, optei pela implementação em **Go puro** após a realização de um **benchmark reproduzível** comparando Go + ECS com Spark (via AWS Glue). 

O resultado foi expressivo: o Go entregou **uma escrita quase 10x mais rápida**, com **uso de memória extremamente baixo** (menos de 160 MB), enquanto o Glue Job apresentou latência elevada de inicialização e tempo total significativamente maior.

Além disso, o mesmo código em Go pode ser reaproveitado em **funções Lambda** para arquivos menores (até \~100 mil linhas), eliminando o cold start do ECS e tornando a solução ainda mais versátil e econômica.

> Para detalhes completos, veja o benchmark documentado em [Benchmark de Ingestão – Go vs PySpark](benchmark.md).

---

## 6. EMR Serverless vs Glue Jobs (Silver e Gold)

### Decisão: **EMR Serverless**

#### Por que optei por EMR Serverless?
* **Maior flexibilidade para organizar o código em etapas reutilizáveis**, com suporte mais natural a projetos modulares em Spark.
* **Execução mais estável e previsível em cargas com joins e transformações complexas**, principalmente após o tuning adequado de configurações.
* **Menor acoplamento com abstrações específicas da AWS**, permitindo maior portabilidade futura se necessário.

#### Alternativa Inicial: **Glue Jobs**
* **Prós**:

  * Provisionamento simplificado com escalabilidade automática.
  * Forte integração com o ecossistema da AWS (IAM, Glue Catalog, Lake Formation, etc).
* **Contras**:
  * **Menor flexibilidade para organização modular do código**, dificultando o reuso entre jobs.
  * **Depuração mais trabalhosa** em falhas silenciosas ou erros de runtime; os logs podem ser incompletos se não forem forçados via `flush`.
  * **Menor controle sobre o ambiente de execução** — embora seja possível ajustar alguns parâmetros Spark, o nível de customização é inferior ao do EMR Serverless.

---

## 7. Iceberg vs Delta Lake

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

## 8. Kinesis vs Apache Kafka

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

## 9. Step Functions vs Apache Airflow

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

## 10. S3 + Athena vs Redshift

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

## 11. Managed Grafana vs QuickSight

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

[Voltar para a página inicial](../README.md#documentação) | [Próximo: Pré-Requisitos](pre-requirements.md)